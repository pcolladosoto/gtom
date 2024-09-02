package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type db struct {
	cli         *mongo.Client
	db          *mongo.Database
	collections map[string]*mongo.Collection
}

func NewDB() (*db, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(lookupEnvDefault("GTOM_URI", "mongodb://localhost:2701")))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		slog.Debug("error trying to ping the database", "err", err)
	}

	return &db{
		cli:         client,
		db:          client.Database(lookupEnvDefault("GTOM_TELEGRAF_DB", "telegraf")),
		collections: make(map[string]*mongo.Collection),
	}, nil
}

func (d *db) Close() error {
	return d.cli.Disconnect(context.TODO())
}

func (d *db) find(collectionName, from, to, filter string) ([]byte, error) {
	collection, ok := d.collections[collectionName]
	if !ok {
		slog.Debug("creating handle for collection", "collection", collectionName)
		collection = d.db.Collection(collectionName)
		d.collections[collectionName] = collection
	}

	// Replace ' with " so that writing the filters isn't such a pain
	var parsedFilter bson.D
	if err := bson.UnmarshalExtJSON([]byte(strings.Replace(filter, "'", "\"", -1)), true, &parsedFilter); err != nil {
		return nil, err
	}

	fromParsed, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return nil, err
	}

	toParsed, err := time.Parse(time.RFC3339, to)
	if err != nil {
		return nil, err
	}

	fFilter := bson.D{
		{Key: "$and",
			Value: bson.A{
				bson.D{{Key: "timestamp", Value: bson.D{{Key: "$gte", Value: fromParsed}}}},
				bson.D{{Key: "timestamp", Value: bson.D{{Key: "$lte", Value: toParsed}}}},
			},
		},
	}

	for _, e := range parsedFilter {
		fFilter = append(fFilter, e)
	}

	slog.Debug("final filter", "filter", fFilter)

	cursor, err := collection.Find(context.TODO(), fFilter)
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	mResults, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return mResults, nil
}
