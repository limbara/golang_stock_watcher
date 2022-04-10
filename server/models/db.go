package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var db *DB

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

type DbConfig struct {
	dbUri      string `validate:"required"`
	dbDatabase string `validate:"required"`
}

func NewDbConfig(uri string, database string) (*DbConfig, error) {
	validate := validator.New()

	config := &DbConfig{
		uri,
		database,
	}

	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("Error validate NewDbConfig : %w", err)
	}

	return config, nil
}

func (dc *DbConfig) URI() string {
	return dc.dbUri
}

func (dc *DbConfig) Database() string {
	return dc.dbDatabase
}

// Get DB stored in global variable. panic if nil
func GetDB() *DB {
	if db == nil {
		panic(errors.New("db was't BootstrapDB correctly"))
	}

	return db
}

// Bootstrap DB stored in global variable
func BootstrapDB(client *mongo.Client, dbConfig *DbConfig) {
	database := client.Database(dbConfig.Database())
	stockRepo := StockRepo{database.Collection("stocks")}

	repositories := make(map[string]Repo)
	repositories[stockRepo.RepoName()] = &stockRepo

	db = &DB{
		client,
		database,
		repositories,
	}
}

// Get MongoDb Client Connection
func InitMongoClient(dbConfig *DbConfig) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbConfig.URI()))
	if err != nil {
		err = fmt.Errorf("initMongoClient Error Connect Mongo : %w", err)
		return nil, err
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		err = fmt.Errorf("initMongoClient Error Ping Mongo : %w", err)
		return nil, err
	}

	return client, nil
}
