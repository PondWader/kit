package gitcli

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func (c *Client) spawnAskPassServer() (askPassServer, error) {
	sockPath := filepath.Join(os.TempDir(), fmt.Sprintf("git-askpass-%d.sock", os.Getpid()))

	askpassScript := fmt.Sprintf(`#!/bin/sh
echo "$1" | nc -U %s
nc -U %s
`, sockPath, sockPath)
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
			defer conn.Close()

			s := bufio.NewScanner(conn)
			if !s.Scan() {
				return
			}

			c.Prompt(s.Text(), false)
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
