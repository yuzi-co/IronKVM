package settingsio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// Service answers the settings export and import routes under
// /api/settings. Every route needs an administrator.
type Service struct {
	Device *Device
	// Source names this board in an exported file.
	Source func() Source
	Now    func() time.Time
}

// NewService returns the service for the board.
func NewService(source func() Source) *Service {
	return &Service{Device: &Device{}, Source: source, Now: time.Now}
}

// The error codes of the routes, beside the message.
const (
	codeBadRequest = -1
	codeRefused    = -2
	codeInvalid    = -3
	codeWrite      = -4
)

// fileNameUnsafe is what a host name must not carry into a file name.
var fileNameUnsafe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// FileName is the name of an exported file: the host and the day.
func FileName(prefix, host string, now time.Time, ext string) string {
	host = fileNameUnsafe.ReplaceAllString(host, "-")
	if host == "" || host == "-" {
		host = "board"
	}
	return fmt.Sprintf("%s-%s-%s.%s", prefix, host, now.UTC().Format("20060102-150405"), ext)
}

// Export answers GET /api/settings/export with the settings file.
func (s *Service) Export(c *gin.Context) {
	source := s.Source()
	now := s.Now()
	doc, skipped, err := s.Device.Export(source, now)
	if err != nil {
		log.Errorf("settings export: %s", err)
		var rsp proto.Response
		rsp.ErrRsp(c, codeWrite, "the export failed")
		return
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		var rsp proto.Response
		rsp.ErrRsp(c, codeWrite, "the export failed")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, FileName("ironkvm-settings", source.Hostname, now, "json")))
	for _, name := range skipped {
		c.Writer.Header().Add("X-Settings-Skipped", name)
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", append(data, '\n'))
}

// importRequest carries a settings file and the sections to import. The
// preview reads only the document.
type importRequest struct {
	Document json.RawMessage `json:"document"`
	Sections []string        `json:"sections"`
}

func (s *Service) readRequest(c *gin.Context) (*importRequest, *Document, bool) {
	var rsp proto.Response

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, MaxDocumentBytes+64<<10))
	if err != nil {
		rsp.ErrRsp(c, codeBadRequest, "the request could not be read")
		return nil, nil, false
	}
	var req importRequest
	if err := json.Unmarshal(body, &req); err != nil || len(req.Document) == 0 {
		rsp.ErrRsp(c, codeBadRequest, "invalid arguments")
		return nil, nil, false
	}
	doc, err := ParseDocument(req.Document)
	if err != nil {
		rsp.ErrRsp(c, codeRefused, err.Error())
		return nil, nil, false
	}
	return &req, doc, true
}

// Preview answers POST /api/settings/preview with what the import would
// change. It writes nothing.
func (s *Service) Preview(c *gin.Context) {
	_, doc, ok := s.readRequest(c)
	if !ok {
		return
	}
	var rsp proto.Response
	rsp.OkRspWithData(c, s.Device.Preview(doc))
}

// ImportResult says what an import wrote.
type ImportResult struct {
	Applied []string `json:"applied"`
	// RebootRequired is always true: the running services read most of
	// these settings only when they start.
	RebootRequired bool `json:"rebootRequired"`
}

// Import answers POST /api/settings/import.
func (s *Service) Import(c *gin.Context) {
	req, doc, ok := s.readRequest(c)
	if !ok {
		return
	}
	var rsp proto.Response

	applied, err := s.Device.Import(doc, req.Sections)
	var importErr *ImportError
	switch {
	case errors.As(err, &importErr):
		rsp.ErrRsp(c, codeInvalid, err.Error())
		return
	case err != nil:
		log.Errorf("settings import: %s", err)
		rsp.ErrRsp(c, codeWrite, err.Error())
		return
	}

	log.Infof("settings import: wrote %v", applied)
	rsp.OkRspWithData(c, ImportResult{Applied: applied, RebootRequired: true})
}
