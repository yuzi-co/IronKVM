package proto

import "time"

// The file browser's wire types. An image is named by its path under /data,
// as the drive API names it, and a file by its path inside the image.

type ImageFile struct {
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	IsDir    bool      `json:"isDir"`
	Modified time.Time `json:"modifiedAt"`
}

type ListFilesRsp struct {
	File string `json:"file"`
	Path string `json:"path"`
	// FS is the filesystem the image holds: fat12, fat16, fat32, exfat,
	// iso9660 or udf.
	FS string `json:"fs"`
	// Writable says whether files can be written into the image now: its
	// filesystem can be written and no drive serves it.
	Writable bool `json:"writable"`
	// Drive is the drive serving the image, "" when none does.
	Drive   string      `json:"drive"`
	Entries []ImageFile `json:"entries"`
	// Truncated says the directory holds more entries than a listing returns.
	Truncated bool   `json:"truncated"`
	Free      uint64 `json:"free"`
	Total     uint64 `json:"total"`
}

type ImageFileReq struct {
	File string `json:"file" form:"file" validate:"required"`
	Path string `json:"path" form:"path" validate:"required"`
}

type FileJobRsp struct {
	// Status is idle, in_progress, success or failed.
	Status     string `json:"status"`
	Name       string `json:"name"`
	Percentage string `json:"percentage"`
	Loaded     int64  `json:"loaded"` // bytes written so far
	Total      int64  `json:"total"`  // expected bytes, 0 when unknown
	Message    string `json:"message"`
	Queued     int    `json:"queued"` // uploads waiting their turn
}
