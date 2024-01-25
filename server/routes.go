package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const routeInfo = "/info"

func NewRouter() *mux.Router {
	m := mux.NewRouter()
	m.HandleFunc(routeInfo, infoHandler).Methods(http.MethodGet)

	return m
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.Write([]byte("This is an awsome Objest Storage Gateway!"))
	log.Info().Msg("Landed on info route")
}
