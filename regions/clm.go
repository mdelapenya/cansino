package regions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	models "github.com/mdelapenya/cansino/models"
)

const clmClass = "agenda evento"

const clmCurrentEventsURL = "https://transparencia.castillalamancha.es/agenda/198?date_filter[value][date]=%02d/%02d/%04d"

const clmPastEventsURL = "https://transparencia.castillalamancha.es/agenda-historico/198?date_filter[value][date]=%02d/%02d/%04d"

var clmCurrentStartDate = models.AgendaDate{
	Day: 8, Month: 7, Year: 2019,
}

var clmHistoricalStartDate = models.AgendaDate{
	Day: 1, Month: 2, Year: 2017,
}

var clmHistoricalEndDate = models.AgendaDate{
	Day: 7, Month: 7, Year: 2019,
}

// CLM returns the CLM region
func CLM() *models.Region {
	return &models.Region{
		Name:      "Castilla-La Mancha",
		DoPost:    false,
		StartDate: clmHistoricalStartDate,
	}
}

// NewAgendaCLM represents the agenda for Castilla-la Mancha
func NewAgendaCLM(region *models.Region, day int, month int, year int) *models.Agenda {
	agendaDate := models.AgendaDate{
		Day: day, Month: month, Year: year,
	}

	agendaURL := clmPastEventsURL
	cssSelector := "div.agenda-historico div div ul"
	if agendaDate.ToDate().After(clmHistoricalEndDate.ToDate()) {
		agendaURL = clmCurrentEventsURL
		cssSelector = "div.view-agenda div div ul"
	}

	loc, _ := time.LoadLocation("Europe/Madrid")

	dateTime := time.Date(
		agendaDate.Year, time.Month(agendaDate.Month), agendaDate.Day,
		0, 0, 0, 0, loc,
	)

	agendaCLM := &models.Agenda{
		AllowedDomains: []string{"transparencia.castillalamancha.es"},
		HTMLSelector:   cssSelector,
		HTMLProcessor:  clmProcessor,
		URLFormat:      agendaURL,
		Date:           dateTime,
		Day:            agendaDate,
		DoPost:         region.DoPost,
		Events:         []models.AgendaEvent{},
		ID:             "clm-" + dateTime.Local().Format("2006-01-02"),
		Owner:          "Presidente",
		Region:         region.Name,
		URL:            fmt.Sprintf(agendaURL, agendaDate.Day, agendaDate.Month, agendaDate.Year),
	}

	return agendaCLM
}

func clmProcessor(a *models.Agenda, e *colly.HTMLElement) {
	if strings.Contains(strings.TrimSpace(e.Attr("class")), clmClass) {
		var event models.AgendaEvent
		e.ForEach("li", func(index int, li *colly.HTMLElement) {
			if "cargo" == strings.TrimSpace(li.Attr("class")) {
				event = models.AgendaEvent{
					Attendance: []models.Attendee{},
					Owner:      a.Owner,
					Region:     a.Region,
				}
			} else if index == 1 {
				description := li.Text
				firstHyphen := strings.Index(description, "-")

				descRunes := []rune(description)
				dateString := string(descRunes[0:firstHyphen])

				dateTime := strings.Split(dateString, ":")
				hour, err := strconv.Atoi(dateTime[0])
				if err != nil {
					hour = 0
				}
				minString := dateTime[1]
				min, err := strconv.Atoi(strings.Split(minString, " ")[0])
				if err != nil {
					min = 0
				}
				loc, _ := time.LoadLocation("Europe/Madrid")

				event.Date = time.Date(
					a.Day.Year, a.Day.ToDate().Month(), a.Day.Day,
					hour, min, 0, 0, loc,
				)
				event.Description = strings.TrimSpace(string(descRunes[firstHyphen+1:]))
				event.OriginalDescription = event.Description
			} else if index == 2 {
				location := li.Text
				arr := strings.Split(location, "Lugar:")
				if len(arr) == 2 {
					event.Location = strings.TrimSpace(arr[1])
				} else {
					event.Location = strings.TrimSpace(location)
				}
				event.OriginalLocation = event.Location
			} else if "ver-mas" == strings.TrimSpace(li.Attr("class")) {
				li.ForEach("ul div div div p", func(_ int, p *colly.HTMLElement) {
					html, err := p.DOM.Html()
					if err != nil {
						html = ""
					}
					attendance := strings.Split(html, "<br/>")
					for _, a := range attendance {
						if a == "" {
							continue
						}
						line := strings.Split(a, " - ")
						var attendee models.Attendee
						if len(line) == 2 {
							attendee = models.Attendee{
								Job:      line[0],
								FullName: line[1],
							}
						} else {
							attendee = models.Attendee{
								Job: line[0],
							}
						}
						event.Attendance = append(event.Attendance, attendee)
					}
				})
			} else {
				// discard LI
			}
		})
		event.ID = "clm-" + event.Date.Local().Format("2006-01-02T15:04:05-0700")
		a.Events = append(a.Events, event)
	}
}
