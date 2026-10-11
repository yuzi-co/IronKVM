package webui

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/application"
	"NanoKVM-Server/utils"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// FeedFile is the UI's document in the update feed, beside latest.json.
const FeedFile = "ui-latest.json"

const (
	maxFeedBytes  = 16 << 10
	uploadPrefix  = ".upload-"
	signatureExt  = ".minisig"
	maxUploadBody = maxPackageBytes + maxSigFileSize + (1 << 20)
)

// Latest is ui-latest.json:
//
//	{
//	  "format": 1,
//	  "version": "1.2.1",
//	  "name": "ironkvm-ui_1.2.1.tar.gz",
//	  "url": "https://.../ironkvm-ui_1.2.1.tar.gz",   optional, as in latest.json
//	  "size_bytes": 5242880,
//	  "sha512": "<base64>",
//	  "api": {"min": 1, "max": 1}
//	}
//
// The signature is fetched from the package's URL with .minisig appended. The
// feed's api range lets the page say whether the package fits before it is
// downloaded; the manifest inside the package is what the install trusts.
type Latest struct {
	Format    int      `json:"format"`
	Version   string   `json:"version"`
	Name      string   `json:"name"`
	URL       string   `json:"url,omitempty"`
	SizeBytes uint64   `json:"size_bytes"`
	Sha512    string   `json:"sha512"`
	API       APIRange `json:"api"`
}

func (l *Latest) validate() error {
	if l.Format != ManifestFormat {
		return fmt.Errorf("unsupported %s format %d", FeedFile, l.Format)
	}
	if !versionPattern.MatchString(l.Version) {
		return errors.New("invalid UI version in the feed")
	}
	if !FileNamePattern.MatchString(l.Name) {
		return errors.New("invalid UI package name in the feed")
	}
	if l.SizeBytes == 0 || l.SizeBytes > maxPackageBytes {
		return errors.New("invalid UI package size in the feed")
	}
	if d, err := base64.StdEncoding.DecodeString(l.Sha512); err != nil || len(d) != sha512.Size {
		return errors.New("invalid UI package sha512 in the feed")
	}
	if l.API.Min < 1 || l.API.Max < l.API.Min {
		return errors.New("invalid API range in the feed")
	}
	return nil
}

// Service is the HTTP side of a Store.
type Service struct {
	store *Store
	// fetchFeed and download are variables so tests run without a network.
	fetchFeed func() (*Latest, string, error)
	download  func(url, target string, maxBytes int64) error
}

// NewService serves store.
func NewService(store *Store) *Service {
	return &Service{store: store, fetchFeed: fetchLatest, download: downloadTo}
}

func fetchLatest() (*Latest, string, error) {
	body, baseURL, err := application.FetchFeedDocument(FeedFile, maxFeedBytes)
	if err != nil {
		return nil, "", err
	}
	var l Latest
	if err := json.Unmarshal(body, &l); err != nil {
		return nil, "", fmt.Errorf("read %s: %w", FeedFile, err)
	}
	if err := l.validate(); err != nil {
		return nil, "", err
	}
	u, err := application.ResolveFeedURL(l.URL, baseURL, l.Name)
	if err != nil {
		return nil, "", err
	}
	return &l, u, nil
}

var downloadClient = utils.NewUpdateHTTPClient(5 * time.Minute)

func downloadTo(url, target string, maxBytes int64) error {
	req, err := utils.NewAuthenticatedRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := downloadClient.Do(req)
	if err != nil {
		return errors.New("the update server is inaccessible")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the update server answered %d", resp.StatusCode)
	}
	if resp.ContentLength > maxBytes {
		return fmt.Errorf("the download exceeds %d bytes", maxBytes)
	}
	return writeLimited(resp.Body, target, maxBytes)
}

func writeLimited(r io.Reader, target string, maxBytes int64) error {
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	n, err := io.Copy(out, io.LimitReader(r, maxBytes+1))
	closeErr := out.Close()
	if err == nil {
		err = closeErr
	}
	if err == nil && n > maxBytes {
		err = fmt.Errorf("the file exceeds %d bytes", maxBytes)
	}
	if err != nil {
		_ = os.Remove(target)
	}
	return err
}

// GetStatus answers GET /api/ui.
func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.store.Status())
}

// LatestRsp is GET /api/ui/latest's answer.
type LatestRsp struct {
	Version string   `json:"version"`
	API     APIRange `json:"api"`
	// Compatible is whether the package states this server's API version.
	Compatible bool `json:"compatible"`
	// Active is whether the served UI already is this version.
	Active bool `json:"active"`
}

// GetLatest answers GET /api/ui/latest from the feed.
func (s *Service) GetLatest(c *gin.Context) {
	var rsp proto.Response
	l, _, err := s.fetchFeed()
	if err != nil {
		log.Debugf("webui: feed: %v", err)
		rsp.ErrRsp(c, -1, fmt.Sprintf("no UI package in the feed: %s", err))
		return
	}
	rsp.OkRspWithData(c, LatestRsp{
		Version:    l.Version,
		API:        l.API,
		Compatible: l.API.Contains(s.store.API),
		Active:     s.store.Status().Active.Version == l.Version,
	})
}

