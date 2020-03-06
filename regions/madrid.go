package regions

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	models "github.com/mdelapenya/cansino/models"
)

const madridClass = "special-section"

const madridCurrentEventsURL = "https://www.comunidad.madrid/views/ajax"

var madridCurrentStartDate = models.AgendaDate{
	Day: 19, Month: 8, Year: 2019,
}

// Madrid returns the Madrid region
func Madrid() *models.Region {
	return &models.Region{
		Name:      "Madrid",
		DoPost:    true,
		StartDate: madridCurrentStartDate,
	}
}

// NewAgendaMadrid represents the agenda for Madrid
func NewAgendaMadrid(region *models.Region, day int, month int, year int) *models.Agenda {
	agendaDate := models.AgendaDate{
		Day: day, Month: month, Year: year,
	}

	agendaURL := madridCurrentEventsURL
	cssSelector := "div.view-agenda div div ul"

	loc, _ := time.LoadLocation("Europe/Madrid")

	dateTime := time.Date(
		agendaDate.Year, time.Month(agendaDate.Month), agendaDate.Day,
		0, 0, 0, 0, loc,
	)

	agendaMadrid := &models.Agenda{
		AllowedDomains: []string{"www.comunidad.madrid"},
		HTMLSelector:   cssSelector,
		JSONProcessor:  madridProcessor,
		URLFormat:      agendaURL,
		Date:           dateTime,
		Day:            agendaDate,
		DoPost:         region.DoPost,
		Payload:        `field_date_value[value][date]=` + dateTime.Local().Format("02/01/2006") + `&field_date_value2[value][date]=` + dateTime.Local().Format("02/01/2006") + `&view_name=goverment_agenda&view_display_id=goverment_agenda_block`,
		Events:         []models.AgendaEvent{},
		ID:             "madrid-" + dateTime.Local().Format("2006-01-02"),
		Owner:          "Presidenta",
		Region:         region.Name,
		URL:            agendaURL,
	}

	return agendaMadrid
}

type madridResult struct {
	Data map[string]string `json:"data"`
}

func madridProcessor(a *models.Agenda, body []byte) {
	var response []map[string]string
	err := json.Unmarshal(body, &response)
	if err != nil {
		//
	}

	result := response[1]
	data := result["data"]
	if strings.Contains(data, "no existen eventos programados en el día seleccionado") {
		return
	}

	doc, err := htmlquery.Parse(strings.NewReader(data))
	htmlEvents := htmlquery.Find(doc, "//div[@about]")
	for _, htmlEvent := range htmlEvents {
		ownerDiv := htmlquery.FindOne(htmlEvent, "//div[contains(@class, 'field-name-field-counselings')]")
		owner := htmlquery.InnerText(ownerDiv)

		if owner != "La Presidenta" {
			continue
		}

		a.Owner = owner

		event := models.AgendaEvent{
			Attendance: []models.Attendee{},
			Owner:      a.Owner,
			Region:     a.Region,
		}

		dateDiv := htmlquery.FindOne(htmlEvent, "//div[contains(@class, 'field-type-date')]")
		dateString := htmlquery.InnerText(dateDiv)
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

		titleDiv := htmlquery.FindOne(htmlEvent, "//div[contains(@class, 'field-name-title')]")
		title := htmlquery.InnerText(titleDiv)

		descriptionDiv := htmlquery.FindOne(htmlEvent, "//div[contains(@class, 'field-name-field-short-description')]")
		description := htmlquery.InnerText(descriptionDiv)

		event.Description = title + " - " + description
		event.OriginalDescription = event.Description

		locationDiv := htmlquery.FindOne(htmlEvent, "//div[contains(@class, 'field-name-field-place')]")
		location := ""
		if locationDiv != nil {
			location = htmlquery.InnerText(locationDiv)
			location = strings.ReplaceAll(location, "Lugar: ", "")
		} else {
			locationDiv = htmlquery.FindOne(htmlEvent, "//div[contains(@class, 'field-name-field-location-address')]")
			if locationDiv != nil {
				location = htmlquery.InnerText(locationDiv)
				location = strings.ReplaceAll(location, "Direccion: ", "")
			}
		}
		event.Location = location
		event.OriginalLocation = event.Location

		event.ID = "madrid-" + event.Date.Local().Format("2006-01-02T15:04:05-0700")
		a.Events = append(a.Events, event)
	}
}
