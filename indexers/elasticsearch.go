package indexers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	es "github.com/elastic/go-elasticsearch/v7"
	esapi "github.com/elastic/go-elasticsearch/v7/esapi"
	models "github.com/mdelapenya/cansino/models"
	log "github.com/sirupsen/logrus"
	apmes "go.elastic.co/apm/module/apmelasticsearch"
)

var esInstance *es.Client

// ElasticsearchIndexer represents an indexer for Elasticsearch
type ElasticsearchIndexer struct {
}

// NewESIndexer returns an Elasticsearch indexer
func NewESIndexer() Indexer {
	return &ElasticsearchIndexer{}
}

// Index indexes an agenda in Elasticsearch
func (ei *ElasticsearchIndexer) Index(ctx context.Context, agenda models.Agenda) error {
	agendaJSON, err := agenda.ToJSON()
	if err != nil {
		return err
	}

	esClient, err := getElasticsearchClient()
	if err != nil {
		return err
	}

	// Set up the request object.
	req := esapi.IndexRequest{
		Index:      "cansino",
		DocumentID: agenda.ID,
		Body:       strings.NewReader(string(agendaJSON)),
		Refresh:    "true",
	}

	// Perform the request with the client.
	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%s", res.Status(), agenda.ID)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and indexed document version.
			fmt.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}

	return nil
}

// getElasticsearchClient returns a client connected to the running elasticseach cluster
func getElasticsearchClient() (*es.Client, error) {
	if esInstance != nil {
		return esInstance, nil
	}

	cfg := es.Config{
		CloudID:   os.Getenv("ELASTIC_CLOUD_ID"),
		Password:  os.Getenv("ELASTIC_CLOUD_AUTH"),
		Username:  os.Getenv("ELASTIC_CLOUD_USERNAME"),
		Transport: apmes.WrapRoundTripper(http.DefaultTransport),
	}
	esClient, err := es.NewClient(cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"config": cfg,
			"error":  err,
		}).Error("Could not obtain an Elasticsearch client")

		return nil, err
	}

	esInstance = esClient

	return esInstance, nil
}
