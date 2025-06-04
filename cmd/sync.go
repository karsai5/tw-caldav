/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"karsai5/tw-caldav/internal/caldav"
	"karsai5/tw-caldav/internal/task"
	"karsai5/tw-caldav/internal/tw"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		remote, err := caldav.NewClient(viper.GetString("url"), viper.GetString("user"), viper.GetString("pass"))
		if err != nil {
			panic(err)
		}
		remoteTasks, err := remote.GetAllTodos()
		if err != nil {
			panic(err)
		}

		local := tw.Taskwarrior{}
		localTasks, err := local.GetAllTasks()
		if err != nil {
			panic(err)
		}

		fmt.Println("REMOTE TASKS")
		task.PrintTable(getArrayFromRemoteTasks(remoteTasks))

		fmt.Println("LOCAL TASKS")
		task.PrintTable(getArrayFromLocalTasks(localTasks))

		for _, t := range remoteTasks {
			if t.LocalId() != nil {
				continue
			}
			slog.Info("Creating remote task locally", "desc", t.Description())
		}

	},
}

func getArrayFromLocalTasks(todos []*tw.Task) []task.Task {
	arr := make([]task.Task, len(todos))
	for i, t := range todos {
		arr[i] = t
	}
	return arr
}

func getArrayFromRemoteTasks(todos []*caldav.Todo) []task.Task {
	arr := make([]task.Task, len(todos))
	for i, t := range todos {
		arr[i] = t
	}
	return arr
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
