package routes

import (
	"github.com/gorilla/mux"
	controllers "github.com/limbara/stock-watcher/controllers/api_v1"
)

func RegisterApiV1SectorRoutes(router *mux.Router) {
	router.HandleFunc("", controllers.GetSectors)
	router.HandleFunc("/", controllers.GetSectors)
	router.HandleFunc("/{id}", controllers.GetSector)
}
