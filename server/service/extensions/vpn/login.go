package vpn

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

// waitDelay bounds how long Wait tolerates a still-open pipe after the process
// itself has gone.
const waitDelay = 5 * time.Second

type loginResult struct {
	url string
	err error
}

// LoginURL starts a login command and returns the URL the user has to visit.
// Tailscale prints it on stderr, NetBird on stdout.
//
// The command keeps running afterwards, until the login is completed in the
// browser, so its output has to keep draining: closing the pipe early hands it
// a SIGPIPE on its next line and kills the login. It also has to be reaped
// rather than left as an orphan, once, on every path out of here.
//
// timeout bounds the wait for the URL. life, when positive, bounds how long
// the command may run after it: netbird up has no timeout of its own and would
// otherwise wait for the browser forever.
func LoginURL(cmd *exec.Cmd, stdout bool, timeout, life time.Duration) (string, error) {
	var pipe io.ReadCloser
	var err error
	if stdout {
		pipe, err = cmd.StdoutPipe()
	} else {
		pipe, err = cmd.StderrPipe()
	}
	if err != nil {
		return "", err
	}

	// Safety net if the command ever leaves a child holding the pipe open.
	cmd.WaitDelay = waitDelay

	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Buffered, so this goroutine still finishes if nobody is listening.
	results := make(chan loginResult, 1)

	go func() {
		url, err := ReadLoginURL(pipe)
		results <- loginResult{url: url, err: err}

		// Wait closes the pipe itself, so it must not be closed here.
		_, _ = io.Copy(io.Discard, pipe)
		_ = cmd.Wait()
	}()

	select {
	case result := <-results:
		if result.err == nil && life > 0 {
			// Kill after Wait is harmless: it reports the process done.
			time.AfterFunc(life, func() { _ = cmd.Process.Kill() })
		}
		return result.url, result.err

	case <-time.After(timeout):
		// Otherwise the handler blocks for the command's whole life.
		_ = cmd.Process.Kill()
		return "", fmt.Errorf("timed out waiting for the login url")
	}
}

// ReadLoginURL reads lines until one holds an https URL and returns that URL.
// Without one, the error carries the last lines it read, which say why.
func ReadLoginURL(reader io.Reader) (string, error) {
	buffered := bufio.NewReader(reader)
	var seen bytes.Buffer
	for {
		line, err := buffered.ReadString('\n')
		for _, field := range strings.Fields(line) {
			if strings.HasPrefix(field, "https://") {
				return field, nil
			}
		}
		seen.WriteString(line)
		if err != nil {
			return "", &CmdError{Err: errors.New("no login url in the output"), Tail: Tail(seen.Bytes(), TailLines)}
		}
	}
}
