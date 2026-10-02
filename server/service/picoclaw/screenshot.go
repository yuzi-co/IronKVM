package picoclaw

import (
	"bytes"
	"context"
	"encoding/base64"
	"image/jpeg"
	"net/http"
	"time"

	"NanoKVM-Server/common"
	"NanoKVM-Server/config"

	"github.com/gin-gonic/gin"
)

var screenshotRetryDelay = 100 * time.Millisecond

const screenshotRetryCount = 30

// picoclawScreenshotSettings returns the size and quality of the screenshots
// the model gets when it does not ask for a size, from server.yaml. A
// variable so the tests can replace it.
var picoclawScreenshotSettings = func() config.Picoclaw {
	return config.GetInstance().Picoclaw.WithDefaults()
}

func (s *Service) Screenshot(c *gin.Context) {
	var query ScreenshotQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		writePicoclawError(c, newPicoclawError(CodeInvalidAction, "invalid screenshot query"))
		return
	}

	sessionID, sessionErr := s.requireSessionID(c)
	if sessionErr != nil {
		writePicoclawError(c, sessionErr)
		return
	}

	releaseAfter, lockErr := s.lock.AcquireTemporary(sessionID)
	if lockErr != nil {
		writePicoclawError(c, lockErr)
		return
	}
	if releaseAfter {
		defer s.lock.Release(sessionID)
	}

	data, meta, err := s.captureScreenshot(c.Request.Context(), query)
	if err != nil {
		writePicoclawError(c, err)
		return
	}

	if query.Format == "base64" {
		meta.ImageBase64 = base64.StdEncoding.EncodeToString(data)
		writeSuccess(c, meta)
		return
	}

	c.Data(http.StatusOK, "image/jpeg", data)
}

func (s *Service) captureScreenshot(ctx context.Context, query ScreenshotQuery) ([]byte, ScreenshotMeta, *PicoclawError) {
	common.CheckScreen()
	values := common.GetScreen().Snapshot()
	width, height, quality := resolveScreenshotRequest(query, values)

	releaseLease, claimFresh, leaseErr := s.acquireCaptureLease(ctx)
	if leaseErr != nil {
		return nil, ScreenshotMeta{}, newPicoclawError(CodeScreenshotFailed, "screenshot capture canceled")
	}
	defer releaseLease()

	sourceWidth, sourceHeight := values.Width, values.Height
	if sourceWidth == 0 || sourceHeight == 0 {
		// The stream is set to the automatic resolution, so the screen size
		// is not known up front and the sizes above are 0x0, which libkvm
		// reads as "the full HDMI frame". Read one, take the screen size
		// from it, and read again at the size the request asks for.
		data, err := s.readScreenshotFrame(ctx, 0, 0, quality, &claimFresh)
		if err != nil {
			return nil, ScreenshotMeta{}, err
		}
		frameWidth, frameHeight, ok := jpegDimensions(data)
		if !ok {
			return data, ScreenshotMeta{Format: "jpeg"}, nil
		}
		sourceWidth, sourceHeight = frameWidth, frameHeight
		width, height, _ = resolveScreenshotRequest(query, common.ScreenValues{
			Width:   frameWidth,
			Height:  frameHeight,
			Quality: values.Quality,
		})
		if width == frameWidth && height == frameHeight {
			return data, screenshotMeta(sourceWidth, sourceHeight, frameWidth, frameHeight), nil
		}
	}

	data, err := s.readScreenshotFrame(ctx, width, height, quality, &claimFresh)
	if err != nil {
		return nil, ScreenshotMeta{}, err
	}
	// The frame says what it really is. The request is the fallback for a
	// frame whose header cannot be read.
	if frameWidth, frameHeight, ok := jpegDimensions(data); ok {
		width, height = frameWidth, frameHeight
	}
	return data, screenshotMeta(sourceWidth, sourceHeight, width, height), nil
}

func screenshotMeta(sourceWidth, sourceHeight, captureWidth, captureHeight uint16) ScreenshotMeta {
	return ScreenshotMeta{
		SourceWidth:   sourceWidth,
		SourceHeight:  sourceHeight,
		CaptureWidth:  captureWidth,
		CaptureHeight: captureHeight,
		Format:        "jpeg",
	}
}

