package caldav

import (
	"bytes"
	"context"
	"fmt"
	"karsai5/tw-caldav/internal/sync/task"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

func NewClient(path string, username string, pass string) (*CalDavService, error) {
	client := webdav.HTTPClientWithBasicAuth(nil, username, pass)
	calDavClient, err := caldav.NewClient(client, path)

	if err != nil {
		return nil, err
	}

	cd := CalDavService{
		Client:    calDavClient,
		BaseURL:   path,
		Username:  username,
		Password:  pass,
		Calendars: make(CalendarNameToPathMap),
	}

	err = cd.PopulateCalendarMap()
	if err != nil {
		return nil, err
	}

	err = cd.CreateDefaultCalendarIfDoesNotExist()
	if err != nil {
		return nil, err
	}

	return &cd, nil
}

type CalendarNameToPathMap map[string]string

type CalDavService struct {
	Client    *caldav.Client
	BaseURL   string
	Username  string
	Password  string
	Calendars CalendarNameToPathMap
}

var DEFAULT_CALENDAR = "default"

func (cd *CalDavService) PopulateCalendarMap() error {
	calendars, err := cd.Client.FindCalendars(context.TODO(), "")
	if err != nil {
		return fmt.Errorf("While getting calendars: %w", err)
	}
	// clear existing calendars
	for key := range cd.Calendars {
		delete(cd.Calendars, key)
	}

	for _, c := range calendars {
		if c.Name == "" {
			slog.Error("Cannot use calendar without name", "path", c.Path)
			continue
		}
		if existingCalendar, exists := cd.Calendars[c.Name]; exists {
			return fmt.Errorf("Two calendars found with the same name '%s' %s %s", c.Name, existingCalendar, c.Path)
		}
		cd.Calendars[c.Name] = c.Path
	}
	slog.Debug("calendarMap", "map", cd.Calendars)
	return nil
}

func (cd *CalDavService) CreateDefaultCalendarIfDoesNotExist() error {
	if _, exists := cd.Calendars[DEFAULT_CALENDAR]; exists {
		return nil
	}

	_, err := cd.CreateCalendar(DEFAULT_CALENDAR, DEFAULT_CALENDAR)
	if err != nil {
		return err
	}
	slog.Info("Default calendar created remotely")

	return nil
}

func (cd *CalDavService) FindOrCreateCalendar(name string) (path string, err error) {
	path, exists := cd.Calendars[name]
	if exists {
		return path, nil
	}
	
	return cd.CreateCalendar(name, name)
}

func (cd *CalDavService) CreateCalendar(path, name string) (finalPath string, err error) {
	body := `
	<C:mkcalendar xmlns:D="DAV:"
           xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:set>
 <D:prop>
   <D:displayname>` + xmlEscape(name) + `</D:displayname>
   <C:supported-calendar-component-set>
     <C:comp name="VTODO"/>
   </C:supported-calendar-component-set>
 </D:prop>
  </D:set>
</C:mkcalendar>
	`

	finalPath = cd.BaseURL + "/" + path
	slog.Debug("Creating calendar", "name", name, "path", finalPath)
	req, err := http.NewRequest("MKCALENDAR", finalPath, strings.NewReader(body))
	if err != nil {
		return finalPath, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/xml")
	req.SetBasicAuth(cd.Username, cd.Password)

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return finalPath, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check for success
	if resp.StatusCode != http.StatusCreated {
		return finalPath, fmt.Errorf("failed to create calendar, status code: %d", resp.StatusCode)
	}

	// TODO: optimise this so that we don't have to recreate the calendar
	// map completely every time a new calendar is created
	err = cd.PopulateCalendarMap()
	if err != nil {
		return "", err
	}

	return cd.Calendars[name], nil
}

func (cd *CalDavService) CreateNewTodo(t task.Task) (finalPath string, err error) {
	syncTime := time.Now()
	taskCalendar := t.Project()
	if taskCalendar == "" {
		taskCalendar = DEFAULT_CALENDAR
	}

	calendarPath, exists := cd.Calendars[taskCalendar]

	if !exists {
		path, err := cd.CreateCalendar(t.Project(), t.Project())
		if err != nil {
			return finalPath, err
		}
		calendarPath = path
	}

	icalPath := fmt.Sprintf("%s%s.ical", calendarPath, *t.LocalId())

	cal := ical.NewCalendar()
	cal.Component.Name = "VCALENDAR"

	addStringProp(&cal.Component.Props, "PRODID", "-//Karsai5//tw-caldav 2025.6.7//EN")
	addTimeProp(&cal.Component.Props, "DTSTAMP", syncTime)
	addStringProp(&cal.Component.Props, "VERSION", "2.0")
	addStringProp(&cal.Component.Props, "UID", *t.LocalId())

	todo := ical.NewComponent("VTODO")

	addTimeProp(&todo.Props, "DTSTAMP", time.Now())
	updatePropsWithInformationFromTask(&todo.Props, t)

	cal.Component.Children = append(cal.Component.Children, todo)

	buf := new(bytes.Buffer)
	encoder := ical.NewEncoder(buf)
	encoder.Encode(cal)
	slog.Debug("Creating caldav ical", "path", icalPath, "ical", buf.String())

	res, err := cd.Client.PutCalendarObject(context.TODO(), icalPath, cal)
	if err != nil {
		return "", err
	}

	return res.Path, nil
}

func updatePropsWithInformationFromTask(props *ical.Props, t task.Task) {
	addStringProp(props, "SUMMARY", t.Description())

	if status := statusToCalDavStatus(t.Status()); status != "" {
		addStringProp(props, "STATUS", status)
	}

	addTimeProp(props, "LAST-MODIFIED", time.Now())

	if t.Due() != nil {
		addTimeProp(props, "DUE", *t.Due())
		addTimeProp(props, "DTSTART", *t.Due())
	}

	if t.Priority() != task.PriorityUnset {
		addStringProp(props, "PRIORITY", fmt.Sprintf("%d", t.Priority()))
	}

	if t.LocalId() != nil {
		addStringProp(props, "UID", *t.LocalId())
	}

	if len(t.Tags()) > 0 {
		addStringProp(props, "CATEGORIES", strings.Join(t.Tags(), ","))
	}

	desc := fmt.Sprintf("taskwarrior_id=%s", *t.LocalId())
	if existingDescription := props.Get("DESCRIPTION"); existingDescription != nil {
		if strings.Index(existingDescription.Value, "taskwarrior_id=") < 0 {
			desc = fmt.Sprintf("%s\n%s", existingDescription.Value, desc)
		}
	}
	addStringProp(props, "DESCRIPTION", desc)
}

func statusToCalDavStatus(s task.Status) string {
	switch s {
	case task.StatusComplete:
		return "COMPLETE"
	case task.StatusDeleted:
		return "CANCELLED"
	default:
		return ""
	}

}

func addStringProp(props *ical.Props, name string, value string) {
	prop := ical.NewProp(name)
	prop.SetText(value)
	props.Set(prop)
}

func addTimeProp(props *ical.Props, name string, value time.Time) {
	prop := ical.NewProp(name)
	prop.SetDateTime(value)
	props.Set(prop)
}

func (cd *CalDavService) GetAllTodos() (todos []Todo, err error) {
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
			todo, err := cd.mapTodo(&cal, &calTodo)
			if err != nil {
				return todos, fmt.Errorf("While creating todo: %w", err)
			}
			todos = append(todos, *todo)
		}
	}
	return todos, nil
}

func (cd *CalDavService) GetTodosForCalendar(calendarPath string) ([]caldav.CalendarObject, error) {
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

func (cd *CalDavService) mapTodo(cal *caldav.Calendar, calObj *caldav.CalendarObject) (*Todo, error) {
	if cal == nil {
		return nil, fmt.Errorf("cal can't be nil")
	}
	if calObj == nil {
		return nil, fmt.Errorf("calObj can't be nil")
	}


	vTodoIndex := slices.IndexFunc(calObj.Data.Children, func(child *ical.Component) bool {
		if child.Name == "VTODO" {
			return true;
		}
		return false
	})

	if vTodoIndex < 0 {
		return nil, fmt.Errorf("Could not find VTODO in calObj")
	}

	todo := Todo{
		calDavService:  cd,
		Calendar:       cal,
		CalendarObject: calObj,
		TodoComponent:  calObj.Data.Children[vTodoIndex],
		Path:           calObj.Path,
	}

	return &todo, nil
}

func xmlEscape(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(input)
}
