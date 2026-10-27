package egress

import (
	"NanoKVM-Server/common"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"
)

// MaxClients is how many players RTSP serves at once, and how many callers
// SRT does (ironkvm-dist#142, #143). Each one beside the browser costs the
// single core its packets and, for SRT, its own MPEG-TS mux; two is what the
// issues ask for and what leaves the browser's picture alone.
const MaxClients = 2

// Viewer is who a player authenticated as, and how to ask whether that still
// holds while it watches.
type Viewer struct {
	Username string
	Role     string
	// Via and Key are what the audit log records: how the player got in,
	// and the name of the API key when it used one.
	Via string
	Key string
	// Valid reports whether the credential is still good. Nil means it does
	// not lapse.
	Valid func() bool
}

// Authenticator checks a token that may only watch: a view-only API key or a
// share link token, as WHEP takes (middleware.AuthenticateViewToken). The
// token is empty when the player sent none, which only a board with
// authentication switched off accepts.
type Authenticator func(token string) (Viewer, bool)

// ParameterSets returns the parameter sets in an access unit: SPS and PPS for
// H.264, and the VPS as well for H.265. Any of them is nil when the unit does
// not carry it.
func ParameterSets(codec uint8, nalus [][]byte) (vps, sps, pps []byte) {
	for _, nalu := range nalus {
		if len(nalu) == 0 {
			continue
		}
		if codec == common.CodecH265 {
			switch h265.NALUType((nalu[0] >> 1) & 0b111111) {
			case h265.NALUType_VPS_NUT:
				vps = nalu
			case h265.NALUType_SPS_NUT:
				sps = nalu
			case h265.NALUType_PPS_NUT:
				pps = nalu
			}
			continue
		}
		switch h264.NALUType(nalu[0] & 0x1f) {
		case h264.NALUTypeSPS:
			sps = nalu
		case h264.NALUTypePPS:
			pps = nalu
		}
	}
	return vps, sps, pps
}
