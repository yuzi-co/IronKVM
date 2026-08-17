package application

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	buildversion "NanoKVM-Server/common/version"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type Latest struct {
	ManifestVersion   int    `json:"manifest_version,omitempty"`
	Version           string `json:"version"`
	Name              string `json:"name"`
	Sha512            string `json:"sha512"`
	LegacySize        uint64 `json:"size"`
	SizeBytes         uint64 `json:"size_bytes,omitempty"`
	UnpackedSizeBytes uint64 `json:"unpacked_size_bytes,omitempty"`

	// ManifestURL lets a feed serve its packages from a different host than the
	// manifest. It is optional: without it the package is looked for beside the
	// manifest, which is what every feed did before this field existed.
	ManifestURL string `json:"url,omitempty"`

	// Url is the resolved download location. It is never read from the document.
	Url string `json:"-"`
}

const (
	maxLatestJSONSize = 64 * 1024
)

var (
	latestClient = utils.NewUpdateHTTPClient(15 * time.Second)
	// Both prefixes, because a board must be able to leave as easily as it
	// arrived: IronKVM ships its own packages, and an official Sipeed package
	// stays installable through the same updater.
	packageNamePattern = regexp.MustCompile(`^(?:nanokvm|ironkvm)_([0-9]+\.[0-9]+\.[0-9]+)\.tar\.gz$`)
	versionPattern     = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)
)

// packageVersion reports the version a package file name states.
//
// An offline upload carries no manifest, so this is the only statement of which
// version the file claims to be, and it is what the extracted package's own
// version file is checked against. The product half of the name is not fixed:
// this fork ships ironkvm_ and Sipeed ships nanokvm_, and both install through
// the same updater. Stripping one literal prefix left an IronKVM package
// expecting the version "ironkvm_1.0.0", which no version file can match, so
// every offline upload of an IronKVM release was refused as a layout error.
//
// The name is matched against the same pattern that guards the download, so
// this cannot become a second place that decides which names are acceptable.
func packageVersion(name string) (string, bool) {
	match := packageNamePattern.FindStringSubmatch(name)
	if match == nil {
		return "", false
	}
	return match[1], true
}

// versionFile is written by the updater. Tests point it elsewhere.
var versionFile = fmt.Sprintf("%s/version", AppDir)

// currentVersion reports the installed application version, carrying the stamp
// of the binary serving it.
func currentVersion() string {
	version := "1.0.0"

	if content, err := os.ReadFile(versionFile); err == nil {
		version = strings.ReplaceAll(string(content), "\n", "")
	}

	return buildversion.Decorate(version)
}

func (s *Service) GetVersion(c *gin.Context) {
	var rsp proto.Response

	currentVersion := currentVersion()

	log.Debugf("current version: %s", currentVersion)

	// latest version
	latestVersion := ""
	latest, err := getLatest()
	if err != nil {
		log.Errorf("failed to get latest version: %s", err)
		rsp.ErrRsp(c, -1, "failed to query latest version")
		return
	}
	latestVersion = latest.Version

	rsp.OkRspWithData(c, &proto.GetVersionRsp{
		Current: currentVersion,
		Latest:  latestVersion,
	})
}

