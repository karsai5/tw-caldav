package tw

import (
	"fmt"
	"karsai5/tw-caldav/internal/sync/task"
	"karsai5/tw-caldav/pkg/taskwarrior"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	task taskwarrior.Task
}

// Update implements task.Task.
func (*Task) Update(t task.Task) error {
	panic("unimplemented")
}

// Status implements task.Task.
func (t *Task) Status() task.Status {
	panic("unimplemented")
}

// Description implements task.Task.
func (t *Task) Description() string {
	return t.task.Description
}

// Due implements task.Task.
func (t *Task) Due() *time.Time {
	return t.task.Due
}

// LastModified implements task.Task.
func (t *Task) LastModified() time.Time {
	return t.task.Modified
}

// LastSynced implements task.Task.
func (t *Task) LastSynced() *time.Time {
	return t.task.LastSync
}

// LocalId implements task.Task.
func (t *Task) LocalId() *string {
	return &t.task.UUID
}

// Priority implements task.Task.
func (t *Task) Priority() task.Priority {

	switch t.task.Priority {
	case "H":
		return task.PriorityHigh
	case "M":
		return task.PriorityMedium
	case "L":
		return task.PriorityLow
	default:
		return task.PriorityUnset
	}

}

// Project implements task.Task.
func (t *Task) Project() string {
	return t.task.Project
}

// RemotePath implements task.Task.
func (t *Task) RemotePath() *string {
	if t.task.RemotePath == "" {
		return nil
	}
	return &t.task.RemotePath
}

// Tags implements task.Task.
func (t *Task) Tags() []string {
	return t.task.Tags
}

type Taskwarrior struct {
}

func escapeQuotes(str string) string {
	return strings.ReplaceAll(str, `"`, `\"`)
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

	out, err := taskwarrior.Run(createAddTaskCmdOptions(t)...)
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

func createAddTaskCmdOptions(t task.Task) []string {
	opts := []string{
		"add",
		t.Description(),
	}
	if t.LastSynced() != nil {
		opts = append(opts, fmt.Sprintf("lastsync:%s", t.LastSynced().Format(taskwarrior.TimeLayout)))
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
	if t.Priority() != task.PriorityUnset {
		opts = append(opts, fmt.Sprintf("pririty:%s", t.Priority().String()))
	}
	for _, tag := range t.Tags() {
		if strings.Contains(tag, " ") {
			slog.Error("Cannot add tag with spaces", "tag", tag)
			continue
		}
		opts = append(opts, fmt.Sprintf("+%s", tag))
	}

	return opts
}

func (t *Taskwarrior) RemoveTask(uuid string) error {
	if uuid == "" {
		return fmt.Errorf("UUID must be set")
	}
	out, err := taskwarrior.Run("rc.confirmation=off", fmt.Sprintf("uuid:%s", uuid), "delete")
	if err != nil {
		return fmt.Errorf("Error deleting task: %s: %w", out, err)
	}
	return nil
}

func extractNumber(s string) (int, error) {
	re := regexp.MustCompile(`\d+`)
	numStr := re.FindString(s)
	return strconv.Atoi(numStr)
}
