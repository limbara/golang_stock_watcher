package routes

import (
	"github.com/gorilla/mux"
	"github.com/limbara/stock-watcher/middlewares"
)

func RegisterApiV1Routes(router *mux.Router) {
	// forcing json response
	router.Use(middlewares.ForceContextWantJsonMiddleware)

	RegisterApiV1StockRoutes(router.PathPrefix("/stocks").Subrouter())
}
