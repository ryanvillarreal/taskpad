package calendar

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
	"github.com/google/uuid"
)

// Service defines the interface for calendar operations.
// This allows swapping CalDAV for another provider later.
type Service interface {
	CreateEvent(ctx context.Context, event Event) (string, error) // returns event ID
	UpdateEvent(ctx context.Context, eventID string, event Event) error
	DeleteEvent(ctx context.Context, eventID string) error
}

// Event represents a calendar event to sync.
type Event struct {
	Title   string
	DueDate time.Time
}

// CalDAVConfig holds the configuration for connecting to a CalDAV server.
type CalDAVConfig struct {
	ServerURL    string // e.g. https://caldav.example.com
	Username     string
	Password     string
	CalendarPath string // e.g. /user/calendars/taskpad/
}

type caldavService struct {
	client       *caldav.Client
	calendarPath string
}

// NewCalDAV creates a new CalDAV calendar service.
func NewCalDAV(cfg CalDAVConfig) (Service, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &basicAuthTransport{
			username:  cfg.Username,
			password:  cfg.Password,
			transport: http.DefaultTransport,
		},
	}

	client, err := caldav.NewClient(httpClient, cfg.ServerURL)
	if err != nil {
		return nil, fmt.Errorf("create caldav client: %w", err)
	}

	return &caldavService{
		client:       client,
		calendarPath: cfg.CalendarPath,
	}, nil
}

func (s *caldavService) CreateEvent(ctx context.Context, event Event) (string, error) {
	eventID := uuid.New().String()
	cal := buildICalEvent(eventID, event)

	path := s.calendarPath + eventID + ".ics"
	_, err := s.client.PutCalendarObject(ctx, path, cal)
	if err != nil {
		return "", fmt.Errorf("create calendar event: %w", err)
	}
	return eventID, nil
}

func (s *caldavService) UpdateEvent(ctx context.Context, eventID string, event Event) error {
	cal := buildICalEvent(eventID, event)

	path := s.calendarPath + eventID + ".ics"
	_, err := s.client.PutCalendarObject(ctx, path, cal)
	if err != nil {
		return fmt.Errorf("update calendar event: %w", err)
	}
	return nil
}

func (s *caldavService) DeleteEvent(ctx context.Context, eventID string) error {
	path := s.calendarPath + eventID + ".ics"
	err := s.client.RemoveAll(ctx, path)
	if err != nil {
		return fmt.Errorf("delete calendar event: %w", err)
	}
	return nil
}

func buildICalEvent(uid string, event Event) *ical.Calendar {
	cal := ical.NewCalendar()
	cal.Props.SetText(ical.PropProductID, "-//taskpad//EN")
	cal.Props.SetText(ical.PropVersion, "2.0")

	vevent := ical.NewEvent()
	vevent.Props.SetText(ical.PropUID, uid)
	vevent.Props.SetText(ical.PropSummary, event.Title)
	vevent.Props.SetDateTime(ical.PropDateTimeStamp, time.Now().UTC())
	vevent.Props.SetDateTime(ical.PropDateTimeStart, event.DueDate)
	vevent.Props.SetDateTime(ical.PropDateTimeEnd, event.DueDate.Add(1*time.Hour))

	cal.Children = append(cal.Children, vevent.Component)

	return cal
}

// basicAuthTransport adds basic auth to every request.
type basicAuthTransport struct {
	username  string
	password  string
	transport http.RoundTripper
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.username, t.password)
	return t.transport.RoundTrip(req)
}
