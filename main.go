package main

import (
	"context"

	"github.com/mdelapenya/cansino/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Soy un cansino!")

	cmd.Execute()

	log.Info("Ya me he cansado 😴")
}

// Scrapper methods required to scrap a site
type Scrapper interface {
	Scrap(context.Context) error
}
