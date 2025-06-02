package caldav

import (
	"context"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

func NewClient(path string, username string, pass string) (*calDav, error) {
	client := webdav.HTTPClientWithBasicAuth(nil, username, pass)
	calDavClient, err := caldav.NewClient(client, path)

	if err != nil {
		return nil, err
	}
	cd := calDav{
		Client: calDavClient,
	}

	return &cd, nil
}

type calDav struct {
	Client *caldav.Client
}

func (cd *calDav) GetTodos(calendar string) ([]caldav.CalendarObject, error) {
	query := &caldav.CalendarQuery{
		CompRequest: caldav.CalendarCompRequest{Name: "VCALENDAR", AllProps: true, AllComps: true},
		CompFilter:  caldav.CompFilter{Name: "VCALENDAR", Comps: []caldav.CompFilter{{Name: "VTODO"}}},
	}
	objects, err := cd.Client.QueryCalendar(context.TODO(), calendar, query)
	if err != nil {
		return []caldav.CalendarObject{}, err
	}
	return objects, nil
}

func (cd *calDav) GetCalendar() (caldav.Calendar, error) {
	cals, err := cd.Client.FindCalendars(context.TODO(), "")
	if err != nil {
		return caldav.Calendar{}, err
	}
	return cals[0], nil
}
