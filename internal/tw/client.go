package tw

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/karsai5/tw-caldav/internal/sync/task"
	"github.com/karsai5/tw-caldav/internal/utils/conv"
	"github.com/karsai5/tw-caldav/pkg/taskwarrior"
)

type Taskwarrior struct {
}

func (tw *Taskwarrior) GetTask(uuid string) (Task, error) {
	rawTasks, err := taskwarrior.List(fmt.Sprintf("uuid:%s", uuid))
	if err != nil {
		return Task{}, fmt.Errorf("While getting tasks from taskwarrior: %w", err)
	}
	if len(rawTasks) != 1 {
		return Task{}, fmt.Errorf("Wrong number of tasks returned, expected 1 got %d", len(rawTasks))
	}
	return Task{task: rawTasks[0]}, err
}

func (t *Taskwarrior) GetAllTasks() (tasks []Task, err error) {
	// rawTasks, err := taskwarrior.List("+PENDING or +COMPLETED")
	rawTasks, err := taskwarrior.List("+PENDING or (+COMPLETED and modified.after:2025-06-13)")
	if err != nil {
		return tasks, fmt.Errorf("While getting tasks from taskwarrior: %w", err)
	}
	for _, t := range rawTasks {
		tasks = append(tasks, Task{task: t})
	}
	return tasks, err
}

func (tw *Taskwarrior) AddTask(t task.Task) (uuid string, err error) {
	addCmdOpts := append([]string{
		"add",
		t.Description(),
	}, createCmdOptionsForMetadata(t)...)

	out, err := taskwarrior.Run(addCmdOpts...)
	if err != nil {
		return "", fmt.Errorf("While adding task: %w", err)
	}

	taskNumber, err := extractNumber(out)
	if err != nil {
		return "", fmt.Errorf("While getting tasknumber: %w", err)
	}

	uuid, err = taskwarrior.Run("_get", fmt.Sprintf("%d.uuid", taskNumber))
	if err != nil {
		return "", fmt.Errorf("While getting uuid of task: %w", err)
	}

	uuid = strings.TrimSpace(uuid)

	slog.Debug("Adding task", "id", taskNumber, "uuid", uuid)

	return uuid, err
}

func createCmdOptionsForMetadata(t task.Task) []string {
	opts := []string{
		fmt.Sprintf("description:%q", t.Description()),
		fmt.Sprintf("remotepath:%q", conv.SafeStringPtr(t.RemotePath())),
		fmt.Sprintf("project:%q", escapeQuotes(t.Project())),
		fmt.Sprintf("priority:%s", t.Priority().String()),
		fmt.Sprintf("status:%s", t.Status().String()),
	}

	if t.Due() != nil {
		opts = append(opts, fmt.Sprintf("due:%s", t.Due().Format(time.RFC3339)))
	} else {
		opts = append(opts, "due:''")
	}

	safeTags := []string{}
	for _, tag := range t.Tags() {
		if strings.Contains(tag, " ") {
			slog.Error("Cannot add tag with spaces", "tag", tag)
			continue
		}
		safeTags = append(safeTags, tag)
	}
	opts = append(opts, fmt.Sprintf("tags:%q", strings.Join(safeTags, ",")))
	return opts
}

func extractNumber(s string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(s)
	return strconv.Atoi(numStr)
}

func escapeQuotes(str string) string {
	return strings.ReplaceAll(str, `"`, `\"`)
}
