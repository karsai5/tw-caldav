package task

import (
	"errors"
	"time"
)

func CreateShellTask(opts ...ShellTaskOption) ShellTask {
	task := ShellTask{}

	for _, opt := range opts {
		opt(&task)
	}

	return task
}

type ShellTaskOption func(shellTask *ShellTask)

func WithTask(t Task) ShellTaskOption {
	return func(shellTask *ShellTask) {
		shellTask.Task = &Internaltask{
			Description:  t.Description(),
			Project:      t.Project(),
			Due:          t.Due(),
			Priority:     t.Priority(),
			Tags:         t.Tags(),
			LastModified: t.LastModified(),
			LastSynced:   t.LastSynced(),
			RemotePath:   t.RemotePath(),
			LocalId:      t.LocalId(),
			Status:       t.Status(),
		}
	}
}

func WithSyncTime(synctime time.Time) ShellTaskOption {
	return func(shellTask *ShellTask) {
		shellTask.Task.LastModified = synctime
		shellTask.Task.LastSynced = &synctime
	}
}

func WithLocalId(uuid string) ShellTaskOption {
	return func(shellTask *ShellTask) {
		shellTask.Task.LocalId = &uuid
	}
}

type Internaltask struct {
	Description  string
	Project      string
	Due          *time.Time
	Priority     Priority
	Tags         []string
	LastModified time.Time
	LastSynced   *time.Time

	RemotePath *string
	LocalId    *string

	Status Status
}

type ShellTask struct {
	Task *Internaltask
}

func (s *ShellTask) SetLastModifiedAndSyncedTime(synctime time.Time) {
	s.Task.LastSynced = &synctime
	s.Task.LastModified = synctime
}

func (s *ShellTask) SetLocalId(uuid string) {
	s.Task.LocalId = &uuid
}

// Description implements Task.
func (s ShellTask) Description() string {
	return s.Task.Description
}

// Due implements Task.
func (s ShellTask) Due() *time.Time {
	return s.Task.Due
}

// LastModified implements Task.
func (s ShellTask) LastModified() time.Time {
	return s.Task.LastModified
}

// LastSynced implements Task.
func (s ShellTask) LastSynced() *time.Time {
	return s.Task.LastSynced
}

// LocalId implements Task.
func (s ShellTask) LocalId() *string {
	return s.Task.LocalId
}

// Priority implements Task.
func (s ShellTask) Priority() Priority {
	return s.Task.Priority
}

// Project implements Task.
func (s ShellTask) Project() string {
	return s.Task.Project
}

// RemotePath implements Task.
func (s ShellTask) RemotePath() *string {
	return s.Task.RemotePath
}

// Status implements Task.
func (s ShellTask) Status() Status {
	return s.Task.Status
}

// Tags implements Task.
func (s ShellTask) Tags() []string {
	return s.Task.Tags
}

// Update implements Task.
func (s ShellTask) Update(t Task) error {
	return errors.New("Update not possible on this kind of task")
}

var _ Task = ShellTask{}
