package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/apex/log"
	"github.com/bwmarrin/snowflake"
	"github.com/gorilla/mux"
	"golang.org/x/xerrors"

	"github.com/hexbee-net/sketch-canvas/pkg/canvas"
	"github.com/hexbee-net/sketch-canvas/pkg/datastore"
)

type contextKey int

const (
	WaitContextKey contextKey = iota
)

const (
	writeTimeout = time.Second * 15
	readTimeout  = time.Second * 15
	idleTimeout  = time.Second * 60
)

type Server struct {
	port         int
	srv          *http.Server
	router       *mux.Router
	storeOptions *datastore.RedisOptions
	node         *snowflake.Node
}

func New(port int, storeOptions *datastore.RedisOptions) (*Server, error) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, xerrors.Errorf("failed to instantiate SnowFlake node: %w", err)
	}

	s := Server{
		port:         port,
		storeOptions: storeOptions,
		router:       mux.NewRouter(),
		node:         node,
	}

	s.router.HandleFunc("/", s.getVersions).Methods("GET")

	v1 := s.router.PathPrefix("/v1/").Subrouter()
	v1.HandleFunc("/", s.getDocumentList).Methods(http.MethodGet)
	v1.HandleFunc("/", s.createDocument).Methods(http.MethodPost)
	v1.HandleFunc("/{id}", s.getDocuments).Methods(http.MethodGet)
	v1.HandleFunc("/{id}", s.deleteDocument).Methods(http.MethodDelete)
	v1.HandleFunc("/{id}/rect", s.addRectangle).Methods(http.MethodPost)
	v1.HandleFunc("/{id}/fill", s.addFloodFill).Methods(http.MethodPost)
	v1.Use(datastore.MiddlewareRedisDatastore(s.storeOptions))

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", s.port),
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      s.router,
	}

	return &s, nil
}

func (s *Server) Start(ctx context.Context) {
	if _, err := datastore.New(s.storeOptions, context.Background()); err != nil {
		log.Warnf("failed to connect to Redis instance at %s", s.storeOptions.Addr)
	}

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		log.Infof("canvas server running on port %d", s.port)

		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to start the http server")
		}
	}()

	// Block until we receive our signal.
	<-c

	log.Infof("Shutting down...")

	// Create a deadline to wait for.
	wait, _ := ctx.Value(WaitContextKey).(time.Duration)

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline.
	_ = s.srv.Shutdown(ctx)
}

func (s *Server) getVersions(w http.ResponseWriter, r *http.Request) {
	log.Info("TODO: getVersions")
}

func (s *Server) getDocumentList(w http.ResponseWriter, r *http.Request) {
	log.Info("TODO: getDocumentList")
}

func (s *Server) createDocument(w http.ResponseWriter, r *http.Request) {
	var doc canvas.Canvas

	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	store := s.getStore(r)
	id := s.node.Generate().String()
	log.
		WithField("doc-key", id).
		Debug("received create document request")

	if err := store.SetDocument(id, doc, r.Context()); err != nil {
		log.WithError(err).Error("failed to set document in redis store")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusCreated)

	if _, err := io.WriteString(w, path.Join(r.RequestURI, id)); err != nil {
		log.
			WithError(err).
			WithField("doc-key", id).
			Error("failed to write http response")

		return
	}
}

func (s *Server) getDocuments(w http.ResponseWriter, r *http.Request) {
	log.Info("TODO: getDocuments")
}

func (s *Server) deleteDocument(w http.ResponseWriter, r *http.Request) {
	log.Info("TODO: deleteDocument")
}

func (s *Server) addRectangle(w http.ResponseWriter, r *http.Request) {
	log.Info("TODO: addRectangle")
}

func (s *Server) addFloodFill(w http.ResponseWriter, r *http.Request) {
	log.Info("TODO: addFloodFill")
}

func (s *Server) getStore(r *http.Request) datastore.DataStore {
	store, ok := r.Context().Value(datastore.ContextKey).(datastore.DataStore)
	if !ok {
		log.Fatal("could not find datastore in request context")
	}

	return store
}
