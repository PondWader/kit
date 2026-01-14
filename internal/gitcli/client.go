package gitcli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type PromptHandler func(prompt string, secret bool) (resp string, err error)

type Client struct {
	Prompt PromptHandler
}

// func (c Client) Clone(url string, dest string) error {
// 	return runCmd(c.Prompt, "", "clone", url, dest)
// }

type Credential struct {
	Username string
	Password string
}

func (c Client) GetCredentials(url string) (cred Credential, err error) {
	lines, err := c.runCmd(c.Prompt, fmt.Sprintf("url=%s\n\n", url), "credential", "fill")
	if err != nil {
		return cred, err
	}

	for _, line := range lines {
		if after, ok := strings.CutPrefix(line, "username="); ok {
			cred.Username = after
		} else if after, ok := strings.CutPrefix(line, "password="); ok {
			cred.Password = after
		}
	}

	return cred, err
}

func (c Client) runCmd(prompt PromptHandler, in string, args ...string) ([]string, error) {
	var inputErr error
	var stdin io.WriteCloser
	var cmd *exec.Cmd
	var err error
	askPass, askPassErr := c.spawnAskPassServer(func(e error) {
		if e != nil {
			inputErr = e
			if cmd != nil {
				cmd.Process.Kill()
			}
		}
	})
	if askPassErr == nil {
		defer askPass.Close()
	}

	cmd = exec.Command("git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if askPassErr == nil {
		cmd.Env = append(cmd.Env, "GIT_ASKPASS="+askPass.Path)
	}

	stderr := bytes.NewBuffer(nil)
	stdout := outputHandler{prompt: prompt}

	cmd.Stdout = &stdout
	cmd.Stderr = stderr
	stdin, err = cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	stdin.Write([]byte(in))

	err = cmd.Wait()
	if stderr.Len() != 0 {
		err = errors.New(stderr.String())
	}
	if inputErr != nil {
		err = inputErr
	}

	return stdout.lines, err
}

type outputHandler struct {
	prompt PromptHandler
	line   bytes.Buffer
	lines  []string
}

func (o *outputHandler) Write(p []byte) (int, error) {
	for _, b := range p {
		if b == '\n' {
			o.lines = append(o.lines, o.line.String())
			o.line.Reset()
		} else {
			o.line.WriteByte(b)
		}
	}

	return len(p), nil
}
