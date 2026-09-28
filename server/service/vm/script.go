package vm

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
)

// ScriptDirectory is a variable so tests can point it somewhere else.
var ScriptDirectory = "/etc/kvm/scripts"

// foregroundTimeout bounds a foreground run. The browser stops waiting after
// ten minutes; a script still going then has no one left to report to, and
// would hold its request and its processes for as long as it liked. It is a
// variable so tests can shorten it.
var foregroundTimeout = 10 * time.Minute

// errScriptTimedOut is the error of a foreground run that was killed for
// running past foregroundTimeout.
var errScriptTimedOut = errors.New("timed out")

func (s *Service) GetScripts(c *gin.Context) {
	var rsp proto.Response

	files := []string{}
	err := filepath.Walk(ScriptDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isScript(info.Name()) {
			files = append(files, info.Name())
		}

		return nil
	})
	// Upload creates the directory, so a board that never had a script has
	// none. That is an empty list, not an error.
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		rsp.ErrRsp(c, -1, "get scripts failed")
		return
	}

	rsp.OkRspWithData(c, &proto.GetScriptsRsp{
		Files: files,
	})

	log.Debugf("get scripts total %d", len(files))
}

func (s *Service) UploadScript(c *gin.Context) {
	var rsp proto.Response

	_, header, err := c.Request.FormFile("file")
	if err != nil {
		rsp.ErrRsp(c, -1, "bad request")
		return
	}

	target, err := utils.SecureJoin(ScriptDirectory, header.Filename)
	if err != nil || !isScript(header.Filename) {
		rsp.ErrRsp(c, -2, "invalid arguments")
		return
	}

	if _, err = os.Stat(ScriptDirectory); err != nil {
		_ = os.MkdirAll(ScriptDirectory, 0o755)
	}

	err = c.SaveUploadedFile(header, target)
	if err != nil {
		rsp.ErrRsp(c, -2, "save failed")
		return
	}

	_ = utils.EnsurePermission(target, 0o100)

	data := &proto.UploadScriptRsp{
		File: header.Filename,
	}
	rsp.OkRspWithData(c, data)

	log.Debugf("upload script %s success", header.Filename)
}

func (s *Service) RunScript(c *gin.Context) {
	var req proto.RunScriptReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	// The name reaches a process argument, so it must be a plain file inside
	// the script directory - never a path, and never shell metacharacters.
	script, err := utils.SecureJoin(ScriptDirectory, req.Name)
	if err != nil || !isScript(req.Name) {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	var output []byte
	cmd := scriptCommand(req.Name, script)

	if req.Type == "foreground" {
		output, err = runForeground(cmd, foregroundTimeout)
	} else {
		cmd.Stdout = nil
		cmd.Stderr = nil
		go func() {
			err := cmd.Run()
			if err != nil {
				log.Errorf("run script %s in background failed: %s", req.Name, err)
			}
		}()
	}

	if err != nil {
		log.Errorf("run script %s failed: %s", req.Name, err.Error())
		// The output of a script that failed is what says why, so it goes
		// back with the error.
		rsp.Err(-2, fmt.Sprintf("run script failed: %s", err))
		rsp.Data = &proto.RunScriptRsp{Log: string(output)}
		c.JSON(http.StatusOK, &rsp)
		return
	}

	rsp.OkRspWithData(c, &proto.RunScriptRsp{
		Log: string(output),
	})

	log.Debugf("run script %s success", req.Name)
}

func (s *Service) DeleteScript(c *gin.Context) {
	var req proto.DeleteScriptReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	file, err := utils.SecureJoin(ScriptDirectory, req.Name)
	if err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if err := os.Remove(file); err != nil {
		log.Errorf("delete script %s failed: %s", file, err)
		rsp.ErrRsp(c, -3, "delete failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("delete script %s success", file)
}

// scriptCommand builds the command that runs a script. The interpreter takes
// the path as an argument, so the name can never become shell text.
//
// A shell script goes through sh rather than being executed directly. Upload
// sets the execute bit, but the kernel still needs a shebang to know what
// interprets the file, and a .sh written as a plain list of commands has none.
// Executing the path itself fails those with ENOEXEC; naming the interpreter
// does not.
func scriptCommand(name string, path string) *exec.Cmd {
	if strings.HasSuffix(strings.ToLower(name), ".py") {
		return exec.Command("python", path)
	}

	return exec.Command("sh", path)
}

// runForeground runs cmd and returns its combined output, killing it once it
// has run for longer than timeout. The script runs in a process group of its
// own, and the whole group is killed, so the commands a shell script started
// go with it instead of living on as orphans.
func runForeground(cmd *exec.Cmd, timeout time.Duration) ([]byte, error) {
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// A process that left the group can still hold the output pipe open.
	// Wait gives up on it this long after the kill instead of hanging.
	cmd.WaitDelay = 5 * time.Second

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-done:
		return output.Bytes(), err
	case <-timer.C:
		// The negative pid names the process group.
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
			log.Errorf("kill script process group %d: %s", cmd.Process.Pid, err)
		}
		<-done
		return output.Bytes(), errScriptTimedOut
	}
}

func isScript(name string) bool {
	nameLower := strings.ToLower(name)
	if strings.HasSuffix(nameLower, ".sh") || strings.HasSuffix(nameLower, ".py") {
		return true
	}

	return false
}
