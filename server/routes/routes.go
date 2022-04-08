package routes

import (
	"github.com/gorilla/mux"
	controllers "github.com/limbara/stock-watcher/controllers"
	apiV1Router "github.com/limbara/stock-watcher/routes/api_v1"
)

func RegisterRoutes(router *mux.Router) {
	router.Path("/").HandlerFunc(controllers.GetIndex)

	apiV1Router.RegisterApiV1Routes(router.PathPrefix("/api/v1").Subrouter())
}
