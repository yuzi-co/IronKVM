package redfish

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/storage"
	"NanoKVM-Server/utils"
)

const (
	insertAction = "VirtualMedia.InsertMedia"
	ejectAction  = "VirtualMedia.EjectMedia"
)

// mediaDef maps a VirtualMedia resource to one of the storage drives.
type mediaDef struct {
	id    string
	drive string
	name  string
	types []string
}

var mediaDefs = []mediaDef{
	{id: "Cd", drive: storage.DriveCdrom, name: "Virtual CD", types: []string{"CD", "DVD"}},
	{id: "Disk", drive: storage.DriveDisk, name: "Virtual Disk", types: []string{"USBStick"}},
}

func mediaDefByID(id string) (mediaDef, bool) {
	for _, def := range mediaDefs {
		if def.id == id {
			return def, true
		}
	}
	return mediaDef{}, false
}

// driveFor returns the drive behind a VirtualMedia resource, if the gadget
// has it. On failure it has answered already.
func (s *Service) driveFor(c *gin.Context) (mediaDef, proto.DriveInfo, bool) {
	def, ok := mediaDefByID(c.Param("id"))
	if !ok {
		notFound(c)
		return mediaDef{}, proto.DriveInfo{}, false
	}

	drives, err := s.deps.ListDrives()
	if err != nil {
		log.Errorf("redfish: list drives: %s", err)
		writeError(c, http.StatusInternalServerError, "GeneralError", "could not read the drives")
		return mediaDef{}, proto.DriveInfo{}, false
	}
	for _, drive := range drives {
		if drive.ID == def.drive {
			return def, drive, true
		}
	}

	notFound(c)
	return mediaDef{}, proto.DriveInfo{}, false
}

// virtualMedia lists the drives the gadget has under base, the collection's
// path. With the virtual disk function off it has none, and the collection is
// empty.
func (s *Service) virtualMedia(base string) gin.HandlerFunc {
	return func(c *gin.Context) {
		drives, err := s.deps.ListDrives()
		if err != nil {
			log.Errorf("redfish: list drives: %s", err)
			writeError(c, http.StatusInternalServerError, "GeneralError", "could not read the drives")
			return
		}

		var members []string
		for _, def := range mediaDefs {
			for _, drive := range drives {
				if drive.ID == def.drive {
					members = append(members, base+"/"+def.id)
				}
			}
		}

		writeJSON(c, http.StatusOK, newCollection(base, "VirtualMediaCollection", "Virtual Media Services", members))
	}
}

// medium is one drive as a VirtualMedia resource under base. Its links and
// action targets stay under base, so a client that found it under the system
// never has to know the manager path.
func (s *Service) medium(base string) gin.HandlerFunc {
	return func(c *gin.Context) {
		def, drive, ok := s.driveFor(c)
		if !ok {
			return
		}

		path := base + "/" + def.id
		body := newResource(path, "VirtualMedia.v1_3_0.VirtualMedia")
		body["Id"] = def.id
		body["Name"] = def.name
		body["MediaTypes"] = def.types
		body["ConnectedVia"] = "Applet"
		body["Inserted"] = drive.File != ""
		body["WriteProtected"] = drive.Ro
		body["Image"] = nil
		body["ImageName"] = nil
		if drive.File != "" {
			body["Image"] = drive.File
			body["ImageName"] = filepath.Base(drive.File)
		}
		body["Actions"] = object{
			"#" + insertAction: object{"target": path + "/Actions/" + insertAction},
			"#" + ejectAction:  object{"target": path + "/Actions/" + ejectAction},
		}

		writeTagged(c, body)
	}
}

