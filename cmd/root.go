package cmd

import (
	"fmt"
	"os"

	"claude-fish/internal"
	"claude-fish/internal/theme"

	"github.com/spf13/cobra"
)

var (
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
	rootCmd.Flags().StringVarP(&themeName, "theme", "t", "claude", "theme: claude, codex, opencode")
	rootCmd.Flags().IntVar(&speed, "speed", 25, "streaming speed in ms/char")
}

func run(cmd *cobra.Command, args []string) {
	th := theme.FindByName(themeName)

	p := internal.NewApp(nil, th, nil, "", speed)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
