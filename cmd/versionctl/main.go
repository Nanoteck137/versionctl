package main

import (
	"os"

	"github.com/nanoteck137/versionctl"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     versionctl.AppName,
	Version: versionctl.Version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func init() {
	rootCmd.SetVersionTemplate(versionctl.VersionTemplate(versionctl.AppName))
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
