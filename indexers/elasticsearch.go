package indexers

import (
	"context"
	"encoding/json"
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

type analysedToken struct {
	Token       string `json:"token"`
	StartOffset int    `json:"start_offset"`
	EndOffset   int    `json:"end_offset"`
	Word        string `json:"word"`
	Position    int    `json:"posistion"`
}

type analysedResponse map[string][]analysedToken

// NewESIndexer returns an Elasticsearch indexer
func NewESIndexer() Indexer {
	return &ElasticsearchIndexer{}
}

// analyze analyses a text using consino analyser to remove stop words
func analyze(ctx context.Context, text string) ([]string, error) {
	esClient, err := getElasticsearchClient()
	if err != nil {
		return []string{}, err
	}

	// Set up the APM transaction
	analyzeTxn := apm.DefaultTracer.StartTransaction("Analyze()", "analyze")
	// Add current user to the transaction metadata
	analyzeTxn.Context.SetUsername("cansino")
	// Store the transaction in a context
	analyzeTxCtx := apm.ContextWithTransaction(ctx, analyzeTxn)
	// Mark the transaction as completed
	defer analyzeTxn.End()

	body := getBody(text)
	// Perform the analyze request with the client.
	analyzeRes, err := esClient.Indices.Analyze(
		esClient.Indices.Analyze.WithIndex("cansino"),
		esClient.Indices.Analyze.WithBody(strings.NewReader(body)),
		esClient.Indices.Analyze.WithPretty(),
		esClient.Indices.Analyze.WithContext(analyzeTxCtx),
	)
	if err != nil {
		log.WithFields(log.Fields{
			"index": "cansino",
			"body":  body,
			"error": err,
		}).Error("Error getting analyze response")
		return []string{}, err
	}
	defer analyzeRes.Body.Close()

	// Deserialize the response into a map.
	r := analysedResponse{}
	if err := json.NewDecoder(analyzeRes.Body).Decode(&r); err != nil {
		// Capture the error
		apm.CaptureError(analyzeTxCtx, err).Send()
		log.WithFields(log.Fields{
			"body":    body,
			"reqBody": analyzeRes.Body,
			"error":   err,
		}).Error("Error parsing the analyze response body")
		return []string{}, err
	}

	// Set the response status as transaction result
	analyzeTxn.Result = analyzeRes.Status()

	tokens := r["tokens"]
	analysedTokens := make([]string, len(tokens))
	for i, token := range tokens {
		analysedTokens[i] = token.Token
	}

	return analysedTokens, nil
}

// Index indexes an agenda in Elasticsearch
func (ei *ElasticsearchIndexer) Index(ctx context.Context, event models.AgendaEvent) error {
	esClient, err := getElasticsearchClient()
	if err != nil {
		return err
	}

	// extract stop words from the description
	analysedDescription, err := analyze(ctx, event.Description)
	event.Description = strings.Join(analysedDescription, " ")

	// extract stop words from the location
	analysedLocation, err := analyze(ctx, event.Location)
	event.Location = strings.Join(analysedLocation, " ")

	// Set up the APM transaction
	txn := apm.DefaultTracer.StartTransaction("Index()", "indexing")
	// Add current user to the transaction metadata
	txn.Context.SetUsername("cansino")
	// Store the transaction in a context
	txCtx := apm.ContextWithTransaction(ctx, txn)
	// Mark the transaction as completed
	defer txn.End()

	eventJSON, err := event.ToJSON()
	if err != nil {
		return err
	}

	stringJSON := string(eventJSON)

	// Set up the request object.
	req := esapi.IndexRequest{
		Index:      "cansino",
		DocumentID: event.ID,
		Body:       strings.NewReader(stringJSON),
		Refresh:    "true",
	}

	// Perform the request with the client.
	res, err := req.Do(txCtx, esClient)
	if err != nil {
		// Capture the error
		apm.CaptureError(txCtx, err).Send()

		log.WithFields(log.Fields{
			"index":      "cansino",
			"documentID": event.ID,
			"json":       stringJSON,
			"error":      err,
		}).Error("Error getting response")
	}
	defer res.Body.Close()

	if res.IsError() {
		log.WithFields(log.Fields{
			"status":     res.Status(),
			"documentID": event.ID,
		}).Error("Error indexing document")
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			// Capture the error
			apm.CaptureError(txCtx, err).Send()
			log.WithFields(log.Fields{
				"error": err,
				"body":  res.Body,
			}).Error("Error parsing the response body")
		} else {
			// Set the response status as transaction result
			txn.Result = res.Status()

			// Print the response status and indexed document version.
			log.WithFields(log.Fields{
				"status":     res.Status(),
				"documentID": event.ID,
				"result":     r["result"],
				"version":    int(r["_version"].(float64)),
			}).Info("Document indexed")
		}
	}

	return nil
}

func getBody(text string) string {
	text = strings.ReplaceAll(text, `"`, `\"`)
	text = strings.ReplaceAll(text, "\n", "")

	return `{"analyzer": "spanish_stop", "text": "` + text + `"}`
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
