package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	"go.elastic.co/apm/module/apmhttp"
)

// Agenda represents an agenda for a day
type Agenda struct {
	AllowedDomains []string                              `json:"-"`
	Date           time.Time                             `json:"date"`
	Day            AgendaDate                            `json:"day"`
	Events         []AgendaEvent                         `json:"events"`
	HTMLSelector   string                                `json:"-"`
	HTMLProcessor  func(a *Agenda, e *colly.HTMLElement) `json:"-"`
	ID             string                                `json:"id"`
	Owner          string                                `json:"owner"`
	URL            string                                `json:"url"`
	URLFormat      string                                `json:"-"`
}

// Scrap scrappes an agenda
func (a *Agenda) Scrap(ctx context.Context) error {
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains(a.AllowedDomains...),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./.cansino_cache"),
		// MaxDepth is 1, so only the links on the scraped page
		// is visited, and no further links are followed
		colly.MaxDepth(1),
	)

	// instrument Colly's HTTP requests with APM Agent Go
	apmHTTPClient := apmhttp.WrapClient(http.DefaultClient)
	c.SetClient(apmHTTPClient)

	c.OnHTML(a.HTMLSelector, a.process)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
		fmt.Printf("Visiting %s\n", r.URL.String())
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	err := c.Visit(a.URL)
	if err != nil {
		fmt.Errorf("Error visiting URL [%s]: %v", a.URL, err)
	}

	return err
}

// ToJSON exports the agenda to JSON
func (a *Agenda) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

func (a *Agenda) process(e *colly.HTMLElement) {
	a.HTMLProcessor(a, e)
}

// AgendaDate represents a day
type AgendaDate struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

//ToDate converts a date into time.Time
func (ad *AgendaDate) ToDate() time.Time {
	return time.Date(ad.Year, time.Month(ad.Month), ad.Day, 0, 0, 0, 0, time.UTC)
}

// AgendaEvent represents an event in the agenda
type AgendaEvent struct {
	Date        time.Time  `json:"date"`
	Description string     `json:"description"`
	ID          string     `json:"id"`
	Location    string     `json:"location"`
	Attendance  []Attendee `json:"attendance"`
	Owner       string     `json:"owner"`
}

// ToJSON exports the event to JSON
func (ae *AgendaEvent) ToJSON() ([]byte, error) {
	return json.Marshal(ae)
}

// Attendee represents a person attending an event
type Attendee struct {
	Job      string `json:"job"`
	FullName string `json:"fullName"`
}
