package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/limbara/stock-watcher/middlewares"
	"github.com/limbara/stock-watcher/models"
	"github.com/limbara/stock-watcher/routes"
	"github.com/limbara/stock-watcher/utils"
	"github.com/spf13/viper"
)

func main() {
	// Initialize env
	utils.BootstrapEnv()

	// Set time zone location globally
	location, err := time.LoadLocation(viper.GetString("APP_TZ"))
	if err == nil {
		time.Local = location
	}

	err = utils.BootstrapLogger(viper.GetString("LOG_PATH"))
	if err != nil {
		log.Fatalf("Error Get Logger :\n %+v", err)
	}
	logger := utils.Logger()

	dbConfig, err := models.NewDbConfig(viper.GetString("MONGODB_URI"), viper.GetString("MONGODB_DATABASE"))
	if err != nil {
		logger.Sugar().Fatalf("Error NewDbConfig:\n%+v", err)
	}

	client, err := models.InitMongoClient(dbConfig)
	if err != nil {
		logger.Sugar().Fatalw("InitMongoClient Fatal Error", "error", err)
	}
	models.BootstrapDB(client, dbConfig)

	RegisterCrons()

	router := createRouter()

	addr := fmt.Sprintf("%s:%s", viper.GetString("APP_HOST"), viper.GetString("APP_PORT"))
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
