package sync

import (
	"log/slog"
	"os"
	"strings"

	"github.com/karsai5/tw-caldav/internal/sync/task"

	"github.com/jedib0t/go-pretty/v6/table"
)

func DebugTask(t task.Task) {
	tags := strings.Join(t.Tags(), ", ")
	slog.Debug("Task", "desc", t.Description(), "project", t.Project(), "due", t.Due(), "priority", t.Priority(), "tags", tags)
}

func PrintTable(tasks []task.Task) {
	tab := table.NewWriter()
	tab.SetOutputMirror(os.Stdout)
	tab.AppendHeader(table.Row{"Status", "desc", "proj", "due", "priority", "tags", "last modified", "path", "id"})
	for _, t := range tasks {
		tags := strings.Join(t.Tags(), ", ")

		desc := t.Description()
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}

		remotePath := ""
		if t.RemotePath() != nil {
			remotePath = "✔️"
		}

		localId := ""
		if t.LocalId() != nil {
			localId = "✔️"
		}

		tab.AppendRow(table.Row{t.Status(), desc, t.Project(), t.Due(), t.Priority(), tags, t.LastModified(), remotePath, localId})
	}
	tab.Render()
}
