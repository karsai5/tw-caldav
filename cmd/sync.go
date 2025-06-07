/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"karsai5/tw-caldav/internal/sync"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		syncProcess, err := sync.NewSyncProcess()
		if err != nil {
			panic(err)
		}

		if err = syncProcess.Sync(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
