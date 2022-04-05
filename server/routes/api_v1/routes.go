package routes

import (
	"github.com/gorilla/mux"
)

func RegisterApiV1Routes(router *mux.Router) {
	RegisterApiV1SectorRoutes(router.PathPrefix("/sectors").Subrouter())
}
