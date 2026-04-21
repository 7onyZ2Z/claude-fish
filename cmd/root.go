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
	Use:   "claude-fish <novel-file>",
	Short: "A terminal novel reader disguised as a CLI coding tool",
	Args:  cobra.ExactArgs(1),
	Run:   run,
}

func init() {
	rootCmd.Flags().StringVarP(&codeFile, "code", "c", "", "code file for boss mode")
	rootCmd.Flags().StringVarP(&themeName, "theme", "t", "claude", "theme: claude, codex, opencode")
	rootCmd.Flags().IntVar(&speed, "speed", 25, "streaming speed in ms/char")
}

func run(cmd *cobra.Command, args []string) {
	novelPath := args[0]

	r := newReader(novelPath)
	if r == nil {
		fmt.Fprintf(os.Stderr, "Unsupported file format: %s\n", filepath.Ext(novelPath))
		os.Exit(1)
	}

	if err := r.Load(novelPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", novelPath, err)
		os.Exit(1)
	}

	var code, cfName string
	if codeFile != "" {
		data, err := os.ReadFile(codeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading code file %s: %v\n", codeFile, err)
			os.Exit(1)
		}
		code = string(data)
		cfName = filepath.Base(codeFile)
	}

	th := theme.FindByName(themeName)

	p := internal.NewApp(r, th, code, cfName, speed, filepath.Base(novelPath))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newReader(path string) reader.Reader {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt":
		return &reader.TXTReader{}
	case ".md", ".markdown":
		return &reader.MarkdownReader{}
	case ".epub":
		return &reader.EPUBReader{}
	default:
		return nil
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
