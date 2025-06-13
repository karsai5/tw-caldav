package caldav

import (
	"bytes"
	"context"
	"fmt"
	"karsai5/tw-caldav/internal/sync/task"
	"log/slog"
	"regexp"
	"strconv"
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
func (t *Todo) Update(u task.Task) (task.Task, error) {
	if t.Project() != u.Project() {
		currentFolderPath, fileName := getpathAndFilename(t.Path)
		newDirPath, err := t.calDavService.FindOrCreateCalendar(u.Project())
		if err != nil {
			return nil, err
		}
		newPath := newDirPath + fileName
		slog.Debug("Moving ical", "oldPath", currentFolderPath, "newPath", newPath)
		err = t.calDavService.Client.Move(context.TODO(), t.Path, newPath, nil)
		if err != nil {
			return nil, fmt.Errorf("While moving task to new calendar: %w", err)
		}
		return task.CreateShellTask(task.WithTask(u), task.WithRemotePath(newPath)), nil
	}

	// NOTE: Option to delete and create task instead

	// if t.Project() != u.Project() {
	// 	err := t.Delete()
	// 	if err != nil {
	// 		return u, fmt.Errorf("While deleting task to move project: %w", err)
	// 	}
	//
	// 	path, err := t.calDavService.CreateNewTodo(u)
	// 	if err != nil {
	// 		return u, fmt.Errorf("While creating new task to move project: %w", err)
	// 	}
	//
	// 	return task.CreateShellTask(task.WithTask(u), task.WithRemotePath(path)), nil
	// }

	updatePropsWithInformationFromTask(&t.TodoComponent.Props, u)

	buf := new(bytes.Buffer)
	encoder := ical.NewEncoder(buf)
	encoder.Encode(t.CalendarObject.Data)
	slog.Debug("Updating caldav ical", "path", t.Path, "ical", buf.String())

	_, err := t.calDavService.Client.PutCalendarObject(context.TODO(), t.Path, t.CalendarObject.Data)
	if err != nil {
		return u, err
	}
	return u, err

}

func getpathAndFilename(s string) (string, string) {
	idx := strings.LastIndex(s, "/")
	if idx == -1 {
		return s, ""
	}
	return s[:idx] + "/", s[idx+1:]
}

// Status implements task.Task.
func (t *Todo) Status() task.Status {
	prop := t.TodoComponent.Props.Get("STATUS")
	if prop == nil {
		return task.StatusPending
	}
	switch prop.Value {
	case "COMPLETED":
		return task.StatusComplete
	case "CANCELLED":
		return task.StatusDeleted
	default:
		return task.StatusPending
	}
}

// LocalId implements task.Task.
func (t *Todo) LocalId() *string {
	prop := t.TodoComponent.Props.Get("DESCRIPTION")
	if prop == nil {
		return nil
	}
	r := regexp.MustCompile("taskwarrior_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
	id := r.FindString(prop.Value)
	id = strings.Replace(id, "taskwarrior_id=", "", -1)
	if id == "" {
		return nil
	}
	return &id
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
	priority, err := strconv.Atoi(prop.Value)
	if err != nil {
		slog.Error("Could not pass priority: %s", "priority", prop.Value)
	}

	return task.Priority(priority)
}

// Project implements task.Task.
func (t *Todo) Project() string {
	proj := t.Calendar.Name
	if proj == DEFAULT_CALENDAR {
		return ""
	}
	return proj
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

func (t *Todo) GetStringProp(key string) string {
	prop := t.TodoComponent.Props.Get(key)
	if prop == nil {
		return ""
	}
	return prop.Value
}

func (t *Todo) Delete() error {
	return t.calDavService.Client.RemoveAll(context.TODO(), t.Path)
}
