package proto

type GetImagesRsp struct {
	Files []string `json:"files"`
	// Sizes maps each file in Files to its size in bytes, so the UI can warn
	// about an image that is empty or too big for the CD drive.
	Sizes map[string]int64 `json:"sizes"`
}

type MountImageReq struct {
	File  string `json:"file" validate:"omitempty"`
	Cdrom bool   `json:"cdrom" validate:"omitempty"`
}

type GetMountedImageRsp struct {
	File string `json:"file"`
}

type GetCdRomRsp struct {
	Cdrom int64 `json:"cdrom"`
}

type DeleteImageReq struct {
	File string `json:"file" validate:"required"`
}

type DriveInfo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	File string `json:"file"`
	Ro   bool   `json:"ro"`
	// Size is the size in bytes of the file the drive serves, 0 when it is
	// empty or serves something other than a regular file (a device).
	Size int64 `json:"size"`
	// Missing says the drive serves a file whose path no longer exists: it
	// was deleted or renamed while in the drive.
	Missing bool `json:"missing"`
}

type GetDrivesRsp struct {
	Drives []DriveInfo `json:"drives"`
}

type InsertDriveReq struct {
	File string `json:"file" validate:"required"`
	Ro   bool   `json:"ro" validate:"omitempty"`
}
