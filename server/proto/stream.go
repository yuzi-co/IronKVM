package proto

type UpdateFrameDetectReq struct {
	Enabled bool `validate:"omitempty"`
}

type GetFrameDetectRsp struct {
	Enabled bool `json:"enabled"`
}

type StopFrameDetectReq struct {
	Duration int `validate:"omitempty"`
}
