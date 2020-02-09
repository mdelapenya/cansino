package main

import (
	"context"
	"fmt"

	regions "github.com/mdelapenya/cansino/regions"
)

func main() {
	fmt.Println("Soy un cansino!")

	clm := regions.NewAgendaCLM()

	clm.Scrap(context.Background())
}

// Scrapper methods required to scrap a site
type Scrapper interface {
	Scrap(context.Context) error
}
