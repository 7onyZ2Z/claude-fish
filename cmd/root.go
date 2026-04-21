package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"claude-fish/internal"
	"claude-fish/internal/reader"
	"claude-fish/internal/theme"

	"github.com/spf13/cobra"
)

var (
	codeFile  string
	themeName string
	speed     int
)

var rootCmd = &cobra.Command{
	Use:   "claude-fish",
	Short: "A terminal novel reader disguised as a CLI coding tool",
	Args:  cobra.NoArgs,
	Run:   run,
}

func init() {
	rootCmd.Flags().StringVarP(&codeFile, "code", "c", "", "code file for boss mode (default: main.go)")
	rootCmd.Flags().StringVarP(&themeName, "theme", "t", "claude", "theme: claude, codex, opencode")
	rootCmd.Flags().IntVar(&speed, "speed", 25, "streaming speed in ms/char")
}

func run(cmd *cobra.Command, args []string) {
	th := theme.FindByName(themeName)

	// Load code file for boss mode, default to main.go
	var code, cfName string
	cf := codeFile
	if cf == "" {
		cf = "main.go"
	}
	if data, err := os.ReadFile(cf); err == nil {
		code = string(data)
		cfName = filepath.Base(cf)
	}

	p := internal.NewApp(nil, th, code, cfName, speed, "")
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newReaderForPath(path string) reader.Reader {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt":
		return &reader.TXTReader{}
	case ".md", ".markdown":
		return &reader.MarkdownReader{}
	case ".epub":
		return &reader.EPUBReader{}
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
