package storage

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/storage/mkfs"
	"NanoKVM-Server/utils"
)

// Exchange images: an empty disk image an admin creates to swap files with
// the host. It lands in the image list like any other, goes into the disk
// drive, and its files are browsed and changed through the file browser.

// reservedFreeBytes is kept free on /data when sizing an exchange image, as
// downloads keep it: the image is sparse, and the host filling it must not
// fill the card the board runs from.
const reservedFreeBytes = 256 << 20

// freeBytes reports the space left on the filesystem holding path. Tests
// replace it.
var freeBytes = func(path string) (int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}

// maxExchangeImage is the largest exchange image the card can hold now.
func maxExchangeImage() (free, largest int64, err error) {
	free, err = freeBytes(imageRoot)
	if err != nil {
		return 0, 0, err
	}
	largest = min(max(free-reservedFreeBytes, 0), mkfs.MaxImageSize)
	return free, largest &^ (1<<20 - 1), nil
}

// exchangeImagePath checks the name an admin gave a new image and returns
// where it goes: in the image directory itself, with ".img" added when the
// name has no such suffix. The name has to pass the checks a file name
// inside an image does, so every host the image reaches can show it.
func exchangeImagePath(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("%w: %q is a path, not a name", utils.ErrInvalidMediaPath, raw)
	}
	if strings.HasPrefix(name, ".") {
		return "", fmt.Errorf("%w: %q would be hidden", utils.ErrInvalidMediaPath, raw)
	}
	if err := utils.ValidateFileName(name); err != nil {
		return "", err
	}
	if strings.HasSuffix(strings.ToLower(name), ".iso") {
		return "", fmt.Errorf("%w: an exchange image is a disk, not a CD: name it .img", utils.ErrInvalidMediaPath)
	}
	if !strings.HasSuffix(strings.ToLower(name), ".img") {
		name += ".img"
		if err := utils.ValidateFileName(name); err != nil {
			return "", err
		}
	}
	return filepath.Join(imageRoot, name), nil
}

func (s *Service) CreateImage(c *gin.Context) {
	var req proto.CreateImageReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, codeInvalidArgument, "invalid arguments")
		return
	}

	path, err := exchangeImagePath(req.Name)
	if err != nil {
		respondFileError(c, err)
		return
	}
	audit.SetTarget(c, fmt.Sprintf("%s (%s, %d MiB)", path, req.FS, req.Size>>20))

	fsType := mkfs.FS(strings.ToLower(req.FS))
	if fsType == "" {
		fsType = mkfs.ExFAT
	}
	if fsType != mkfs.ExFAT && fsType != mkfs.FAT32 {
		rsp.ErrRsp(c, codeInvalidArgument, fmt.Sprintf("unknown filesystem %q: exfat or fat32", req.FS))
		return
	}

	if req.Size < mkfs.MinImageSize || req.Size%(1<<20) != 0 {
		rsp.ErrRsp(c, codeInvalidArgument, fmt.Sprintf("the size must be whole MiB, %d MiB at least", mkfs.MinImageSize>>20))
		return
	}
	_, largest, err := maxExchangeImage()
	if err != nil {
		rsp.ErrRsp(c, codeOperationFailed, "could not read the free space: "+err.Error())
		return
	}
	if req.Size > largest {
		respondFileError(c, fmt.Errorf("%w: the card has room for %d MiB at most", errNoSpaceOnCard, largest>>20))
		return
	}

	label := mkfs.LabelFor(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), fsType)
	if err := mkfs.CreateImage(path, req.Size, fsType, label); err != nil {
		log.Errorf("create image %s failed: %s", path, err)
		respondFileError(c, err)
		return
	}

	rsp.OkRspWithData(c, &proto.CreateImageRsp{File: path})
	log.Infof("created %s: %d MiB, %s", path, req.Size>>20, fsType)
}

// errNoSpaceOnCard refuses an image larger than /data can hold. It counts as
// errNoSpace, so it answers with the same code.
var errNoSpaceOnCard error = noSpaceOnCard{}

type noSpaceOnCard struct{}

func (noSpaceOnCard) Error() string        { return "not enough space on the card" }
func (noSpaceOnCard) Is(target error) bool { return target == errNoSpace }

func (s *Service) GetImageSpace(c *gin.Context) {
	var rsp proto.Response

	free, largest, err := maxExchangeImage()
	if err != nil {
		rsp.ErrRsp(c, codeOperationFailed, "could not read the free space: "+err.Error())
		return
	}
	rsp.OkRspWithData(c, &proto.ImageSpaceRsp{Min: mkfs.MinImageSize, Max: largest, Free: free})
}
