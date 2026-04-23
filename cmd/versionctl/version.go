package main

import (
	"fmt"
	"os"

	"github.com/nanoteck137/versionctl"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		if err := versionctl.EnsureRepoRootOrChdir(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		v, err := versionctl.ResolveVersion()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		fmt.Println(v)
	},
}


func init() {
	rootCmd.AddCommand(versionCmd)
}
