package regions

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	models "github.com/mdelapenya/cansino/models"
)

const cylEventsURL = "https://comunicacion.jcyl.es/web/jcyl/Comunicacion/es/PlantillaCalendarioBuscadorComponente/1284877983791/_/_/_?param[0]=%04d&param[1]=%02d&param[2]=%02d&parametro2=1281372093473&parametro3=1284233390583"

var juntaCYLStartDate = models.AgendaDate{
	Day: 21, Month: 11, Year: 2012,
}

// CYL returns the CYL region
func CYL() *models.Region {
	return &models.Region{
		Name:      "Castilla-León",
		DoPost:    false,
		StartDate: juntaCYLStartDate,
	}
}

// NewAgendaCYL represents the agenda for CYL
func NewAgendaCYL(region *models.Region, day int, month int, year int) *models.Agenda {
	agendaDate := models.AgendaDate{
		Day: day, Month: month, Year: year,
	}

	agendaURL := cylEventsURL
	cssSelector := "#contenidos"

	loc, _ := time.LoadLocation("Europe/Madrid")

	dateTime := time.Date(
		agendaDate.Year, time.Month(agendaDate.Month), agendaDate.Day,
		0, 0, 0, 0, loc,
	)

	agendaCYL := &models.Agenda{
		AllowedDomains: []string{"comunicacion.jcyl.es"},
		HTMLSelector:   cssSelector,
		HTMLProcessor:  cylProcessor,
		URLFormat:      agendaURL,
		Date:           dateTime,
		Day:            agendaDate,
		DoPost:         region.DoPost,
		Events:         []models.AgendaEvent{},
		ID:             "cyl-" + dateTime.Local().Format("2006-01-02"),
		Owner:          "Presidente",
		Region:         region.Name,
		URL:            fmt.Sprintf(agendaURL, agendaDate.Year, agendaDate.Month, agendaDate.Day),
	}

	return agendaCYL
}

func cylProcessor(a *models.Agenda, e *colly.HTMLElement) {
	e.ForEach("ul", func(index int, ul *colly.HTMLElement) {
		ul.ForEach("li.destacada a", func(index int, anchor *colly.HTMLElement) {
			var event = models.AgendaEvent{
				Attendance: []models.Attendee{},
				Owner:      a.Owner,
				Region:     a.Region,
			}

			anchor.ForEach("span", func(index int, span *colly.HTMLElement) {
				switch spanClass := span.Attr("class"); spanClass {
				case "fecha":
					hour := 0
					min := 0
					loc, _ := time.LoadLocation("Europe/Madrid")
					span.ForEach("span.hora", func(index int, timeSpan *colly.HTMLElement) {
						dateString := timeSpan.Text

						dateTime := strings.Split(dateString, ":")
						hour, _ = strconv.Atoi(dateTime[0])

						minString := dateTime[1]
						minString = strings.ReplaceAll(minString, "h", "")
						min, _ = strconv.Atoi(strings.Split(minString, " ")[0])
					})

					event.Date = time.Date(
						a.Day.Year, a.Day.ToDate().Month(), a.Day.Day,
						hour, min, 0, 0, loc,
					)
				case "subtitulo":
					event.Owner = span.Text
				case "lugar":
					event.OriginalLocation = span.Text
					event.Location = event.OriginalLocation
				case "descripcion":
					if event.OriginalDescription != "" {
						// avoid the issue when there are more than two description spans
						break
					}

					event.OriginalDescription = span.Text
					event.Description = event.OriginalDescription
				default:
					// NOOP
				}
			})

			event.ID = "cyl-" + event.Date.Local().Format("2006-01-02T15:04:05-0700")
			a.Events = append(a.Events, event)
		})
	})
}
