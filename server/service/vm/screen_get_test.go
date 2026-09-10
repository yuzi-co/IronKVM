package vm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/common"
	"NanoKVM-Server/proto"
)

// The defect this closes. The menu kept its own copy of every capture setting
// and pushed it to the server on load, because there was no way to ask what the
// server held. A setting changed in one browser was invisible to the next one,
// and a setting the server had restored from the card was overwritten by
// whatever the browser remembered. There has to be a read.

// restoreScreen puts the capture settings back after a test. The screen
// singleton is process wide, so a test that changes it and does not put it back
// is read by the next one.
func restoreScreen(t *testing.T) {
	t.Helper()

	before := common.GetScreen().Snapshot()

	t.Cleanup(func() {
		common.SetScreen("resolution", int(before.Height))
		// One key, two fields: the value's size decides which.
		common.SetScreen("quality", int(before.Quality))
		common.SetScreen("quality", int(before.BitRate))
		common.SetScreen("fps", before.FPS)
		common.SetScreen("gop", int(before.GOP))
		common.SetScreen("codec", int(before.Codec))
	})
}

func getScreen(t *testing.T) proto.GetScreenRsp {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/vm/screen", nil)

	NewService().GetScreen(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	var envelope struct {
		Code int                `json:"code"`
		Data proto.GetScreenRsp `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode %s: %s", recorder.Body.String(), err)
	}
	if envelope.Code != 0 {
		t.Fatalf("code = %d, want 0: %s", envelope.Code, recorder.Body.String())
	}

	return envelope.Data
}

func TestGetScreenReportsWhatTheServerHolds(t *testing.T) {
	restoreScreen(t)

	common.SetScreen("resolution", 720)
	common.SetScreen("fps", 50)
	common.SetScreen("gop", 12)
	common.SetScreen("codec", common.CodecH265)

	got := getScreen(t)

	if got.Width != 1280 || got.Height != 720 {
		t.Errorf("resolution = %dx%d, want 1280x720", got.Width, got.Height)
	}
	if got.FPS != 50 {
		t.Errorf("fps = %d, want 50", got.FPS)
	}
	if got.GOP != 12 {
		t.Errorf("gop = %d, want 12", got.GOP)
	}
	if got.Codec != common.CodecH265 {
		t.Errorf("codec = %d, want H.265 (%d)", got.Codec, common.CodecH265)
	}
}

// One API key carries either a JPEG quality or an H.264 bitrate depending on
// its size, and the browser cannot tell which it is being given. Report the two
// separately so each delivery path reads its own.
func TestGetScreenReportsTheQualityAndTheBitrateSeparately(t *testing.T) {
	restoreScreen(t)

	common.SetScreen("quality", 60)
	common.SetScreen("quality", 2000)

	got := getScreen(t)

	if got.Quality != 60 {
		t.Errorf("quality = %d, want 60", got.Quality)
	}
	if got.BitRate != 2000 {
		t.Errorf("bitRate = %d, want 2000", got.BitRate)
	}
}

// The menu draws a tick beside the value it is given, so the read has to answer
// with a value the menu offers. CheckScreen is the rule the capture loop
// applies before it starts, and a quality the board does not offer is repaired
// to one it does.
func TestGetScreenAnswersWithTheValuesAStreamWouldUse(t *testing.T) {
	restoreScreen(t)

	common.SetScreen("quality", 77)

	if got := getScreen(t); got.Quality != 80 {
		t.Fatalf("quality = %d, want the repaired 80", got.Quality)
	}
}
