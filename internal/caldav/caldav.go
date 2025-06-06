package caldav

import (
	"context"
	"fmt"

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
		Client: calDavClient,
	}

	return &cd, nil
}

type CalDavService struct {
	Client *caldav.Client
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

	if len(calObj.Data.Children) > 1 {
		return nil, fmt.Errorf("Incorrect number of children for calObj, should be 1, found %d", len(calObj.Data.Children))
	}

	todo := Todo{
		calDavService:  cd,
		Calendar:       cal,
		CalendarObject: calObj,
		TodoComponent:  calObj.Data.Children[0],
		Path:           calObj.Path,
	}

	return &todo, nil
}
