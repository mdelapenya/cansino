package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocolly/colly"
)

// Agenda represents an agenda for a day
type Agenda struct {
	AllowedDomains []string                   `yaml:"domains"`
	Date           AgendaDate                 `yaml:"day"`
	Events         []AgendaEvent              `yaml:"events"`
	HTMLSelector   string                     `yaml:"htmlSelector"`
	HTMLProcessor  func(e *colly.HTMLElement) `json:"-"`
	Owner          string                     `yaml:"owner"`
	URL            string                     `yaml:"url"`
}

// Scrap scrappes an agenda
func (a *Agenda) Scrap(ctx context.Context) error {
	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains(a.AllowedDomains...),

		// Cache responses to prevent multiple download of pages
		// even if the collector is restarted
		colly.CacheDir("./.cansino_cache"),
	)

	c.OnHTML(a.HTMLSelector, a.HTMLProcessor)

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	err := c.Visit(a.ToURL())
	if err != nil {
		println(err)
	}

	return err
}

// ToURL formats the URL
func (a *Agenda) ToURL() string {
	return fmt.Sprintf(a.URL, a.Date.Day, a.Date.Month, a.Date.Year)
}

// AgendaDate represents a day
type AgendaDate struct {
	Day   int `yaml:"day"`
	Month int `yaml:"month"`
	Year  int `yaml:"year"`
}

//ToDate converts a date into time.Time
func (ad *AgendaDate) ToDate() time.Time {
	return time.Date(ad.Year, time.Month(ad.Month), ad.Day, 0, 0, 0, 0, time.UTC)
}

// AgendaEvent represents an event in the agenda
type AgendaEvent struct {
	Date        time.Time  `yaml:"date"`
	Description string     `yaml:"description"`
	Location    string     `yaml:"location"`
	Attendance  []Attendee `yaml:"attendance"`
}

// Attendee represents a person attending an event
type Attendee struct {
	Job      string `yaml:"job"`
	FullName string `yaml:"fullName"`
}
