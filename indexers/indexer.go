package indexers

import (
	"context"
	"errors"

	"github.com/mdelapenya/cansino/models"
)

// Indexer methods required to index a site
type Indexer interface {
	Index(context.Context, models.AgendaEvent) error
}

// GetIndexer returns the indexer by name
func GetIndexer(name string) (Indexer, error) {
	if name == "elasticsearch" {
		return NewESIndexer(), nil
	}

	return nil, errors.New("indexer " + name + " not found")
}
