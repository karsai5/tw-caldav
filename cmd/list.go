/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"karsai5/tw-caldav/internal/caldav"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// ctx := context.TODO()
		// Set global logger with custom options
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: time.Kitchen,
			}),
		))

		client, err := caldav.NewClient(url, username, password)
		if err != nil {
			panic(err)
		}

		cal, err := client.GetCalendar()
		if err != nil {
			panic(err)
		}

		slog.Info("Found calendar", "path", cal.Path)

		objects, err := client.GetTodos(cal.Path)
		if err != nil {
			panic(err)
		}

		for _, obj := range objects {
			// slog.Info("Calendar object", "path", obj.Path, "uid", obj.Data.Props)
			// slog.Info("Calendar object", "obj", obj)

			vtodo := obj.Data.Children[0]
			slog.Info("Calendar child", "summary", vtodo.Props.Get("SUMMARY").Value, "UID", vtodo.Props.Get("UID").Value, "TWID", vtodo.Props.Get("TWID"), "LAST-MODIFIED", vtodo.Props.Get("LAST-MODIFIED"))
			// slog.Info("Calendar child", "obj", vtodo)

			// vtodo.Props.SetText("TWID", "taskwarrior-ID")
			//
			// _, err = client.Client.PutCalendarObject(context.TODO(), obj.Path, obj.Data)
			// if err != nil {
			// 	panic(err)
			// }
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
