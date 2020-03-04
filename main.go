package main

import (
	"context"

	"github.com/mdelapenya/cansino/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Println("Soy un cansino!")

	cmd.Execute()
}

// Scrapper methods required to scrap a site
type Scrapper interface {
	Scrap(context.Context) error
}
