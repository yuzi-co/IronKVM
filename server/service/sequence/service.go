package sequence

import (
	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

type Service struct {
	manager *Manager
}

func NewService() *Service {
	return &Service{manager: GetManager()}
}

// StartAutoTap starts pressing a key over and over, and answers at once. The
// status request reports the run, and the stop request ends it.
func (s *Service) StartAutoTap(c *gin.Context) {
	var req proto.AutoTapReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	p, err := autoTapProgram(req)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	s.start(c, p)
}

// PlayMacro plays a saved macro, or the steps in the request, and answers at
// once.
func (s *Service) PlayMacro(c *gin.Context) {
	var req proto.PlayMacroReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	name, steps := truncate(req.Name, maxNameLen), req.Steps
	if req.ID != "" {
		macro, err := getMacro(req.ID)
		if err != nil {
			log.Errorf("failed to read macro %s: %v", req.ID, err)
			rsp.ErrRsp(c, -2, err.Error())
			return
		}
		name, steps = macro.Name, macro.Steps
	}

	compiled, err := compileSteps(steps)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	s.start(c, program{kind: KindMacro, name: name, steps: compiled})
}

func (s *Service) start(c *gin.Context, p program) {
	var rsp proto.Response

	status, err := s.manager.start(p)
	if err != nil {
		rsp.Data = status
		rsp.ErrRsp(c, -3, err.Error())
		return
	}
	rsp.OkRspWithData(c, status)
	log.Debugf("%s sequence %q started", p.kind, p.name)
}

// GetStatus reports the sequence running now, or the last one.
func (s *Service) GetStatus(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.manager.Status())
}

// Stop ends the running sequence. A run that has already ended is not an
// error: the button that asks for it races the run's own end.
func (s *Service) Stop(c *gin.Context) {
	var rsp proto.Response

	status, err := s.manager.Stop()
	if err != nil && !errors.Is(err, errNotRunning) {
		rsp.Data = status
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	rsp.OkRspWithData(c, status)
}

func (s *Service) GetMacros(c *gin.Context) {
	var rsp proto.Response

	macros, err := listMacros()
	if err != nil {
		log.Errorf("failed to read macros: %v", err)
		rsp.ErrRsp(c, -1, "get macros failed")
		return
	}
	rsp.OkRspWithData(c, &proto.GetMacrosRsp{Macros: macros})
}

func (s *Service) SaveMacro(c *gin.Context) {
	var req proto.SaveMacroReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	macro, err := putMacro(proto.Macro{ID: req.ID, Name: req.Name, Steps: req.Steps})
	if err != nil {
		log.Errorf("failed to save macro: %v", err)
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	rsp.OkRspWithData(c, macro)
	log.Debugf("saved macro %s, %d steps", macro.ID, len(macro.Steps))
}

func (s *Service) DeleteMacro(c *gin.Context) {
	var req proto.DeleteMacroReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if err := deleteMacro(req.ID); err != nil {
		log.Errorf("failed to delete macro %s: %v", req.ID, err)
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	rsp.OkRsp(c)
	log.Debugf("deleted macro %s", req.ID)
}
