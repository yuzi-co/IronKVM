// The file browser: the files inside an image, listed, downloaded, uploaded
// and deleted from the web UI, so files can be exchanged with the host
// through an image it mounts.
//
// Ported from sipeed/NanoKVM (server/service/storage/files.go and the helpers
// of media.go, commit 5572e1de, PR #957), onto IronKVM's drive API: an image
// is named by its path under /data, as the drives name it, and whether the
// host can see it is read from the drives rather than a medium state.
// Copyright (c) Sipeed. Licensed under the GNU General Public License v3.0.

package storage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/storage/diskfs"
	"NanoKVM-Server/utils"
)

// Error codes the web UI tells apart. The response envelope is shared with
// every endpoint, so they stay in a range of their own.
const (
	codeInvalidArgument = -1
	codeOperationFailed = -2
	codeMediumBusy      = -12
	codeNotBrowsable    = -13
	codeNoSpace         = -14
	codeReadOnly        = -15
)

const (
	// uploadMaxBytes is a hard ceiling for a single upload, so a runaway
	// request cannot fill the image before the free space check has had its
	// say.
	uploadMaxBytes = 8 << 30
	// multipartSlack is what the multipart framing adds to a file at most:
	// the Content-Length of an upload covers both.
	multipartSlack = 64 << 10
	// maxListEntries bounds a listing. The board has little memory, and a
	// directory larger than this is not one anybody browses by eye.
	maxListEntries = 5000
)

var (
	// errImageServed refuses opening an image the host can see in a way
	// that is not safe while it can.
	errImageServed = errors.New("the host has the image")
	errNoSpace     = errors.New("not enough space in the image")
)

// imageOpenError is errImageOpen. It counts as errImageLoaded, so callers
// that refuse a loaded image refuse one the file browser has open too.
type imageOpenError struct{}

func (imageOpenError) Error() string        { return "the image is open in the file browser" }
func (imageOpenError) Is(target error) bool { return target == errImageLoaded }

var errImageOpen error = imageOpenError{}

// mountImage opens an image's filesystem. Tests replace it with a directory.
var mountImage = diskfs.Mount

// openHere reports whether the file browser has an image open, and whether
// for writing. Tests replace it.
var openHere = diskfs.State

// browserAllows checks an insert against what the file browser has open: an
// image open for writing may not reach the host at all, and one open for
// reading may reach it only read-only. A device such as the Ventoy disk is
// checked by the images it reads. The caller holds driveMu.
func browserAllows(file string, hostWrites bool) error {
	images := []string{file}
	if dev, ok := findDevice(filepath.Clean(file)); ok && dev.Images != nil {
		images = dev.Images()
	}
	for _, image := range images {
		open, writing := openHere(image)
		if open && (writing || hostWrites) {
			return fmt.Errorf("%w: %s, close it first", errImageOpen, filepath.Base(image))
		}
	}
	return nil
}

// servedBy returns the drive that serves file, directly or through a device
// such as the Ventoy disk, and whether the host can write it there. The
// caller holds driveMu.
func servedBy(file string) (string, bool, error) {
	holder, err := loadedDrive(file)
	if err != nil {
		return "", false, err
	}
	if holder != "" {
		d, err := findDrive(holder)
		if err != nil {
			return "", false, err
		}
		info, err := readDrive(d)
		if err != nil {
			return "", false, err
		}
		return holder, !info.Ro, nil
	}

	// A device reads its images through read-only loop devices.
	_, holder, err = deviceServing(file)
	if err != nil {
		return "", false, err
	}
	return holder, false, nil
}

