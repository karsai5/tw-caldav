package tw

import (
	"fmt"
	"time"

	"github.com/karsai5/tw-caldav/internal/sync/task"
	"github.com/karsai5/tw-caldav/pkg/taskwarrior"
)

type Task struct {
	task taskwarrior.Task
}

// Status implements task.Task.
func (t *Task) Status() task.Status {
	switch t.task.Status {
	case "completed":
		return task.StatusComplete
	case "deleted":
		return task.StatusDeleted
	default:
		return task.StatusPending
	}
}

// Description implements task.Task.
func (t *Task) Description() string {
	return t.task.Description
}

// Due implements task.Task.
func (t *Task) Due() *time.Time {
	return t.task.Due
}

func (t *Task) Delete() error {
	out, err := taskwarrior.Run("rc.confirmation=off", fmt.Sprintf("uuid:%s", *t.LocalId()), "delete")
	if err != nil {
		return fmt.Errorf("Error deleting task: %s: %w", out, err)
	}
	return nil
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

// Update implements task.Task.
func (t *Task) Update(u task.Task) (task.Task, error) {
	modCmdOpts := append(
		[]string{fmt.Sprintf("uuid:%s", *t.LocalId()), "mod"},
		createCmdOptionsForMetadata(u)...,
	)

	out, err := taskwarrior.Run(modCmdOpts...)
	if err != nil {
		return nil, fmt.Errorf("While updating local task: %s: %w", string(out), err)
	}

	return u, nil
}
