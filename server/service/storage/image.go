package storage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/utils"
)

const (
	imageDirectory = "/data"
	// legacyNoImageDevice is what older builds wrote to mean "no image". It is
	// only ever read now, never written.
	legacyNoImageDevice = "/dev/mmcblk0p3"
)

var errNoCdDrive = errors.New("no CD drive")

// hidOnly reports whether the gadget runs without its mass storage function.
// Tests replace it.
var hidOnly = func() bool {
	mode, err := hid.GetMode()
	return err == nil && mode == hid.ModeHidOnly
}

// isMountableImage reports whether a client-supplied path may be handed to the
// USB mass storage gadget. Without this check any file or block device on the
// KVM - the raw eMMC, /etc/shadow - can be exported to the attached machine.
func isMountableImage(path string) bool {
	if !utils.IsPathInside(imageDirectory, path) {
		return false
	}

	name := strings.ToLower(path)

	return strings.HasSuffix(name, ".iso") || strings.HasSuffix(name, ".img")
}

func (s *Service) GetImages(c *gin.Context) {
	var rsp proto.Response
	var images []string

	err := filepath.Walk(imageDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			name := strings.ToLower(info.Name())
			if strings.HasSuffix(name, ".iso") || strings.HasSuffix(name, ".img") {
				images = append(images, path)
			}
		}

		return nil
	})
	if err != nil {
		rsp.ErrRsp(c, -2, "get images failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetImagesRsp{
		Files: images,
	})
	log.Debugf("get images success, total %d", len(images))
}

// normalizeMountedImage reads back what the gadget is currently serving.
// Devices that have not rebooted since this change still hold the old eMMC
// fallback, which has to keep reading as "nothing mounted".
func normalizeMountedImage(content string) string {
	image := strings.TrimSpace(content)
	if image == legacyNoImageDevice {
		return ""
	}
	return image
}

// legacyMount serves the single-drive mount call older clients make. An
// image goes into the CD drive when cdrom is set, otherwise into a writable
// disk. No image ejects both drives.
func legacyMount(file string, cdrom bool) error {
	if file == "" {
		for _, d := range driveDefs {
			if err := ejectDrive(d.id); err != nil && !errors.Is(err, errNoDrive) {
				return err
			}
		}
		return nil
	}

	if !cdrom {
		return insertDrive(DriveDisk, file, false)
	}

	// Switching lun.0 into a CD-ROM was the path this replaces, so a gadget
	// without lun.1 fails rather than falling back to it.
	err := insertDrive(DriveCdrom, file, true)
	if errors.Is(err, errNoDrive) {
		return errNoCdDrive
	}
	return err
}

// legacyMounted returns the CD's image if it has one, else the disk's.
func legacyMounted() (string, error) {
	drives, err := listDrives()
	if err != nil {
		return "", err
	}

	file := ""
	for _, d := range drives {
		if d.File == "" {
			continue
		}
		if d.ID == DriveCdrom {
			return d.File, nil
		}
		file = d.File
	}
	return file, nil
}

// legacyCdrom returns 1 when the CD drive holds a medium.
func legacyCdrom() (int64, error) {
	drives, err := listDrives()
	if err != nil {
		return 0, err
	}
	for _, d := range drives {
		if d.ID == DriveCdrom && d.File != "" {
			return 1, nil
		}
	}
	return 0, nil
}

func (s *Service) MountImage(c *gin.Context) {
	var req proto.MountImageReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if hidOnly() {
		rsp.ErrRsp(c, -2, errNoDrive.Error())
		return
	}

	if err := legacyMount(req.File, req.Cdrom); err != nil {
		log.Errorf("mount image %q (cdrom %t) failed: %s", req.File, req.Cdrom, err)
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	rsp.OkRsp(c)
	log.Debugf("mount image %s success", req.File)
}

func (s *Service) GetMountedImage(c *gin.Context) {
	var rsp proto.Response

	if hidOnly() {
		rsp.OkRspWithData(c, &proto.GetMountedImageRsp{File: ""})
		return
	}

	file, err := legacyMounted()
	if err != nil {
		rsp.ErrRsp(c, -2, "read failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetMountedImageRsp{File: file})
}

func (s *Service) GetCdRom(c *gin.Context) {
	var rsp proto.Response

	if hidOnly() {
		rsp.OkRspWithData(c, &proto.GetCdRomRsp{Cdrom: 0})
		return
	}

	cdrom, err := legacyCdrom()
	if err != nil {
		rsp.ErrRsp(c, -2, "read failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetCdRomRsp{Cdrom: cdrom})
}

func (s *Service) GetDrives(c *gin.Context) {
	var rsp proto.Response

	if hidOnly() {
		rsp.OkRspWithData(c, &proto.GetDrivesRsp{Drives: []proto.DriveInfo{}})
		return
	}

	drives, err := listDrives()
	if err != nil {
		log.Errorf("read drives failed: %s", err)
		rsp.ErrRsp(c, -2, "read drives failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetDrivesRsp{Drives: drives})
}

func (s *Service) InsertDrive(c *gin.Context) {
	var req proto.InsertDriveReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if hidOnly() {
		rsp.ErrRsp(c, -2, errNoDrive.Error())
		return
	}

	id := c.Param("id")
	if err := insertDrive(id, req.File, req.Ro); err != nil {
		log.Errorf("insert %q into %q failed: %s", req.File, id, err)
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	rsp.OkRsp(c)
	log.Debugf("inserted %s into %s", req.File, id)
}

func (s *Service) EjectDrive(c *gin.Context) {
	var rsp proto.Response

	if hidOnly() {
		rsp.ErrRsp(c, -2, errNoDrive.Error())
		return
	}

	id := c.Param("id")
	if err := ejectDrive(id); err != nil {
		log.Errorf("eject %q failed: %s", id, err)
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	rsp.OkRsp(c)
	log.Debugf("ejected %s", id)
}

func (s *Service) DeleteImage(c *gin.Context) {
	var req proto.DeleteImageReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	// The previous check ran HasPrefix on a lowercased copy while removing the
	// original path, so "/data/../root/x.iso" passed it.
	if !isMountableImage(req.File) {
		rsp.ErrRsp(c, -2, "invalid arguments")
		return
	}

	if err := removeImage(req.File); err != nil {
		log.Errorf("failed to remove file %s: %s", req.File, err)
		if errors.Is(err, errImageLoaded) {
			rsp.ErrRsp(c, -4, err.Error())
			return
		}
		rsp.ErrRsp(c, -3, "remove file failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("delete image %s success", req.File)
}