// openImage opens the filesystem inside an image, and returns the drive that
// serves the image, if one does. Everything the file browser does goes
// through here, so the two ways to corrupt an image are refused in one place:
// writing it while the host can see it, and reading it while the host can
// write it, which would show a filesystem changing under the board's cache.
// The check and the mount are one step under driveMu, so a drive cannot take
// the image in between.
func openImage(file string, writable bool) (diskfs.Image, string, error) {
	resolved, err := resolveImage(file)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return nil, "", err
	}
	if !info.Mode().IsRegular() {
		return nil, "", errInvalidImage
	}

	driveMu.Lock()
	defer driveMu.Unlock()

	drive := ""
	hostWrites := false
	if !hidOnly() {
		if drive, hostWrites, err = servedBy(resolved); err != nil {
			return nil, "", err
		}
	}
	if writable && drive != "" {
		return nil, drive, fmt.Errorf("%w in the %s drive: eject it before changing its files", errImageServed, drive)
	}
	if hostWrites {
		return nil, drive, fmt.Errorf("%w in the %s drive and can write it: eject it, or insert it read-only, to browse it", errImageServed, drive)
	}

	image, err := mountImage(resolved, writable)
	if err != nil {
		return nil, drive, err
	}
	return image, drive, nil
}

func closeImage(image io.Closer) {
	if err := image.Close(); err != nil {
		log.Errorf("close image failed: %s", err)
	}
}

func (s *Service) ListFiles(c *gin.Context) {
	var rsp proto.Response

	file := c.Query("file")
	dir, err := utils.ValidateMediumPath(c.Query("path"))
	if err != nil {
		respondFileError(c, err)
		return
	}

	image, drive, err := openImage(file, false)
	if err != nil {
		respondFileError(c, err)
		return
	}
	defer closeImage(image)

	entries, truncated, err := image.List(dir, maxListEntries)
	if err != nil {
		respondFileError(c, fmt.Errorf("list %s: %w", dir, err))
		return
	}

	files := make([]proto.ImageFile, 0, len(entries))
	for _, entry := range entries {
		files = append(files, proto.ImageFile{
			Name:     entry.Name,
			Size:     entry.Size,
			IsDir:    entry.IsDir,
			Modified: entry.ModTime,
		})
	}

	total, free, _ := image.StatFS()

	rsp.OkRspWithData(c, &proto.ListFilesRsp{
		File:      file,
		Path:      dir,
		FS:        string(image.Kind()),
		Writable:  image.Kind().Writable() && drive == "",
		Drive:     drive,
		Entries:   files,
		Truncated: truncated,
		Free:      free,
		Total:     total,
	})
}

func (s *Service) DownloadFile(c *gin.Context) {
	var rsp proto.Response

	entryPath, err := utils.ValidateMediumPath(c.Query("path"))
	if err != nil {
		respondFileError(c, err)
		return
	}
	if entryPath == "/" {
		rsp.ErrRsp(c, codeInvalidArgument, "a file path is required")
		return
	}

	image, _, err := openImage(c.Query("file"), false)
	if err != nil {
		respondFileError(c, err)
		return
	}
	defer closeImage(image)

	file, entry, err := image.Open(entryPath)
	if err != nil {
		respondFileError(c, fmt.Errorf("open %s: %w", entryPath, err))
		return
	}
	defer func() {
		_ = file.Close()
	}()

	// ServeContent adds range support and the headers a browser download
	// needs; the file streams straight out of the mounted image.
	c.Header("Content-Disposition", attachmentHeader(path.Base(entryPath)))
	c.Header("Content-Type", "application/octet-stream")
	http.ServeContent(c.Writer, c.Request, entry.Name, entry.ModTime, file)
}

