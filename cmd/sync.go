/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"karsai5/tw-caldav/internal/caldav"
	"karsai5/tw-caldav/internal/taskwarrior"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/google/uuid"
	"github.com/lmittmann/tint"
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

		tasks, err := taskwarrior.List("+PENDING")
		if err != nil {
			panic(err)
		}

		for _, t := range tasks {
			if t.CalDavId != "" {
				continue
			}
			slog.Info("Processing new task", "id", t.Id, "desc", t.Description)
			id := uuid.New().String()
			path := fmt.Sprintf("%s%s.ics", cal.Path, id)

			props := ical.Props{}
			props.SetText("SUMMARY", cleanString(t.Description))
			props.SetText("TWID", t.UUID)
			props.SetText("UID", id)
			props.SetDate("DTSTAMP", time.Now())
			props.SetDate("LAST-MODIFIED", t.Modified)
			task := ical.Component{
				Name:  "VTODO",
				Props: props,
			}

			calProps := ical.Props{}
			calProps.SetText("PRODID", "-//karsai5//taskwarrior sync//EN")
			calProps.SetText("VERSION", "2.0")

			calToInsert := ical.Calendar{
				Component: &ical.Component{
					Name:     "VCALENDAR",
					Props:    calProps,
					Children: []*ical.Component{&task},
				},
			}

			_, err = client.Client.PutCalendarObject(context.TODO(), path, &calToInsert)
			if err != nil {
				panic(err)
			}
			t.Append(fmt.Sprintf("caldavid:%s", id))
		}

	},
}

func cleanString(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str, "\r", ""), "\n", "")
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
