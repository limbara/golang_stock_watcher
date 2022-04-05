package routes

import (
	"github.com/gorilla/mux"
	apiV1Router "github.com/limbara/stock-watcher/routes/api_v1"
)

func RegisterRoutes(router *mux.Router) {
	apiV1Router.RegisterApiV1Routes(router.PathPrefix("/api/v1").Subrouter())
}
