package main

import (
	"context"
	"fmt"

	regions "github.com/mdelapenya/cansino/regions"
)

func main() {
	fmt.Println("Soy un cansino!")

	clm := regions.NewAgendaCLM(7, 2, 2020)

	clm.Scrap(context.Background())

	json, err := clm.ToJSON()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(json))
}

// Scrapper methods required to scrap a site
type Scrapper interface {
	Scrap(context.Context) error
}