func (s *Service) UploadFile(c *gin.Context) {
	var rsp proto.Response

	// The image travels in the query string so the multipart body holds
	// nothing but the file: reading a form field first would make net/http
	// buffer the upload before the image could even be opened.
	file := c.Query("file")
	dir, err := utils.ValidateMediumPath(c.Query("path"))
	if err != nil {
		respondFileError(c, err)
		return
	}
	audit.SetTarget(c, file+": "+dir)

	// Content-Length covers the multipart framing as well, so it is an upper
	// bound on the file: good enough to refuse an upload that cannot fit.
	size := c.Request.ContentLength

	image, _, err := openImage(file, true)
	if err != nil {
		respondFileError(c, err)
		return
	}
	defer closeImage(image)

	if image.Kind().IsFAT() && size > diskfs.MaxFATFileSize+multipartSlack {
		respondFileError(c, fmt.Errorf("%w: %s holds no file of 4 GiB or more", errNoSpace, image.Kind()))
		return
	}

	// A chunked request states no length, so the free space is what caps it
	// instead: without that, an upload that never ends could fill the image.
	limit := int64(uploadMaxBytes)
	if _, free, err := image.StatFS(); err == nil {
		if size > 0 {
			if uint64(size) > free {
				respondFileError(c, fmt.Errorf("%w: it has %d MiB free", errNoSpace, free>>20))
				return
			}
		} else if free < uint64(limit) {
			limit = int64(free)
		}
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit+multipartSlack)

	reader, err := c.Request.MultipartReader()
	if err != nil {
		rsp.ErrRsp(c, codeInvalidArgument, "invalid multipart request")
		return
	}

	part, err := nextFilePart(reader)
	if err != nil {
		respondFileError(c, err)
		return
	}
	defer func() {
		_ = part.Close()
	}()

	name, err := uploadName(part.FileName())
	if err != nil {
		respondFileError(c, err)
		return
	}

	target := name
	if dir != "/" {
		target = path.Join(dir, name)
	}
	audit.SetTarget(c, file+": "+target)

	// An upload takes the single transfer slot, so the status endpoint can
	// report its progress and uploads write one at a time.
	done, err := transfers.begin(c.Request.Context(), filepath.Base(file)+": "+target)
	if err != nil {
		respondFileError(c, err)
		return
	}

	progress := utils.NewProgressWriter(io.Discard, size, func(loaded, total int64) {
		transfers.progress(done, loaded, total)
	})
	defer progress.Stop()

	var src io.Reader = io.TeeReader(part, progress)
	if image.Kind().IsFAT() {
		src = &capReader{r: src, left: diskfs.MaxFATFileSize}
	}

	written, err := image.Write(target, src)
	if err != nil {
		transfers.finish(done, jobStateFailed, err.Error())

		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			err = fmt.Errorf("%w: the upload is larger than the image can hold", errNoSpace)
		}
		respondFileError(c, err)
		return
	}
	transfers.finish(done, jobStateSuccess, "")

	rsp.OkRspWithData(c, &proto.ImageFile{Name: name, Size: written})
	log.Infof("wrote %s (%d bytes) into %s", target, written, file)
}

func (s *Service) DeleteFile(c *gin.Context) {
	var req proto.ImageFileReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, codeInvalidArgument, "invalid arguments")
		return
	}

	entryPath, err := utils.ValidateMediumPath(req.Path)
	if err != nil {
		respondFileError(c, err)
		return
	}
	if entryPath == "/" {
		rsp.ErrRsp(c, codeInvalidArgument, "a file path is required")
		return
	}
	audit.SetTarget(c, req.File+": "+entryPath)

	// A directory goes with everything below it, the temporary of an upload
	// into it included.
	if running, ok := transfers.active(); ok {
		prefix := filepath.Base(req.File) + ": "
		if inside, found := strings.CutPrefix(running, prefix); found && pathWithin(inside, entryPath) {
			respondFileError(c, fmt.Errorf("%w: an upload into %s is running", diskfs.ErrBusy, entryPath))
			return
		}
	}

	image, _, err := openImage(req.File, true)
	if err != nil {
		respondFileError(c, err)
		return
	}
	defer closeImage(image)

	if err := image.Remove(entryPath); err != nil {
		respondFileError(c, fmt.Errorf("remove %s: %w", entryPath, err))
		return
	}

	rsp.OkRsp(c)
	log.Infof("deleted %s from %s", entryPath, req.File)
}

func (s *Service) FileJobStatus(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, transfers.status())
}

// nextFilePart walks the multipart body to the uploaded file, skipping any
// field that came before it.
func nextFilePart(reader *multipart.Reader) (*multipart.Part, error) {
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			return nil, fmt.Errorf("%w: no file in the request", utils.ErrInvalidMediaPath)
		}
		if err != nil {
			return nil, fmt.Errorf("read multipart request: %w", err)
		}
		if part.FormName() != "file" {
			_ = part.Close()
			continue
		}
		return part, nil
	}
}

