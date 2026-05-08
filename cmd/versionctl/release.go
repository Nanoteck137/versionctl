package main

import (
	"fmt"
	"os"

	"github.com/kr/pretty"
	"github.com/nanoteck137/versionctl/app"
	"github.com/nanoteck137/versionctl/config"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use: "release",
	Run: func(cmd *cobra.Command, args []string) {
		version, _ := cmd.Flags().GetString("version")
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

		pretty.Println(conf)

		label := ""

		err = app.Release(conf, version, label)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	releaseCmd.Flags().String("version", "", "force set version")

	rootCmd.AddCommand(releaseCmd)
}
