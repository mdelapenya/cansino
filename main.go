package main

import (
	"context"

	regions "github.com/mdelapenya/cansino/regions"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Println("Soy un cansino!")

	err := regions.ProcessCLM(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error processing Agenda CLM")
		return
	}
}

// Scrapper methods required to scrap a site
type Scrapper interface {
	Scrap(context.Context) error
}
