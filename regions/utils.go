package regions

import (
	"errors"
	"time"

	"github.com/mdelapenya/cansino/models"
)

// AgendaFactory returns an agenda object based on its name
func AgendaFactory(region *models.Region, day int, month int, year int) (*models.Agenda, error) {
	if region.Name == "Castilla-La Mancha" {
		return NewAgendaCLM(region, day, month, year), nil
	} else if region.Name == "Madrid" {
		return NewAgendaMadrid(region, day, month, year), nil
	}

	return &models.Agenda{}, errors.New("No such region")
}

// RegionFactory returns a region based on its name
func RegionFactory(name string) (*models.Region, error) {
	if name == "Castilla-La Mancha" {
		return CLM(), nil
	} else if name == "Madrid" {
		return Madrid(), nil
	}

	return &models.Region{}, errors.New("No such region")
}

// RangeDate returns a date range function over start date to end date inclusive.
// After the end of the range, the range function returns a zero date,
// date.IsZero() is true.
func RangeDate(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, start.Location())
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, start.Location())

	return func() time.Time {
		if start.After(end) {
			return time.Time{}
		}

		date := start
		start = start.AddDate(0, 0, 1)
		return date
	}
}