func getLatest() (*Latest, error) {
	baseURL, err := resolveUpdateBaseURL()
	if err != nil {
		return nil, err
	}

	// latestClient carries both the proxy and the redirect rule, so the manifest
	// fetch reaches a custom update server the same way the download does.
	manifestURL, err := joinUpdateURL(baseURL, "latest.json")
	if err != nil {
		return nil, err
	}
	parsedManifestURL, err := url.Parse(manifestURL)
	if err != nil {
		return nil, err
	}
	query := parsedManifestURL.Query()
	query.Set("now", fmt.Sprintf("%d", time.Now().Unix()))
	parsedManifestURL.RawQuery = query.Encode()

	request, err := utils.NewAuthenticatedRequest("GET", parsedManifestURL.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := latestClient.Do(request)
	if err != nil {
		log.Debugf("failed to request version from %s", parsedManifestURL.Redacted())
		return nil, errors.New("update server is inaccessible")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxLatestJSONSize+1))
	if err != nil {
		log.Errorf("failed to read response: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Errorf("server responded with status code: %d", resp.StatusCode)
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	if len(body) > maxLatestJSONSize {
		return nil, fmt.Errorf("latest manifest exceeds %d bytes", maxLatestJSONSize)
	}

	latest, err := parseLatest(body, baseURL)
	if err != nil {
		return nil, err
	}

	log.Debugf("get application latest version: %s", latest.Version)
	return latest, nil
}

// parseLatest reads the manifest and refuses anything it would not be safe to
// act on. The name decides both the URL and where the package is written, and
// it is written before the checksum has had a chance to reject the package, so
// it has to be a plain file name.
func parseLatest(body []byte, baseURL string) (*Latest, error) {
	var latest Latest
	if err := json.Unmarshal(body, &latest); err != nil {
		log.Errorf("failed to unmarshal response: %s", err)
		return nil, err
	}
	if err := validateLatest(&latest); err != nil {
		return nil, err
	}

	// validateLatest has already refused any name that is not
	// (nanokvm|ironkvm)_X.Y.Z.tar.gz, so the name cannot carry a path separator by the
	// time it reaches the URL or the file on disk. It bounds size_bytes for a
	// version 2 manifest, but nothing bounds the legacy size, and the package
	// is written to the SD card the device boots from. Refuse an oversized
	// manifest here rather than starting a download that cannot finish.
	if latest.LegacySize > maxPackageSize {
		log.Errorf("refusing update package of %d bytes", latest.LegacySize)
		return nil, fmt.Errorf("package is too large")
	}

	resolved, err := resolveDownloadURL(latest.ManifestURL, baseURL, latest.Name)
	if err != nil {
		return nil, err
	}
	latest.Url = resolved

	return &latest, nil
}

// resolveDownloadURL decides where the package is fetched from.
//
// A manifest may name the package outright, which is what lets the manifest and
// the package live on different hosts: a few hundred bytes of JSON and a 26 MB
// tarball do not want the same kind of hosting, and a release asset lives under
// a per-tag path that no fixed base URL can reach.
//
// It must be absolute and it must be https. The sha512 that would catch a
// substituted package comes from this same document, so transport security is
// the only thing between the device and whatever is on the path.
//
// Naming the package grants no trust the manifest did not already have. A
// hostile feed could always serve any bytes it liked from its own directory.
func resolveDownloadURL(manifestURL string, baseURL string, name string) (string, error) {
	if manifestURL == "" {
		return joinUpdateURL(baseURL, name)
	}

	parsed, err := url.Parse(manifestURL)
	if err != nil {
		return "", fmt.Errorf("invalid update package url: %w", err)
	}
	if parsed.Scheme != "https" {
		return "", errors.New("update package url must use https")
	}
	if parsed.Host == "" {
		return "", errors.New("update package url must be absolute")
	}

	return parsed.String(), nil
}

func joinUpdateURL(baseURL string, element string) (string, error) {
	joined, err := url.JoinPath(baseURL, element)
	if err != nil {
		return "", fmt.Errorf("join update URL: %w", err)
	}
	return joined, nil
}

func validateLatest(latest *Latest) error {
	if !versionPattern.MatchString(latest.Version) {
		return errors.New("invalid latest version")
	}
	if !packageNamePattern.MatchString(latest.Name) {
		return errors.New("invalid update package name")
	}
	digest, err := base64.StdEncoding.DecodeString(latest.Sha512)
	if err != nil || len(digest) != 64 {
		return errors.New("invalid update package sha512")
	}
	if latest.LegacySize == 0 {
		return errors.New("invalid update package size")
	}
	switch latest.ManifestVersion {
	case 0, 1:
		return nil
	case 2:
		if latest.SizeBytes == 0 || latest.SizeBytes > maxPackageSize {
			return errors.New("invalid update package size_bytes")
		}
		if latest.UnpackedSizeBytes == 0 || latest.UnpackedSizeBytes > maxExpandedSize {
			return errors.New("invalid update package unpacked_size_bytes")
		}
		return nil
	default:
		return errors.New("unsupported update manifest version")
	}
}

func preflightManifestSpace(path string, latest *Latest) error {
	if latest.ManifestVersion != 2 {
		return nil
	}
	return ensureFreeSpace(path, latest.SizeBytes)
}

func validateDownloadedSize(latest *Latest, written uint64) error {
	if written > maxPackageSize {
		return fmt.Errorf("update package exceeds %d bytes", maxPackageSize)
	}
	if latest.ManifestVersion == 2 && written != latest.SizeBytes {
		return fmt.Errorf("update package size mismatch: expected %d bytes, got %d", latest.SizeBytes, written)
	}
	return nil
}

func validateExpandedSize(latest *Latest, expanded uint64) error {
	if latest.ManifestVersion == 2 && expanded != latest.UnpackedSizeBytes {
		return fmt.Errorf("update package expanded size mismatch: expected %d bytes, got %d", latest.UnpackedSizeBytes, expanded)
	}
	return nil
}
