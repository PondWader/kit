package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PondWader/kit/internal/ansi"
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

		result, err := setupBashrc()
		if err != nil {
			printError(err)
			os.Exit(1)
		}

		fmt.Printf(ansi.Cyan("Updated ~/.bashrc %s\n"), ansi.BrightBlack(fmt.Sprintf("(%d added, %d unchanged)", len(result.Added), len(result.Unchanged))))
		for _, line := range result.Added {
			fmt.Printf("%s %s\n", ansi.Green("+"), line)
		}
		for _, line := range result.Unchanged {
			fmt.Printf("%s %s %s\n", ansi.Yellow("~"), line, ansi.BrightBlack("(already present)"))
		}
		fmt.Printf("%s\n", ansi.BrightBlack("Run `source ~/.bashrc` to apply changes"))
	},
}

func setupBashrc() (bashrcUpdateResult, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return bashrcUpdateResult{}, err
	}

	bashrcPath := filepath.Join(homeDir, ".bashrc")

	kitHome, err := kit.ResolveHome()
	if err != nil {
		return bashrcUpdateResult{}, err
	}

	binPath := makeTransferablePath(filepath.Join(kitHome, "bin"), homeDir)
	libPath := makeTransferablePath(filepath.Join(kitHome, "lib"), homeDir)
	pathLine := fmt.Sprintf("export PATH=\"%s:$PATH\"", binPath)
	ldLibraryPathLine := fmt.Sprintf("export LD_LIBRARY_PATH=\"%s:$LD_LIBRARY_PATH\"", libPath)

	contents, err := os.ReadFile(bashrcPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return bashrcUpdateResult{}, err
	}

	result := prependMissingBashrcLines(string(contents), []bashrcLine{
		{Comment: bashrcPathNote, Line: pathLine},
		{Comment: bashrcLibNote, Line: ldLibraryPathLine},
	})
	if len(result.Added) > 0 {
		if err := os.WriteFile(bashrcPath, []byte(result.Content), 0644); err != nil {
			return bashrcUpdateResult{}, err
		}
	}

	return result, nil
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

type bashrcUpdateResult struct {
	Content   string
	Added     []string
	Unchanged []string
}

func prependMissingBashrcLines(content string, lines []bashrcLine) bashrcUpdateResult {
	missing := make([]string, 0, len(lines)*2)
	addedLines := make([]string, 0, len(lines))
	unchangedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(content, line.Line) {
			unchangedLines = append(unchangedLines, line.Line)
			continue
		}
		missing = append(missing, line.Comment, line.Line)
		addedLines = append(addedLines, line.Line)
	}

	if len(missing) == 0 {
		return bashrcUpdateResult{Content: content, Added: addedLines, Unchanged: unchangedLines}
	}

	prefix := strings.Join(missing, "\n")
	if strings.TrimSpace(content) == "" {
		return bashrcUpdateResult{Content: prefix + "\n", Added: addedLines, Unchanged: unchangedLines}
	}

	return bashrcUpdateResult{Content: prefix + "\n\n" + content, Added: addedLines, Unchanged: unchangedLines}
}
