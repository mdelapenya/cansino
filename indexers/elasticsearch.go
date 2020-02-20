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
	apm "go.elastic.co/apm"
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
func (ei *ElasticsearchIndexer) Index(ctx context.Context, event models.AgendaEvent) error {
	eventJSON, err := event.ToJSON()
	if err != nil {
		return err
	}

	esClient, err := getElasticsearchClient()
	if err != nil {
		return err
	}

	// Set up the APM transaction
	txn := apm.DefaultTracer.StartTransaction("Index()", "indexing")
	// Add current user to the transaction metadata
	txn.Context.SetUsername("cansino")
	// Store the transaction in a context
	txCtx := apm.ContextWithTransaction(ctx, txn)
	// Mark the transaction as completed
	defer txn.End()

	// Set up the request object.
	req := esapi.IndexRequest{
		Index:      "cansino",
		DocumentID: event.ID,
		Body:       strings.NewReader(string(eventJSON)),
		Refresh:    "true",
	}

	// Perform the request with the client.
	res, err := req.Do(txCtx, esClient)
	if err != nil {
		// Capture the error
		apm.CaptureError(txCtx, err).Send()

		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%s", res.Status(), event.ID)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			// Capture the error
			apm.CaptureError(txCtx, err).Send()
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Set the response status as transaction result
			txn.Result = res.Status()

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
