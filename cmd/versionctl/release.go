package main

import (
	"fmt"
	"os"

	"github.com/nanoteck137/versionctl"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use: "release",
	Run: func(cmd *cobra.Command, args []string) {
		// NOTE(patrik): versionctl release [patch|minor|major] [--dry-run] [--label <label>] [--pre-cmd \"cmd\"]

		err := versionctl.EnsureRepoRootOrChdir()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		part := "patch" // "minor" "major"

		dryRun := false
		label := ""
		preCmd := ""

		err = versionctl.Release(part, dryRun, label, preCmd)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
