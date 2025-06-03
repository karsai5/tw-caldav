package tw

import (
	"fmt"
	"karsai5/tw-caldav/internal/task"
	"karsai5/tw-caldav/pkg/taskwarrior"
	"time"
)

type Task struct {
	task taskwarrior.Task
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
	// TODO: implement
	return nil
}

// LocalId implements task.Task.
func (t *Task) LocalId() *string {
	return &t.task.UUID
}

// Priority implements task.Task.
func (t *Task) Priority() task.Priority {

	switch t.task.Priority {
	case "H":
		return task.High
	case "M":
		return task.Medium
	case "L":
		return task.Low
	default:
		return task.Unset
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

func (t *Taskwarrior) GetAllTasks() (tasks []*Task, err error) {
	rawTasks, err := taskwarrior.List("")
	if err != nil {
		return tasks, fmt.Errorf("While getting tasks from taskwarrior: %w", err)
	}
	for _, t := range rawTasks {
		tasks = append(tasks, &Task{task: t})
	}
	return tasks, err
}
