/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"karsai5/tw-caldav/internal/sync"
	"strings"

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
		// Create new remote tasks locally
		// for _, t := range processedTasks.newLocalTasks {
		// 	if t.LocalId() != nil {
		// 		continue
		// 	}
		//
		// 	slog.Info("Creating remote task locally", "desc", t.Description())
		// 	_, err := local.CreateTask(t, synctime)
		// 	if err != nil {
		// 		slog.Error("Could not create local task", "err", err)
		// 	}
		//
		// 	// TODO: Update existing task
		// 	// t.SetLocalId(&uuid)
		// 	// t.SetLastSynced(&synctime)
		//
		// 	// _, err = remote.Client.PutCalendarObject(context.TODO(), t.Path, t.CalendarObject.Data)
		// 	// if err != nil {
		// 	// 	slog.Error("Could not update remote task", "err", err)
		// 	// }
		// }
	},
}

func escapeQuotes(str string) string {
	return strings.ReplaceAll(str, `"`, `\"`)
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