// insertMedia loads a local image. Image is a path under the image directory
// or a bare file name in it; a URL is refused, because the board does not
// download images.
func (s *Service) insertMedia(c *gin.Context) {
	def, _, ok := s.driveFor(c)
	if !ok {
		return
	}
	params, ok := decodeParams(c)
	if !ok {
		return
	}

	for name := range params {
		switch name {
		case "Image", "Inserted", "WriteProtected":
		case "TransferMethod", "TransferProtocolType":
			// A local file has neither, so they are accepted only when absent.
			if present(params, name) {
				writeError(c, http.StatusBadRequest, "ActionParameterNotSupported", "", name, insertAction)
				return
			}
		default:
			writeError(c, http.StatusBadRequest, "ActionParameterNotSupported", "", name, insertAction)
			return
		}
	}

	if !present(params, "Image") {
		writeError(c, http.StatusBadRequest, "ActionParameterMissing", "", insertAction, "Image")
		return
	}
	var image string
	if err := json.Unmarshal(params["Image"], &image); err != nil || image == "" {
		writeError(c, http.StatusBadRequest, "ActionParameterValueTypeError", "", string(params["Image"]), "Image", insertAction)
		return
	}

	inserted := true
	if present(params, "Inserted") {
		if err := json.Unmarshal(params["Inserted"], &inserted); err != nil {
			writeError(c, http.StatusBadRequest, "ActionParameterValueTypeError", "", string(params["Inserted"]), "Inserted", insertAction)
			return
		}
	}
	if !inserted {
		writeError(c, http.StatusBadRequest, "ActionParameterValueNotInList", "", "false", "Inserted", insertAction)
		return
	}

	writeProtected := true
	if present(params, "WriteProtected") {
		if err := json.Unmarshal(params["WriteProtected"], &writeProtected); err != nil {
			writeError(c, http.StatusBadRequest, "ActionParameterValueTypeError", "", string(params["WriteProtected"]), "WriteProtected", insertAction)
			return
		}
	}
	if def.drive == storage.DriveCdrom {
		writeProtected = true
	}

	if strings.Contains(image, "://") {
		writeError(c, http.StatusBadRequest, "ActionParameterValueNotInList",
			"only an image already on the board under "+s.deps.ImageDir+" can be inserted", image, "Image", insertAction)
		return
	}

	file := image
	if !filepath.IsAbs(file) {
		file = filepath.Join(s.deps.ImageDir, file)
	}
	file = filepath.Clean(file)

	info, err := os.Stat(file)
	if !utils.IsPathInside(s.deps.ImageDir, file) || err != nil || !info.Mode().IsRegular() {
		writeError(c, http.StatusBadRequest, "ResourceNotFound", "", "Image", image)
		return
	}

	if err := s.deps.InsertDrive(def.drive, file, writeProtected); err != nil {
		writeDriveError(c, insertAction, image, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ejectMedia empties the drive, with the forced eject where the kernel has it.
func (s *Service) ejectMedia(c *gin.Context) {
	def, _, ok := s.driveFor(c)
	if !ok {
		return
	}
	params, ok := decodeParams(c)
	if !ok {
		return
	}
	for name := range params {
		writeError(c, http.StatusBadRequest, "ActionParameterNotSupported", "", name, ejectAction)
		return
	}

	if err := s.deps.EjectDrive(def.drive); err != nil {
		writeDriveError(c, ejectAction, "", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// writeDriveError maps the storage layer's errors to Redfish messages.
func writeDriveError(c *gin.Context, action string, image string, err error) {
	switch {
	case errors.Is(err, storage.ErrInvalidImage):
		writeError(c, http.StatusBadRequest, "ResourceNotFound", "", "Image", image)
	case errors.Is(err, storage.ErrInOtherDrive):
		writeError(c, http.StatusConflict, "ResourceInUse", err.Error())
	case errors.Is(err, storage.ErrMediumLocked):
		writeError(c, http.StatusBadRequest, "ActionNotSupported", err.Error(), action)
	case errors.Is(err, storage.ErrNoDrive):
		notFound(c)
	default:
		log.Errorf("redfish: %s: %s", action, err)
		writeError(c, http.StatusInternalServerError, "GeneralError", err.Error())
	}
}
