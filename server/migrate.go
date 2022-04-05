package main

import (
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/limbara/stock-watcher/utils"
	"go.mongodb.org/mongo-driver/mongo"
)

func Migrate(client *mongo.Client) error {
	appEnv, err := utils.LoadAppEnv()
	if err != nil {
		return err
	}

	driver, err := mongodb.WithInstance(client, &mongodb.Config{
		DatabaseName: appEnv.MongodbDatabase,
	})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mongodb", driver)
	m.Up()
	if err != nil {
		return err
	}

	return nil
}
