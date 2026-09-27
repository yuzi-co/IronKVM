package proto

type GetImagesRsp struct {
	Files []string `json:"files"`
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
}

type GetDrivesRsp struct {
	Drives []DriveInfo `json:"drives"`
}

type InsertDriveReq struct {
	File string `json:"file" validate:"required"`
	Ro   bool   `json:"ro" validate:"omitempty"`
}
