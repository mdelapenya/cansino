package main

import (
	"context"
	"fmt"

	indexers "github.com/mdelapenya/cansino/indexers"
	regions "github.com/mdelapenya/cansino/regions"
)

func main() {
	fmt.Println("Soy un cansino!")

	clm := regions.NewAgendaCLM(7, 2, 2020)

	clm.Scrap(context.Background())

	indexer, _ := indexers.GetIndexer("elasticsearch")
	err := indexer.Index(context.Background(), *clm)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// Scrapper methods required to scrap a site
type Scrapper interface {
	Scrap(context.Context) error
}
