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
)

func main() {
	appEnv, err := utils.LoadAppEnv()
	if err != nil {
		log.Fatalf("Error Load App Env:\n %+v", err)
	}
	logger, err := utils.Logger()
	if err != nil {
		log.Fatalf("Error Get Logger :\n %+v", err)
	}

	client, err := models.InitMongoClient()
	if err != nil {
		logger.Sugar().Fatalw("InitMongoClient Fatal Error", "error", err)
	}
	logger.Sugar().Infow("Init Mongo Success")

	if err := models.BootstrapDB(client); err != nil {
		logger.Sugar().Fatalw("BootstrapDB Fatal Error", "error", err)
	}

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
		logger.Sugar().Fatalf("Server ListenAndServe Error:\n%+v", server.ListenAndServe())
	}()

	<-done
	logger.Sugar().Infof("Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		// Disconned Mongo
		models.DisconnectMongoClient(client)

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
