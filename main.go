package main

import (
	"context"
	"fmt"

	regions "github.com/mdelapenya/cansino/regions"
)

func main() {
	fmt.Println("Soy un cansino!")

	err := regions.ProcessCLM(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Scrapper methods required to scrap a site
type Scrapper interface {
	Scrap(context.Context) error
}
