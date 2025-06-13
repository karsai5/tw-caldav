package tw

import (
	"fmt"
	"karsai5/tw-caldav/internal/sync/task"
	"karsai5/tw-caldav/internal/utils/conv"
	"karsai5/tw-caldav/pkg/taskwarrior"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Taskwarrior struct {
}

func (t *Taskwarrior) GetAllTasks() (tasks []Task, err error) {
	rawTasks, err := taskwarrior.List("+PENDING or +COMPLETED")
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
	}

	if t.Due() != nil {
		slog.Debug("time", "t", t.Due().String())
		opts = append(opts, fmt.Sprintf("due:%s", t.Due().Format(time.RFC3339)))
	}

	safeTags := []string{}
	for _, tag := range t.Tags() {
		if strings.Contains(tag, " ") {
			slog.Error("Cannot add tag with spaces", "tag", tag)
			continue
		}
		safeTags = append(safeTags, tag)
	}
	opts = append(opts, fmt.Sprintf("tags:'%s'", strings.Join(t.Tags(), ",")))
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

