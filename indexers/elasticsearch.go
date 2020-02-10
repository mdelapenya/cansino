package indexers

import (
	"context"
	"fmt"

	models "github.com/mdelapenya/cansino/models"
)

// ElasticsearchIndexer represents an indexer for Elasticsearch
type ElasticsearchIndexer struct {
}

// NewESIndexer returns an Elasticsearch indexer
func NewESIndexer() Indexer {
	return &ElasticsearchIndexer{}
}

// Index indexes an agenda in Elasticsearch
func (ei *ElasticsearchIndexer) Index(ctx context.Context, agenda models.Agenda) error {
	fmt.Printf("Indexing agenda!")
	json, err := agenda.ToJSON()
	if err != nil {
		return err
	}

	fmt.Println(string(json))
	return nil
}
