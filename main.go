package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	log "github.com/sirupsen/logrus"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/meilisearch/meilisearch-go"
)

var ctx = context.Background()

func main() {
	drop := flag.Bool("drop", false, "Drop the index (destination name)")
	meta := flag.Bool("meta", false, "Re-create the index metadata")
	only := flag.String("only", "", "Only synchronize one index by its name")
	configPath := flag.String("config", "config.yml", "Path to the config file")

	flag.Parse()

	var config Config
	if err := config.Parse(*configPath); err != nil {
		log.WithError(err).Fatalf("Failed to parse config")
	}

	ms := meilisearch.New(config.MeiliHost, meilisearch.WithAPIKey(config.MeiliKey))
	if !ms.IsHealthy() {
		log.Fatal("MeiliSearch is not healthy")
	}

	pg, err := pgxpool.New(ctx, config.Postgres)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Postgres")
	}

	for idx, index := range config.Indexes {
		if *only != "" && index.Destination != *only {
			log.Printf("Skipping index %s", index.Destination)
			continue
		}

		log.Printf("Synchronizing index %s", index.Destination)
		start := time.Now()

		if index.Cursor.Column != "" && *drop {
			log.WithFields(log.Fields{
				"index": index.Destination,
			}).Warn("Dropping index, cursor will be reset")
			index.Cursor.LastSync = time.Time{}
		}

		if index.Cursor.Column != "" && index.Cursor.LastSync.IsZero() {
			log.WithFields(log.Fields{
				"index": index.Destination,
			}).Warn("Cursor is at zero, indexing all documents")
		}

		indexMs, err := MakeIndex(index, *drop, *meta, ms)
		if err != nil {
			log.WithError(err).WithField("index", index.Destination).Error("Failed to make index")
			continue
		}

		if err := SyncIndex(ctx, 10000, pg, index, func(batch []map[string]interface{}) error {
			_, err := indexMs.AddDocuments(batch)
			if err != nil {
				return err
			}

			log.WithField("batch", len(batch)).Info("Added batch to index")

			return nil
		}); err != nil {
			log.WithError(err).WithField("index", index.Destination).Error("Failed to sync index")
			continue
		}

		if index.Cursor.Column != "" {
			config.Indexes[idx].Cursor.LastSync = time.Now()
			log.WithFields(log.Fields{
				"index": index.Destination,
				"time":  config.Indexes[idx].Cursor.LastSync,
			}).Info("Updated cursor")
		}

		log.WithFields(log.Fields{
			"index": index.Destination,
			"time":  time.Since(start),
		}).Info("Index synchronized")
	}

	if err := config.Save(*configPath); err != nil {
		log.WithError(err).Error("Failed to save config")
	}

	log.Println("Done indexing")
}

// MakeIndex creates a new index in MeiliSearch and returns the index manager.
func MakeIndex(config IndexConfig, drop bool, meta bool, ms meilisearch.ServiceManager) (meilisearch.IndexManager, error) {
	idx := ms.Index(config.Destination)

	if drop {
		if _, err := ms.DeleteIndex(config.Destination); err != nil {
			return nil, err
		}

		if _, err := ms.CreateIndex(&meilisearch.IndexConfig{
			Uid:        config.Destination,
			PrimaryKey: config.Primary,
		}); err != nil {
			return nil, err
		}

		log.WithField("index", config.Destination).Info("Created index")
	}

	if meta {
		if _, err := idx.UpdateFilterableAttributes(&config.Filterable); err != nil {
			return nil, err
		}

		if _, err := idx.UpdateSortableAttributes(&config.Sortable); err != nil {
			return nil, err
		}

		if _, err := idx.UpdateSearchableAttributes(&config.Searchable); err != nil {
			return nil, err
		}

		log.WithField("index", config.Destination).Info("Updated index metadata")
	}

	return idx, nil
}

// SyncIndex synchronizes the index in Postgres with the index in MeiliSearch.
func SyncIndex(ctx context.Context, batchSize int, pg *pgxpool.Pool, index IndexConfig, onBatch func([]map[string]interface{}) error) error {
	query := fmt.Sprintf("SELECT * FROM %s", index.Source)
	if index.Cursor.Column != "" {
		query += fmt.Sprintf(" WHERE %s > '%s'", index.Cursor.Column, index.Cursor.LastSync.Format(time.RFC3339))
	}

	offset := 0
	wg := sync.WaitGroup{}
	for {
		rows, err := pg.Query(ctx, fmt.Sprintf("%s LIMIT %d OFFSET %d", query, batchSize, offset))
		if err != nil {
			return err
		}

		batch := []map[string]interface{}{}
		for rows.Next() {
			var row map[string]interface{}
			if err := pgxscan.ScanRow(&row, rows); err != nil {
				return err
			}
			batch = append(batch, row)
		}

		if len(batch) == 0 {
			break
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := onBatch(batch); err != nil {
				log.WithError(err).Error("Failed to add batch to index")
			}
		}()

		offset += batchSize
	}

	wg.Wait()

	return nil
}