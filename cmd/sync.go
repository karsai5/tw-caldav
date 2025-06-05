/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"karsai5/tw-caldav/internal/caldav"
	"karsai5/tw-caldav/internal/task"
	"karsai5/tw-caldav/internal/tw"
	"karsai5/tw-caldav/pkg/taskwarrior"
	"log/slog"
	"strings"
	"time"

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
		synctime := time.Now()
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

		// Create new remote tasks locally
		for _, t := range remoteTasks {
			if t.LocalId() != nil {
				continue
			}

			slog.Info("Creating remote task locally", "desc", t.Description())
			uuid, err := createTaskInTaskwarrior(&local, t, synctime)
			if err != nil {
				slog.Error("Could not create local task", "err", err)
			}

			t.SetLocalId(&uuid)
			t.SetLastSynced(&synctime)

			_, err = remote.Client.PutCalendarObject(context.TODO(), t.Path, t.CalendarObject.Data)
			if err != nil {
				slog.Error("Could not update remote task", "err", err)
			}
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

func createTaskInTaskwarrior(tw *tw.Taskwarrior, t task.Task, synctime time.Time) (uuid string, err error) {
	opts := []string{
		t.Description(),
		fmt.Sprintf("lastsync:%s", synctime.Format(taskwarrior.TimeLayout)),
	}
	if t.RemotePath() != nil {
		opts = append(opts, fmt.Sprintf("remotepath:%q", *t.RemotePath()))
	}
	if t.Project() != "" {
		opts = append(opts, fmt.Sprintf("project:%q", escapeQuotes(t.Project())))
	}
	if t.Due() != nil {
		opts = append(opts, fmt.Sprintf("due:%s", t.Due()))
	}
	if t.Priority() != task.Unset {
		opts = append(opts, fmt.Sprintf("pririty:%s", t.Priority().String()))
	}
	for _, tag := range t.Tags() {
		if strings.Contains(tag, " ") {
			slog.Error("Cannot add tag with spaces", "tag", tag)
			continue
		}
		opts = append(opts, fmt.Sprintf("+%s", tag))
	}
	return tw.AddTask(opts...)
}

func escapeQuotes(str string) string {
	return strings.ReplaceAll(str, `"`, `\"`)
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
