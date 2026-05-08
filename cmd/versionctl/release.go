package main

import (
	"fmt"
	"os"

	"github.com/nanoteck137/versionctl/app"
	"github.com/nanoteck137/versionctl/config"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use: "release",
	Run: func(cmd *cobra.Command, args []string) {
		// NOTE(patrik): versionctl release [patch|minor|major] [--dry-run] [--label <label>] [--pre-cmd \"cmd\"]

		err := app.EnsureRepoRootOrChdir()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		conf, err := config.Load()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		_ = conf

		label := ""

		err = app.Release(conf, label)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
