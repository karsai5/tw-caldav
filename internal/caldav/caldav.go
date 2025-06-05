package caldav

import (
	"context"
	"fmt"
	"karsai5/tw-caldav/internal/task"
	"log/slog"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

func NewClient(path string, username string, pass string) (*CalDav, error) {
	client := webdav.HTTPClientWithBasicAuth(nil, username, pass)
	calDavClient, err := caldav.NewClient(client, path)

	if err != nil {
		return nil, err
	}
	cd := CalDav{
		Client: calDavClient,
	}

	return &cd, nil
}

type CalDav struct {
	Client *caldav.Client
}

func (cd *CalDav) GetAllTodos() (todos []*Todo, err error) {
	calendars, err := cd.Client.FindCalendars(context.TODO(), "")
	if err != nil {
		return todos, fmt.Errorf("While getting calendars: %w", err)
	}

	for _, cal := range calendars {
		calTodos, err := cd.GetTodosForCalendar(cal.Path)
		if err != nil {
			return todos, fmt.Errorf("While getting todos for calendar: %w", err)
		}
		for _, calTodo := range calTodos {
			todo, err := createTodo(&cal, &calTodo)
			if err != nil {
				return todos, fmt.Errorf("While creating todo: %w", err)
			}
			todos = append(todos, todo)
		}
	}
	return todos, nil
}

func (cd *CalDav) GetTodosForCalendar(calendarPath string) ([]caldav.CalendarObject, error) {
	query := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{Name: "VCALENDAR", AllProps: true, AllComps: true},
		CompFilter:  caldav.CompFilter{Name: "VCALENDAR", Comps: []caldav.CompFilter{{Name: "VTODO"}}},
	}
	objects, err := cd.Client.QueryCalendar(context.TODO(), calendarPath, query)
	if err != nil {
		return []caldav.CalendarObject{}, err
	}
	return objects, nil
}

func createTodo(cal *caldav.Calendar, calObj *caldav.CalendarObject) (*Todo, error) {
	if cal == nil {
		return nil, fmt.Errorf("cal can't be nil")
	}
	if calObj == nil {
		return nil, fmt.Errorf("calObj can't be nil")
	}

	if len(calObj.Data.Children) > 1 {
		return nil, fmt.Errorf("Incorrect number of children for calObj, should be 1, found %d", len(calObj.Data.Children))
	}

	todo := Todo{
		Calendar:       cal,
		CalendarObject: calObj,
		TodoComponent:  calObj.Data.Children[0],
		Path:           calObj.Path,
	}

	return &todo, nil
}

type Todo struct {
	Calendar       *caldav.Calendar
	CalendarObject *caldav.CalendarObject
	TodoComponent  *ical.Component
	Path           string
}

// LocalId implements task.Task.
func (t *Todo) LocalId() *string {
	prop := t.TodoComponent.Props.Get("TW-ID")
	if prop == nil {
		return nil
	}
	return &prop.Value
}

func (t *Todo) SetLocalId(str *string) {
	prop := ical.NewProp("TW-ID")
	if str != nil {
		prop.SetText(*str)
	}
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

func (t *Todo) SetLastSynced(lastsync *time.Time) {
	prop := ical.NewProp("TW-LAST-SYNCED")
	if lastsync != nil {
		prop.SetDate(*lastsync)
	}
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

func (t *Todo) GetStringProp(key string) string {
	prop := t.TodoComponent.Props.Get(key)
	if prop == nil {
		return ""
	}
	return prop.Value
}
