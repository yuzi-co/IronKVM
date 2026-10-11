package allowlist

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
)

// The web UI's Allowed networks setting, under /api/network/allowlist.

// State is what the setting page shows.
type State struct {
	// Networks is the list in force, empty when every address is allowed.
	Networks []string `json:"networks"`
	// Client is the address this request came from, so the page can say
	// which entry keeps the owner in.
	Client string `json:"client"`
}

type setRequest struct {
	Networks []string `json:"networks"`
}

// saveMu serialises saves, which read and rewrite server.yaml.
var saveMu sync.Mutex

// save writes the list to server.yaml. Tests replace it.
var save = func(networks []string) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}
	conf.Security.AllowedNetworks = networks
	return config.Write(conf)
}

// Service holds the settings handlers.
type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetSettings(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, state(c))
}

func (s *Service) SetSettings(c *gin.Context) {
	var req setRequest
	var rsp proto.Response

	if err := c.ShouldBindJSON(&req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	list, err := Parse(req.Networks)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}

	// The lockout guard. A list that refuses the address saving it would end
	// this session on the next request, and the owner might have no other
	// way in, so it is refused here instead.
	if client, ok := remoteAddr(c.Request); ok && !permits(list, client) {
		rsp.ErrRsp(c, -2, fmt.Sprintf(
			"the list does not include your address %s; saving it would lock you out", client))
		return
	}

	networks := format(list)

	saveMu.Lock()
	defer saveMu.Unlock()

	if err := save(networks); err != nil {
		log.Errorf("allowlist: failed to save: %s", err)
		rsp.ErrRsp(c, -3, "failed to save the setting")
		return
	}
	config.GetInstance().Security.AllowedNetworks = networks
	set(list)

	if len(networks) == 0 {
		log.Infof("allowlist: every address allowed")
	} else {
		log.Infof("allowlist: allowed networks set to %v", networks)
	}
	rsp.OkRspWithData(c, state(c))
}

func state(c *gin.Context) State {
	networks := format(prefixes())
	client := ""
	if addr, ok := remoteAddr(c.Request); ok {
		client = addr.String()
	}
	return State{Networks: networks, Client: client}
}
