package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/limbara/stock-watcher/middlewares"
	"github.com/limbara/stock-watcher/models"
	"github.com/limbara/stock-watcher/routes"
	"github.com/limbara/stock-watcher/utils"
)

func main() {
	logger, err := utils.Logger()
	defer logger.Sync()
	if err != nil {
		log.Fatalf("Error Get Logger :\n %+v", err)
	}

	client, err := models.InitMongoClient()
	if err != nil {
		logger.Sugar().Fatalw("InitMongoClient Fatal Error", "error", err)
	}
	defer models.DisconnectMongoClient(client)
	logger.Sugar().Infow("Init Mongo Success")

	if err := models.BootstrapDB(client); err != nil {
		logger.Sugar().Fatalw("BootstrapDB Fatal Error", "error", err)
	}

	if err := Migrate(client); err != nil {
		logger.Sugar().Fatalw("Migrate Fatal Error", "error", err)
	}
	logger.Sugar().Infow("Migrations Success")

	router := mux.NewRouter()

	router.Use(middlewares.AddContextRequestIdMiddleware)
	router.Use(middlewares.AddContextWantJsonMiddleware)
	router.Use(middlewares.RequestLoggerMiddleware)
	router.Use(middlewares.ErrorHandlerMiddleware)

	routes.RegisterRoutes(router)

	logger.Sugar().Infow("Server Starts At http://localhost:8080")
	logger.Sugar().Errorw("Server Error", http.ListenAndServe(":8080", router))
}
