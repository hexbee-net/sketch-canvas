package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/mux"
	"golang.org/x/xerrors"

	"github.com/hexbee-net/sketch-canvas/pkg/canvas"
	"github.com/hexbee-net/sketch-canvas/pkg/datastore"
	"github.com/hexbee-net/sketch-canvas/pkg/keygen"
)

type contextKey int

const (
	RequestIDContextKey contextKey = iota
	WaitContextKey      contextKey = iota
	DatastoreContextKey contextKey = iota
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
	keygen       keygen.KeyGen
}

func New(port int, storeOptions *datastore.RedisOptions) (*Server, error) {
	keyGen, err := keygen.New()
	if err != nil {
		return nil, xerrors.Errorf("failed to instantiate key generator: %w", err)
	}

	s := Server{
		port:         port,
		storeOptions: storeOptions,
		router:       mux.NewRouter(),
		keygen:       keyGen,
	}
	s.router.Use(MiddlewareRequestID)

	s.setupRoutes(s.router, MiddlewareRedisDatastore(s.storeOptions))

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", s.port),
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      s.router,
	}

	return &s, nil
}

func (s *Server) setupRoutes(router *mux.Router, datastoreMiddleware mux.MiddlewareFunc) {
	router.HandleFunc("/", s.getVersions).Methods("GET")

	v1 := router.PathPrefix("/v1/").Subrouter()
	v1.HandleFunc("/", s.getDocumentList).Methods(http.MethodGet)
	v1.HandleFunc("/", s.createDocument).Methods(http.MethodPost)
	v1.HandleFunc("/{id}", s.getDocuments).Methods(http.MethodGet)
	v1.HandleFunc("/{id}", s.deleteDocument).Methods(http.MethodDelete)
	v1.HandleFunc("/{id}/rect", s.addRectangle).Methods(http.MethodPost)
	v1.HandleFunc("/{id}/fill", s.addFloodFill).Methods(http.MethodPost)
	v1.Use(datastoreMiddleware)
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
	var (
		doc       canvas.Canvas
		requestID = s.getRequestID(r)
		reqLog    = log.
				WithField("operation-id", "create-doc").
				WithField("request-id", requestID)
	)

	reqLog.Debug("received create document request")

	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		reqLog.
			WithField("body", r.Body).
			WithError(err).
			Infof("failed to decode request body")
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	store := s.getStore(r)
	id := s.keygen.Generate()

	if err := store.SetDocument(id, &doc, r.Context()); err != nil {
		reqLog.
			WithError(err).
			Error("failed to set document in redis store")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	reqLog.
		WithField("doc-key", id).
		Infof("document created")

	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(path.Join(r.RequestURI, id))); err != nil {
		reqLog.
			WithField("doc-key", id).
			WithError(err).
			Error("failed to write http response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

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
	store, ok := r.Context().Value(DatastoreContextKey).(datastore.DataStore)
	if !ok {
		log.Fatal("could not find datastore in request context")
	}

	return store
}

func (s *Server) getRequestID(r *http.Request) int64 {
	id, ok := r.Context().Value(RequestIDContextKey).(int64)
	if !ok {
		log.Error("failed to retrieve request id")
	}

	return id
}
