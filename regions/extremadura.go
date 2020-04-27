package regions

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	models "github.com/mdelapenya/cansino/models"
)

const juntaExtremaduraEventsURL = "http://www.juntaex.es/web/agenda-presidencia?year=%04d&month=%02d&day=%02d"

var juntaExtremaduraStartDate = models.AgendaDate{
	Day: 1, Month: 3, Year: 2012,
}

// Extremadura returns the Extremadura region
func Extremadura() *models.Region {
	return &models.Region{
		Name:      "Extremadura",
		DoPost:    false,
		StartDate: juntaExtremaduraStartDate,
	}
}

// NewAgendaExtremadura represents the agenda for Extremadura
func NewAgendaExtremadura(region *models.Region, day int, month int, year int) *models.Agenda {
	agendaDate := models.AgendaDate{
		Day: day, Month: month, Year: year,
	}

	agendaURL := juntaExtremaduraEventsURL
	cssSelector := "#mainContent"

	loc, _ := time.LoadLocation("Europe/Madrid")

	dateTime := time.Date(
		agendaDate.Year, time.Month(agendaDate.Month), agendaDate.Day,
		0, 0, 0, 0, loc,
	)

	agendaExtremadura := &models.Agenda{
		AllowedDomains: []string{"www.juntaex.es"},
		HTMLSelector:   cssSelector,
		HTMLProcessor:  juntaExtremaduraProcessor,
		URLFormat:      agendaURL,
		Date:           dateTime,
		Day:            agendaDate,
		DoPost:         region.DoPost,
		Events:         []models.AgendaEvent{},
		ID:             "extremadura-" + dateTime.Local().Format("2006-01-02"),
		Owner:          "Presidente",
		Region:         region.Name,
		URL:            fmt.Sprintf(agendaURL, agendaDate.Year, agendaDate.Month, agendaDate.Day),
	}

	return agendaExtremadura
}

func juntaExtremaduraProcessor(a *models.Agenda, e *colly.HTMLElement) {
	e.ForEach("div", func(index int, mainDiv *colly.HTMLElement) {
		mainDiv.ForEach("blockquote", func(index int, blockquote *colly.HTMLElement) {
			var event = models.AgendaEvent{
				Attendance: []models.Attendee{},
				Owner:      a.Owner,
				Region:     a.Region,
			}

			processHeading := func(index int, headerDiv *colly.HTMLElement) {
				header := headerDiv.Text
				header = strings.ReplaceAll(header, "\t", "")
				header = strings.ReplaceAll(header, "\n", "")
				firstHyphen := strings.Index(header, "-")

				headerRunes := []rune(header)
				dateString := string(headerRunes[0:firstHyphen])

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
				event.Location = strings.TrimSpace(string(headerRunes[firstHyphen+1:]))
				event.OriginalLocation = event.Location
			}

			// in 2020-03-22, for soome reason the HTML markup changes
			blockquote.ForEach("div.eventHeading", func(index int, headerDiv *colly.HTMLElement) {
				processHeading(index, headerDiv)
			})
			blockquote.ForEach("p.eventHeading", func(index int, headerDiv *colly.HTMLElement) {
				processHeading(index, headerDiv)
			})

			blockquote.ForEach("div.eventShortDescription p", func(index int, descriptionDiv *colly.HTMLElement) {
				description := descriptionDiv.Text
				presidentIndex := strings.Index(description, "El presidente del Gobierno de Extremadura,")
				if presidentIndex == -1 {
					presidentIndex = strings.Index(description, "La presidenta del Gobierno de Extremadura,")
				}

				if presidentIndex >= 0 {
					re := regexp.MustCompile(`(^.*)((El|La) president[a|e] del Gobierno de Extremadura, .*?,)(.*$)`)

					matches := re.FindAllStringSubmatch(description, 4)
					if len(matches) == 1 && len(matches[0]) == 5 {
						description = strings.TrimSpace(matches[0][4])
					} else {
						description = description
					}
				}

				event.Description = description
				event.OriginalDescription = event.Description
			})

			event.ID = "extremadura-" + event.Date.Local().Format("2006-01-02T15:04:05-0700")
			a.Events = append(a.Events, event)
		})
	})
}