// Update answers POST /api/ui/update: download the feed's package and its
// signature, and install it.
func (s *Service) Update(c *gin.Context) {
	var rsp proto.Response
	result, err := s.withUpdateLock(func(work string) (InstallResult, error) {
		l, pkgURL, err := s.fetchFeed()
		if err != nil {
			return InstallResult{}, err
		}
		if !l.API.Contains(s.store.API) {
			return InstallResult{}, fmt.Errorf("UI %s supports server API %d..%d, and this server speaks %d",
				l.Version, l.API.Min, l.API.Max, s.store.API)
		}
		if free, err := freeBytes(work); err == nil && free < l.SizeBytes+freeReserve {
			return InstallResult{}, fmt.Errorf("not enough space to download %d MiB", l.SizeBytes>>20)
		}
		pkg := filepath.Join(work, l.Name)
		if err := s.download(pkgURL, pkg, int64(l.SizeBytes)); err != nil {
			return InstallResult{}, fmt.Errorf("download: %w", err)
		}
		if err := checkSha512(pkg, l.Sha512, l.SizeBytes); err != nil {
			return InstallResult{}, err
		}
		sig := pkg + signatureExt
		if err := s.download(pkgURL+signatureExt, sig, maxSigFileSize); err != nil {
			// A feed without signatures is the owner's choice as much as
			// an empty keyring is; the policy decides whether that matters.
			log.Warnf("webui: no signature for %s: %v", l.Name, err)
		}
		return s.store.Install(pkg, sig)
	})
	if err != nil {
		rsp.ErrRsp(c, -1, fmt.Sprintf("UI update failed: %s", err))
		return
	}
	rsp.OkRspWithData(c, result)
}

func checkSha512(path, want string, size uint64) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if uint64(info.Size()) != size {
		return fmt.Errorf("the package is %d bytes and the feed says %d", info.Size(), size)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if base64.StdEncoding.EncodeToString(h.Sum(nil)) != want {
		return errors.New("the package does not match the feed's sha512")
	}
	return nil
}

// Upload answers POST /api/ui/upload: a multipart form with the package in
// "package" and its signature in "signature".
func (s *Service) Upload(c *gin.Context) {
	var rsp proto.Response
	if c.Request.ContentLength > maxUploadBody {
		rsp.ErrRsp(c, -1, "the upload is too large")
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBody)

	result, err := s.withUpdateLock(func(work string) (InstallResult, error) {
		reader, err := c.Request.MultipartReader()
		if err != nil {
			return InstallResult{}, fmt.Errorf("invalid multipart data: %w", err)
		}
		pkg, sig, err := receiveUpload(reader, work)
		if err != nil {
			return InstallResult{}, err
		}
		return s.store.Install(pkg, sig)
	})
	if err != nil {
		rsp.ErrRsp(c, -1, fmt.Sprintf("UI install failed: %s", err))
		return
	}
	rsp.OkRspWithData(c, result)
}

// receiveUpload writes the two parts into work. The package keeps its name,
// which must be a UI package name, and the signature is put beside it under
// the name a signature has.
func receiveUpload(reader *multipart.Reader, work string) (string, string, error) {
	var pkg, sig string
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", "", fmt.Errorf("read upload: %w", err)
		}
		switch part.FormName() {
		case "package":
			if pkg != "" {
				return "", "", errors.New("more than one package in the upload")
			}
			name := filepath.Base(part.FileName())
			if !FileNamePattern.MatchString(name) {
				return "", "", fmt.Errorf("%q is not a UI package name (ironkvm-ui_<version>.tar.gz)", part.FileName())
			}
			pkg = filepath.Join(work, name)
			if err := writeLimited(part, pkg, maxPackageBytes); err != nil {
				return "", "", err
			}
		case "signature":
			if sig != "" {
				return "", "", errors.New("more than one signature in the upload")
			}
			sig = filepath.Join(work, "upload"+signatureExt)
			if err := writeLimited(part, sig, maxSigFileSize); err != nil {
				return "", "", err
			}
		}
		_ = part.Close()
	}
	if pkg == "" {
		return "", "", errors.New("the upload has no package")
	}
	final := pkg + signatureExt
	if sig != "" {
		if err := os.Rename(sig, final); err != nil {
			return "", "", err
		}
	}
	return pkg, final, nil
}

// Rollback answers POST /api/ui/rollback.
func (s *Service) Rollback(c *gin.Context) {
	var rsp proto.Response
	if !application.AcquireUpdateLock() {
		rsp.ErrRsp(c, -1, "an update is in progress")
		return
	}
	defer application.ReleaseUpdateLock()
	info, err := s.store.Rollback()
	if err != nil {
		rsp.ErrRsp(c, -1, fmt.Sprintf("UI rollback failed: %s", err))
		return
	}
	rsp.OkRspWithData(c, info)
}

// withUpdateLock runs fn with the application update lock held, in a work
// directory under the store's root that is removed afterwards. The root is on
// the filesystem the release is unpacked to.
func (s *Service) withUpdateLock(fn func(work string) (InstallResult, error)) (InstallResult, error) {
	if !application.AcquireUpdateLock() {
		return InstallResult{}, errors.New("an update is in progress")
	}
	defer application.ReleaseUpdateLock()

	if err := os.MkdirAll(s.store.Root, 0o755); err != nil {
		return InstallResult{}, err
	}
	work, err := os.MkdirTemp(s.store.Root, uploadPrefix)
	if err != nil {
		return InstallResult{}, err
	}
	defer func() { _ = os.RemoveAll(work) }()
	return fn(work)
}
