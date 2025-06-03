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
	Unset  Priority = 0
	Low             = 9
	Medium          = 5
	High            = 1
)
