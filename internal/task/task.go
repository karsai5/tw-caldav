package task

import "time"

type Task interface {
	Description() string
	Project() string
	Due() *time.Time
	Priority() Priority
	Tags() []string
	LastModified() time.Time
	LastSynced() *time.Time

	RemotePath() *string
	LocalId() *string

	Status() Status
}

type Status int8

func (s Status) String() string {
	switch s {
	case StatusComplete:
		return "complete"
	case StatusPending:
		return "pending"
	case StatusDeleted:
		return "deleted"
	default:
		return ""
	}
}

type Priority int8

func (p Priority) String() string {
	switch p {
	case 9:
		return "L"
	case 5:
		return "M"
	case 1:
		return "H"
	default:
		return ""
	}
}

const (
	PriorityUnset  Priority = 0
	PriorityLow             = 9
	PriorityMedium          = 5
	PriorityHigh            = 1
)

const (
	StatusUnset Status = iota
	StatusComplete
	StatusPending
	StatusDeleted
)
