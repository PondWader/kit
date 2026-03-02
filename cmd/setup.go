package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	kit "github.com/PondWader/kit/pkg"
)

const (
	setupBashrcArg = "bashrc"
	bashrcPathNote = "# Added by `kit setup bashrc`: add Kit binaries to PATH"
	bashrcLibNote  = "# Added by `kit setup bashrc`: add Kit libraries to LD_LIBRARY_PATH"
)

var SetupCommand = Command{
	Name:             "setup",
	Usage:            "<bashrc>",
	Description:      "configure shell setup",
	RequiredArgCount: 1,
	Run: func(fs *flag.FlagSet) {
		target := fs.Arg(0)
		if target != setupBashrcArg {
			printError(errors.New("only \"bashrc\" is supported for setup right now"))
			os.Exit(1)
		}

		if err := setupBashrc(); err != nil {
			printError(err)
			os.Exit(1)
		}

		fmt.Println("Updated ~/.bashrc with kit environment variables")
	},
}

func setupBashrc() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")

	kitHome, err := kit.ResolveHome()
	if err != nil {
		return err
	}

	binPath := makeTransferablePath(filepath.Join(kitHome, "bin"), homeDir)
	libPath := makeTransferablePath(filepath.Join(kitHome, "lib"), homeDir)
	pathLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", binPath)
	ldLibraryPathLine := fmt.Sprintf("export LD_LIBRARY_PATH=\"%s:$LD_LIBRARY_PATH\"", libPath)

	contents, err := os.ReadFile(bashrcPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	updated := prependMissingBashrcLines(string(contents), []bashrcLine{
		{Comment: bashrcPathNote, Line: pathLine},
		{Comment: bashrcLibNote, Line: ldLibraryPathLine},
	})
	if err := os.WriteFile(bashrcPath, []byte(updated), 0644); err != nil {
		return err
	}

	return nil
}

func makeTransferablePath(path, homeDir string) string {
	homeDir = filepath.Clean(homeDir)
	path = filepath.Clean(path)

	homePrefix := homeDir + string(os.PathSeparator)
	if path == homeDir {
		return "$HOME"
	}
	if strings.HasPrefix(path, homePrefix) {
		return "$HOME/" + strings.TrimPrefix(path, homePrefix)
	}

	return filepath.ToSlash(path)
}

type bashrcLine struct {
	Comment string
	Line    string
}

func prependMissingBashrcLines(content string, lines []bashrcLine) string {
	missing := make([]string, 0, len(lines)*2)
	for _, line := range lines {
		if strings.Contains(content, line.Line) {
			continue
		}
		missing = append(missing, line.Comment, line.Line)
	}

	if len(missing) == 0 {
		return content
	}

	prefix := strings.Join(missing, "\n")
	if strings.TrimSpace(content) == "" {
		return prefix + "\n"
	}

	return prefix + "\n\n" + content
}
