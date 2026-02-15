package main

import (
	"fmt"

	"github.com/scottbrown/setlist"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Printing the current version",
	Long:  "Printing the current version and commit information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s (%s)\n", AppName, setlist.VERSION, setlist.COMMIT)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
