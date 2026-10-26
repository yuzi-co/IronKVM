package hid

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// HostLayoutFile holds the ID of the keyboard layout active on the host, set
// by the owner. The MCP server and PicoClaw type with it when a call names no
// layout, and the paste dialog starts on it. Without the file they type US.
// It is a variable so that tests can move it.
var HostLayoutFile = "/etc/kvm/keyboard-layout"

const defaultHostLayout = "us"

// HostLayoutID returns the ID of the host's layout, and whether the owner set
// one. A file naming no known layout counts as unset.
func HostLayoutID() (string, bool) {
	data, err := os.ReadFile(HostLayoutFile)
	if err != nil {
		return defaultHostLayout, false
	}
	id := strings.TrimSpace(string(data))
	if id == "" {
		return defaultHostLayout, false
	}
	layout, err := GetLayout(id)
	if err != nil {
		log.Warnf("host keyboard layout: %v, typing US", err)
		return defaultHostLayout, false
	}
	return layout.ID, true
}

// TypingLayout returns the layout to type a text with: the one named, or the
// host's layout when id is empty. An unknown name is an error that lists the
// IDs there are, for an agent to pick from.
func TypingLayout(id string) (*Layout, error) {
	if strings.TrimSpace(id) == "" {
		id, _ = HostLayoutID()
	}
	layout, err := GetLayout(id)
	if err != nil {
		return nil, fmt.Errorf("unknown keyboard layout %q; the layouts are %s", id, strings.Join(LayoutIDs(), ", "))
	}
	return layout, nil
}

func (s *Service) GetHostLayout(c *gin.Context) {
	var rsp proto.Response

	id, configured := HostLayoutID()
	rsp.OkRspWithData(c, &proto.GetHostLayoutRsp{
		Layout:     id,
		Configured: configured,
	})
}

// SetHostLayout stores the host's layout under its ID. An empty layout
// removes the file, which leaves US.
func (s *Service) SetHostLayout(c *gin.Context) {
	var req proto.SetHostLayoutReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if strings.TrimSpace(req.Layout) == "" {
		if err := os.Remove(HostLayoutFile); err != nil && !os.IsNotExist(err) {
			rsp.ErrRsp(c, -3, "reset failed")
			return
		}
		rsp.OkRsp(c)
		log.Debugf("host keyboard layout reset")
		return
	}

	layout, err := GetLayout(req.Layout)
	if err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	if err := os.WriteFile(HostLayoutFile, []byte(layout.ID), 0o644); err != nil {
		rsp.ErrRsp(c, -3, "write failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("set host keyboard layout: %s", layout.ID)
}
