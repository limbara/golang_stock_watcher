package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/limbara/stock-watcher/middlewares"
	"github.com/limbara/stock-watcher/models"
	"github.com/limbara/stock-watcher/routes"
	"github.com/limbara/stock-watcher/utils"
)

func main() {
	err := utils.BootstrapEnv()
	if err != nil {
		log.Fatalf("Error BootstrapEnv:\n %+v", err)
	}
	appEnv := utils.GetAppEnv()
	logPath, ok := utils.GetEnvOrDefault("LogPath", reflect.ValueOf("./storage/error")).Interface().(string)
	if !ok {
		log.Fatalf("Error GetEnvOrDefault assertion to string")
	}

	// Set time zone location globally
	location, err := time.LoadLocation(appEnv.AppTimezone)
	if err == nil {
		time.Local = location
	}

	err = utils.BootstrapLogger(logPath)
	if err != nil {
		log.Fatalf("Error Get Logger :\n %+v", err)
	}
	logger := utils.Logger()

	dbConfig, err := models.NewDbConfig(appEnv.DbUser, appEnv.DbPassword, appEnv.DbHost, appEnv.DbPort, appEnv.DbDatabase, appEnv.DbAuthSource)
	if err != nil {
		logger.Sugar().Fatalf("Error NewDbConfig:\n%+v", err)
	}

	client, err := models.InitMongoClient(dbConfig)
	if err != nil {
		logger.Sugar().Fatalw("InitMongoClient Fatal Error", "error", err)
	}
	models.BootstrapDB(client, dbConfig)

	if err := Migrate(client); err != nil {
		logger.Sugar().Fatalw("Migrate Fatal Error", "error", err)
	}
	logger.Sugar().Infow("Migrations Success")

	RegisterCrons()

	router := createRouter()

	addr := fmt.Sprintf(":%s", appEnv.AppPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger.Sugar().Infof("Server Starts At %s", addr)

	go func() {
		logger.Sugar().Fatal(server.ListenAndServe())
	}()

	<-done
	logger.Sugar().Infof("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		// Disconned Mongo
		client.Disconnect(ctx)

		// Flush all Log
		logger.Sync()

		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		logger.Sugar().Fatalf("Server Shutdown Failed:\n%+v", err)
	}

	logger.Sugar().Infof("Server Exited Properly")
}

func createRouter() *mux.Router {
	staticDir := http.Dir("./static")

	router := mux.NewRouter()
	// because if all route not match then all the registered middlewares won't be executed, have to manually add WantJson middleware
	router.NotFoundHandler = middlewares.AddContextWantJsonMiddleware(middlewares.RouteNotFoundHandlerMiddleware())

	router.Use(middlewares.AddContextRequestIdMiddleware)
	router.Use(middlewares.AddContextWantJsonMiddleware)
	router.Use(middlewares.RequestLoggerMiddleware)
	router.Use(middlewares.ErrorHandlerMiddleware)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(staticDir)))
	routes.RegisterRoutes(router)

	return router
}
