package objectstorage

import (
	"context"
	"errors"
	"net/http"

	"github.com/ZeleniJure/object-storage/server"
	"github.com/ZeleniJure/object-storage/storagebackend"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type ObjectStorage struct {
	ctx    context.Context
	backend *storagebackend.S3Storage
}

const apiPrefix = "/object/"

func New(s server.Server) {
	backend := storagebackend.New()
	o := &ObjectStorage{s.Ctx, backend}
	o.registerRoutes(s.Routes)
}

func (o *ObjectStorage) registerRoutes(m *mux.Router) {
	m.PathPrefix(apiPrefix + "{id}").HandlerFunc(o.getId).Methods(http.MethodGet)
	m.PathPrefix(apiPrefix + "{id}").HandlerFunc(o.putId).Methods(http.MethodPut)
}

func (o *ObjectStorage) getId(w http.ResponseWriter, r *http.Request) {
	// TODO leaking connections because body is not read. Would need to check
	// lib if this is still true...
	id, err := o.parseId(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Info().Str("id", id).Msg("Get ID")
	object, err := o.backend.Get(o.ctx, id)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(object.Bytes())
}

func (o *ObjectStorage) putId(w http.ResponseWriter, r *http.Request) {
	// TODO leaking connections because body is not read. Would need to check
	// lib if this is still true...
	id, err := o.parseId(r)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Info().Str("id", id).Msg("Get ID")
	err = o.backend.Put(id, r.Body)
	if err != nil {
		server.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	server.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (o *ObjectStorage) parseId(r *http.Request) (string, error) {
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		return "", errors.New("Object ID missing")
	}

	return id, nil
}
