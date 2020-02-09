package clm

import (
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
	models "github.com/mdelapenya/cansino/models"
)

const clmClass = "agenda evento"

const currentEventsURL = "https://transparencia.castillalamancha.es/agenda/198?date_filter[value][date]=%02d/%02d/%04d"

const pastEventsURL = "https://transparencia.castillalamancha.es/agenda-historico/198?date_filter[value][date]=%02d/%02d/%04d"

var currentStartDate = models.AgendaDate{
	Day: 8, Month: 7, Year: 2019,
}

var historicalStartDate = models.AgendaDate{
	Day: 1, Month: 2, Year: 2017,
}

var historicalEndDate = models.AgendaDate{
	Day: 7, Month: 7, Year: 2019,
}

// AgendaCLM holds the data for CLM's agenda
var AgendaCLM *models.Agenda

// NewAgendaCLM represents the agenda for Castilla-la Mancha
func NewAgendaCLM() *models.Agenda {
	AgendaCLM = &models.Agenda{
		AllowedDomains: []string{"transparencia.castillalamancha.es"},
		HTMLSelector:   "div.agenda-historico div div ul",
		HTMLProcessor:  clmProcessor,
		URL:            pastEventsURL,
		Date: models.AgendaDate{
			// example date
			Day: 1, Month: 7, Year: 2018,
		},
		Events: []models.AgendaEvent{},
		Owner:  "Presidente",
	}

	return AgendaCLM
}

func clmProcessor(e *colly.HTMLElement) {
	if clmClass == strings.TrimSpace(e.Attr("class")) {
		var event models.AgendaEvent
		e.ForEach("li", func(index int, li *colly.HTMLElement) {
			if "cargo" == strings.TrimSpace(li.Attr("class")) {
				event = models.AgendaEvent{
					Attendance: []models.Attendee{},
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
					AgendaCLM.Date.Year, AgendaCLM.Date.ToDate().Month(), AgendaCLM.Date.Day,
					hour, min, 0, 0, loc,
				)
				event.Description = strings.TrimSpace(string(descRunes[firstHyphen+1:]))
			} else if index == 2 {
				event.Location = li.Text
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
						attendee := models.Attendee{
							Job:      line[0],
							FullName: line[1],
						}
						event.Attendance = append(event.Attendance, attendee)
					}
				})
			} else {
				// discard LI
			}
		})
		AgendaCLM.Events = append(AgendaCLM.Events, event)
	}
}
