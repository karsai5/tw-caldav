package task

import (
	"log/slog"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

func DebugTask(t Task) {
	tags := strings.Join(t.Tags(), ", ")
	slog.Debug("Task", "desc", t.Description(), "project", t.Project(), "due", t.Due(), "priority", t.Priority(), "tags", tags)
}

func PrintTable(tasks []Task) {
	tab := table.NewWriter()
	tab.SetOutputMirror(os.Stdout)
	tab.AppendHeader(table.Row{"desc", "proj", "due", "priority", "tags"})
	for _, t := range tasks {
		tags := strings.Join(t.Tags(), ", ")
		desc := t.Description()
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}
		tab.AppendRow(table.Row{desc, t.Project(), t.Due(), t.Priority(), tags})
	}
	tab.Render()
}
