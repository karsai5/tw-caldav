package caldav

import (
	"context"
	"karsai5/tw-caldav/internal/sync/task"
	"karsai5/tw-caldav/internal/utils/comp"
	"log/slog"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

type Todo struct {
	calDavService  *CalDavService
	Calendar       *caldav.Calendar
	CalendarObject *caldav.CalendarObject
	TodoComponent  *ical.Component
	Path           string
}

// Update implements task.Task.
func (t *Todo) Update(u task.Task) error {
	if newDesc := u.Description(); t.Description() != newDesc {
		t.setDescription(newDesc)
	}

	if newProj := u.Project(); t.Project() != newProj {
		panic("unimplemented")
	}

	if newDue := u.Due(); !comp.EqualTimePtrs(newDue, t.Due()) {
		panic("unimplemented")
	}

	if newPriority := u.Priority(); t.Priority() != newPriority {
		panic("unimplemented")
	}

	// TODO: Tags
	// TODO: Status

	if updatedLocalId := u.LocalId(); !comp.EqualPtrs(t.LocalId(), updatedLocalId) {
		t.setLocalId(*updatedLocalId)
	}

	if updatedLastModified := u.LastModified(); t.LastModified() != updatedLastModified {
		t.setLastModified(updatedLastModified)
	}

	if updatedLastSynced := u.LastSynced(); !comp.EqualPtrs(t.LastSynced(), updatedLastSynced) {
		t.setLastSynced(*updatedLastSynced)
	}

	_, err := t.calDavService.Client.PutCalendarObject(context.TODO(), t.Path, t.CalendarObject.Data)

	return err
}

// Status implements task.Task.
func (t *Todo) Status() task.Status {
	prop := t.TodoComponent.Props.Get("TW-ID")
	if prop == nil {
		return task.StatusPending
	}
	switch prop.Value {
	case "COMPLETED":
		return task.StatusComplete
	case "CANCELLED":
		return task.StatusDeleted
	default:
		return task.StatusUnset
	}
}

// LocalId implements task.Task.
func (t *Todo) LocalId() *string {
	prop := t.TodoComponent.Props.Get("TW-ID")
	if prop == nil {
		return nil
	}
	return &prop.Value
}

func (t *Todo) setLocalId(str string) {
	prop := ical.NewProp("TW-ID")
	prop.SetText(str)
	t.TodoComponent.Props.Set(prop)
}

// RemotePath implements task.Task.
func (t *Todo) RemotePath() *string {
	return &t.CalendarObject.Path
}

// LastModified implements task.Task.
func (t *Todo) LastModified() time.Time {
	prop := t.TodoComponent.Props.Get("LAST-MODIFIED")
	time, err := prop.DateTime(&time.Location{})
	if err != nil {
		slog.Error("Could not parse time: %s", "time", prop.Value)
	}
	return time
}

func (t *Todo) setLastModified(lastsync time.Time) {
	prop := ical.NewProp("LAST-MODIFIED")
	prop.SetDate(lastsync)
	t.TodoComponent.Props.Set(prop)
}

// LastSynced implements task.Task.
func (t *Todo) LastSynced() *time.Time {
	prop := t.TodoComponent.Props.Get("TW-LAST-SYNCED")
	if prop == nil {
		return nil
	}
	time, err := prop.DateTime(&time.Location{})
	if err != nil {
		slog.Error("Could not parse time: %s", "time", prop.Value)
		return nil
	}
	return &time
}

func (t *Todo) setLastSynced(lastsync time.Time) {
	prop := ical.NewProp("TW-LAST-SYNCED")
	prop.SetDate(lastsync)
	t.TodoComponent.Props.Set(prop)
}

// Due implements task.Task.
func (t *Todo) Due() *time.Time {
	prop := t.TodoComponent.Props.Get("DUE")
	if prop == nil {
		return nil
	}
	time, err := prop.DateTime(&time.Location{})
	if err != nil {
		slog.Error("Could not parse time: %s", "time", prop.Value)
	}
	return &time
}

// Priority implements task.Task.
func (t *Todo) Priority() task.Priority {
	prop := t.TodoComponent.Props.Get("PRIORITY")
	if prop == nil {
		return 0
	}
	priority, err := prop.Int()
	if err != nil {
		slog.Error("Could not pass priority: %s", "priority", prop.Value)
	}

	return task.Priority(priority)
}

// Project implements task.Task.
func (t *Todo) Project() string {
	return t.Calendar.Name
}

// Tags implements task.Task.
func (t *Todo) Tags() []string {
	prop := t.TodoComponent.Props.Get("CATEGORIES")
	if prop == nil {
		return []string{}
	}
	tags := strings.Split(prop.Value, ",")
	return tags
}

func (t *Todo) Description() string {
	return t.GetStringProp("SUMMARY")
}

func (t *Todo) setDescription(desc string) {
	prop := ical.NewProp("TW-LAST-SYNCED")
	prop.SetText(desc)
	t.TodoComponent.Props.Set(prop)
}

func (t *Todo) GetStringProp(key string) string {
	prop := t.TodoComponent.Props.Get(key)
	if prop == nil {
		return ""
	}
	return prop.Value
}
