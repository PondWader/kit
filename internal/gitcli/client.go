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
	stdout := lineWriter{}

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

type lineWriter struct {
	line  bytes.Buffer
	lines []string
}

func (o *lineWriter) Write(p []byte) (int, error) {
	var i int
	for j, b := range p {
		if b == '\n' {
			o.lines = append(o.lines, o.line.String())
			o.line.Reset()
			i = j + 1
		} else {
			o.line.Write(p[i:j])
		}
	}
	if i < len(p) {
		o.line.Write(p[i:])
	}

	return len(p), nil
}