// uploadName reduces the browser supplied file name to something the image
// can hold. The client picks it, so it is treated as untrusted.
func uploadName(raw string) (string, error) {
	name := path.Base(strings.ReplaceAll(raw, `\`, "/"))
	if err := utils.ValidateFileName(name); err != nil {
		return "", err
	}
	return name, nil
}

// capReader fails a read past left bytes, for a filesystem that cannot hold a
// larger file, rather than letting the write fail at the limit with an error
// that does not say why.
type capReader struct {
	r    io.Reader
	left int64
}

func (c *capReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.left -= int64(n)
	if c.left < 0 {
		return n, fmt.Errorf("%w: FAT holds no file of 4 GiB or more", errNoSpace)
	}
	return n, err
}

// pathWithin reports whether p is dir or lies below it. Both are compared
// upper cased, the way FAT and exFAT match names.
func pathWithin(p, dir string) bool {
	p = strings.ToUpper(strings.TrimPrefix(p, "/"))
	dir = strings.ToUpper(strings.Trim(dir, "/"))
	return p == dir || strings.HasPrefix(p, dir+"/")
}

// attachmentHeader builds the Content-Disposition of a download the way
// browsers want it: an ASCII fallback for everything old, plus the real name
// percent-encoded, because a file name is not necessarily ASCII.
func attachmentHeader(name string) string {
	fallback := strings.Map(func(r rune) rune {
		if r < 0x20 || r > 0x7e || r == '"' || r == '\\' {
			return '_'
		}
		return r
	}, name)

	return fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", fallback, rfc5987Encode(name))
}

// rfc5987Encode percent-encodes every byte that is not an RFC 5987 attr-char.
// url.PathEscape is not that set: it leaves "=", ":", "@" and friends alone,
// which a strict parser reads as the end of the value.
func rfc5987Encode(value string) string {
	const hex = "0123456789ABCDEF"

	var encoded strings.Builder
	for i := 0; i < len(value); i++ {
		b := value[i]
		if isAttrChar(b) {
			encoded.WriteByte(b)
			continue
		}
		encoded.WriteByte('%')
		encoded.WriteByte(hex[b>>4])
		encoded.WriteByte(hex[b&0x0f])
	}
	return encoded.String()
}

func isAttrChar(b byte) bool {
	switch {
	case b >= 'A' && b <= 'Z', b >= 'a' && b <= 'z', b >= '0' && b <= '9':
		return true
	}
	return strings.IndexByte("!#$&+-.^_`|~", b) >= 0
}

func respondFileError(c *gin.Context, err error) {
	var rsp proto.Response

	switch {
	case errors.Is(err, utils.ErrInvalidMediaPath), errors.Is(err, errInvalidImage):
		rsp.ErrRsp(c, codeInvalidArgument, err.Error())
	case errors.Is(err, errImageServed), errors.Is(err, diskfs.ErrBusy),
		errors.Is(err, errImageOpen), errors.Is(err, errTooManyJobs):
		rsp.ErrRsp(c, codeMediumBusy, err.Error())
	case errors.Is(err, diskfs.ErrUnsupported):
		rsp.ErrRsp(c, codeNotBrowsable, err.Error())
	case errors.Is(err, errNoSpace):
		rsp.ErrRsp(c, codeNoSpace, err.Error())
	case errors.Is(err, diskfs.ErrReadOnly):
		rsp.ErrRsp(c, codeReadOnly, err.Error())
	case errors.Is(err, fs.ErrNotExist):
		rsp.ErrRsp(c, codeOperationFailed, "no such file or directory")
	case errors.Is(err, fs.ErrExist):
		rsp.ErrRsp(c, codeOperationFailed, "a file with that name already exists")
	default:
		rsp.ErrRsp(c, codeOperationFailed, err.Error())
	}
}
