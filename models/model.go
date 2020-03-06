package models

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gocolly/colly/v2"
	log "github.com/sirupsen/logrus"
	"go.elastic.co/apm/module/apmhttp"
)

// Agenda represents an agenda for a day
type Agenda struct {
	AllowedDomains []string   `json:"-"`
	Date           time.Time  `json:"date"`
	Day            AgendaDate `json:"day"`
	// DoPost: if the public agenda requires POST http method
	DoPost        bool                                  `json:"-"`
	Events        []AgendaEvent                         `json:"events"`
	HTMLSelector  string                                `json:"-"`
	HTMLProcessor func(a *Agenda, e *colly.HTMLElement) `json:"-"`
	// JSONProcessor only for processing POST requests
	JSONProcessor func(a *Agenda, body []byte) `json:"-"`
	ID            string                       `json:"id"`
	Owner         string                       `json:"owner"`
	Region        string                       `json:"-"`
	Payload       string                       `json:"-"`
	URL           string                       `json:"url"`
	URLFormat     string                       `json:"-"`
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

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		if a.DoPost {
			r.Headers.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
		} else {
			r.Ctx.Put("url", r.URL.String())
		}

		log.WithFields(log.Fields{
			"url": r.URL.String(),
		}).Debug("Visiting url")
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		log.WithFields(log.Fields{
			"url":      r.Request.URL,
			"response": r,
			"error":    err,
		}).Error("Failed to parse HTML")
	})

	var err error
	if a.DoPost {
		c.OnResponse(func(r *colly.Response) {
			a.JSONProcessor(a, r.Body)
		})

		err = c.PostRaw(a.URL, []byte(a.Payload))
	} else {
		c.OnHTML(a.HTMLSelector, a.htmlProcess)

		err = c.Visit(a.URL)
	}
	if err != nil {
		log.WithFields(log.Fields{
			"url":   a.URL,
			"error": err,
		}).Error("Error visiting URL")
	}

	return err
}

// ToJSON exports the agenda to JSON
func (a *Agenda) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

func (a *Agenda) htmlProcess(e *colly.HTMLElement) {
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
	Date                time.Time  `json:"date"`
	Description         string     `json:"description"`
	OriginalDescription string     `json:"originalDescription"`
	ID                  string     `json:"id"`
	Location            string     `json:"location"`
	OriginalLocation    string     `json:"originalLocation"`
	Attendance          []Attendee `json:"attendance"`
	Owner               string     `json:"owner"`
	Region              string     `json:"region"`
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

// Region represents a region
type Region struct {
	Name      string
	DoPost    bool
	StartDate AgendaDate // when the agenda started to share agendas publicly
}
