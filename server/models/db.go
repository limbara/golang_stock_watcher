package models

import (
	"context"
	"fmt"

	"github.com/limbara/stock-watcher/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Db *DB

type Repo interface {
	RepoName() string
	Collection() *mongo.Collection
}

type DB struct {
	Client       *mongo.Client
	Database     *mongo.Database
	repositories map[string]Repo
}

func (db *DB) GetRepo(name string) Repo {
	return db.repositories[name]
}

func BootstrapDB(client *mongo.Client) error {
	env, err := utils.LoadAppEnv()
	if err != nil {
		err = fmt.Errorf("BootstrapDB Error Loading Env  : %w", err)
		return err
	}

	database := client.Database(env.MongodbDatabase)
	stockRepo := StockRepo{database.Collection("stocks")}

	repositories := make(map[string]Repo)
	repositories[stockRepo.RepoName()] = &stockRepo

	Db = &DB{
		client,
		database,
		repositories,
	}

	return nil
}

func InitMongoClient() (*mongo.Client, error) {
	env, err := utils.LoadAppEnv()
	if err != nil {
		err = fmt.Errorf("initMongoClient Error Loading Env  : %w", err)
		return nil, err
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(env.MongodbUrl))
	if err != nil {
		err = fmt.Errorf("initMongoClient Error Connect Mongo : %w", err)
		return nil, err
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		err = fmt.Errorf("initMongoClient Error Ping Mongo : %w", err)
		return client, err
	}

	return client, nil
}

func DisconnectMongoClient(client *mongo.Client) error {
	return Db.Client.Disconnect(context.TODO())
}
