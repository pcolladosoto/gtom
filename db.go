package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MONGODB_URI = "mongodb://mirrors.ft.uam.es:27017/"
	TELEGRAF_DB = "telegrafData"
)

type db struct {
	cli         *mongo.Client
	db          *mongo.Database
	collections map[string]*mongo.Collection
}

func NewDB() *db {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(MONGODB_URI))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		slog.Debug("error trying to ping the database", "err", err)
	}

	return &db{
		cli:         client,
		db:          client.Database(TELEGRAF_DB),
		collections: make(map[string]*mongo.Collection),
	}
}

func (d *db) Close() error {
	return d.cli.Disconnect(context.TODO())
}

func (d *db) find(collectionName, from, to string, filter []byte) ([]byte, error) {
	collection, ok := d.collections[collectionName]
	if !ok {
		slog.Debug("creating handle for collection", "collection", collectionName)
		collection = d.db.Collection(collectionName)
		d.collections[collectionName] = collection
	}

	var parsedFilter bson.D
	if err := bson.UnmarshalExtJSON(filter, true, &parsedFilter); err != nil {
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
