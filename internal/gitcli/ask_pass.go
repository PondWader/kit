package gitcli

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func (c *Client) spawnAskPassServer(errHandler func(e error)) (askPassServer, error) {
	if runtime.GOOS == "windows" {
		return askPassServer{}, errors.New("askpass is currently not supported on Windows")
	}

	sockPath := filepath.Join(os.TempDir(), fmt.Sprintf("kit-git-askpass-%d-%d.sock", os.Getpid(), rand.Int64()))

	askpassScript := fmt.Sprintf(`#!/bin/sh
echo "$1" | nc -U %s
`, sockPath)
	f, err := os.CreateTemp("", "askpass_*.sh")
	if err != nil {
		return askPassServer{}, err
	}
	defer f.Close()
	if err = f.Chmod(0700); err != nil {
		return askPassServer{}, err
	}
	if _, err = f.Write([]byte(askpassScript)); err != nil {
		return askPassServer{}, err
	}

	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		return askPassServer{}, err
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			s := bufio.NewScanner(conn)
			if !s.Scan() {
				conn.Close()
				return
			}

			prompt := s.Text()
			resp, err := c.Prompt(prompt, isSecret(prompt))
			if err != nil {
				errHandler(err)
			} else {
				conn.Write([]byte(resp + "\n"))
			}
			conn.Close()
		}
	}()

	return askPassServer{f.Name(), sockPath, ln}, nil
}

type askPassServer struct {
	Path     string
	sockPath string
	ln       net.Listener
}

func (s askPassServer) Close() error {
	s.ln.Close()
	os.Remove(s.sockPath)
	return os.Remove(s.Path)
}

func isSecret(prompt string) bool {
	prompt = strings.ToLower(prompt)
	return strings.Contains(prompt, "password") || strings.Contains(prompt, "passphrase")
}
