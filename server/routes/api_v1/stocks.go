package routes

import (
	"github.com/gorilla/mux"
	controllers "github.com/limbara/stock-watcher/controllers/api_v1"
)

func RegisterApiV1StockRoutes(router *mux.Router) {
	router.Path("").HandlerFunc(controllers.GetStocks)
	router.Path("/").HandlerFunc(controllers.GetStocks)
}