// jpegDimensions reads the size of a JPEG from its header.
func jpegDimensions(data []byte) (uint16, uint16, bool) {
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width > 0xffff || cfg.Height > 0xffff {
		return 0, 0, false
	}
	return uint16(cfg.Width), uint16(cfg.Height), true
}

// readScreenshotFrame reads one JPEG frame, retrying while libkvm has none
// ready. claimFresh is the capture lease's "this frame predates the lease"
// check; it is cleared once used so that a second read in the same capture
// does not discard a frame again.
func (s *Service) readScreenshotFrame(ctx context.Context, width, height, quality uint16, claimFresh *func() bool) ([]byte, *PicoclawError) {
	for attempt := 0; attempt < screenshotRetryCount; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, newPicoclawError(CodeScreenshotFailed, "screenshot capture canceled")
		}
		data, result := s.vision.ReadMjpeg(width, height, quality)
		switch {
		case result == 5 || result == -3 || result == -4 || result == -5:
			if attempt < screenshotRetryCount-1 {
				timer := time.NewTimer(screenshotRetryDelay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return nil, newPicoclawError(CodeScreenshotFailed, "screenshot capture canceled")
				case <-timer.C:
				}
				continue
			}
			return nil, newPicoclawError(CodeScreenshotNoSignal, "no HDMI signal or frame unavailable")
		case result < 0 || len(data) == 0:
			return nil, newPicoclawError(CodeScreenshotFailed, "failed to capture screenshot")
		default:
			if *claimFresh != nil && (*claimFresh)() {
				*claimFresh = nil
				continue
			}
			*claimFresh = nil
			return data, nil
		}
	}

	return nil, newPicoclawError(CodeScreenshotFailed, "failed to capture screenshot")
}

func resolveScreenshotRequest(query ScreenshotQuery, values common.ScreenValues) (uint16, uint16, uint16) {
	width := values.Width
	height := values.Height
	quality := values.Quality

	if query.Format == "base64" {
		settings := picoclawScreenshotSettings()
		// At most 16:9, so the 960 wide default gives 960x540 for a 16:9
		// source and 720x540 for a 4:3 one.
		maxWidth := uint16(settings.ScreenshotWidth)
		maxHeight := uint16(settings.ScreenshotWidth * 9 / 16)
		width, height = fitWithinBounds(width, height, maxWidth, maxHeight)
		maxQuality := uint16(settings.ScreenshotQuality)
		if quality == 0 || quality > maxQuality {
			quality = maxQuality
		}
	}

	width, height = applyRequestedDimensions(width, height, query.Width, query.Height)
	if query.Quality > 0 {
		quality = query.Quality
	}

	return width, height, quality
}

func applyRequestedDimensions(defaultWidth uint16, defaultHeight uint16, requestedWidth uint16, requestedHeight uint16) (uint16, uint16) {
	switch {
	case requestedWidth > 0 && requestedHeight > 0:
		return requestedWidth, requestedHeight
	case requestedWidth > 0:
		return fitWithinBounds(defaultWidth, defaultHeight, requestedWidth, 0)
	case requestedHeight > 0:
		return fitWithinBounds(defaultWidth, defaultHeight, 0, requestedHeight)
	default:
		return defaultWidth, defaultHeight
	}
}

func fitWithinBounds(sourceWidth uint16, sourceHeight uint16, maxWidth uint16, maxHeight uint16) (uint16, uint16) {
	if sourceWidth == 0 || sourceHeight == 0 {
		return sourceWidth, sourceHeight
	}
	if maxWidth == 0 && maxHeight == 0 {
		return sourceWidth, sourceHeight
	}

	width := int(sourceWidth)
	height := int(sourceHeight)
	limitedWidth := width
	limitedHeight := height

	if maxWidth > 0 && limitedWidth > int(maxWidth) {
		limitedWidth = int(maxWidth)
		limitedHeight = height * limitedWidth / width
	}

	if maxHeight > 0 && limitedHeight > int(maxHeight) {
		limitedHeight = int(maxHeight)
		limitedWidth = width * limitedHeight / height
	}

	if limitedWidth <= 0 {
		limitedWidth = 1
	}
	if limitedHeight <= 0 {
		limitedHeight = 1
	}

	return uint16(limitedWidth), uint16(limitedHeight)
}
