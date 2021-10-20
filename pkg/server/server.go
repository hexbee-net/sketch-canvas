package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
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

const DefaultPageLimit = 10

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
	v1.HandleFunc("/docs/", s.getDocumentList).Methods(http.MethodGet)
	v1.HandleFunc("/docs/", s.createDocument).Methods(http.MethodPost)
	v1.HandleFunc("/docs/{id}", s.getDocument).Methods(http.MethodGet)
	v1.HandleFunc("/docs/{id}", s.deleteDocument).Methods(http.MethodDelete)
	v1.HandleFunc("/docs/{id}/rect", s.addRectangle).Methods(http.MethodPost)
	v1.HandleFunc("/docs/{id}/fill", s.addFloodFill).Methods(http.MethodPost)
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
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) getDocumentList(w http.ResponseWriter, r *http.Request) {
	var (
		url    = r.URL
		store  = s.getStore(r)
		reqLog = log.WithField("operation-id", "get-doc-list").WithField("request-id", s.getRequestID(r))
	)

	reqLog.Debug("received get document list request")

	// Parse query parameters
	req := struct {
		Cursor uint64 `schema:"q"`
		Limit  int64  `schema:"limit"`
	}{
		Cursor: 0,
		Limit:  DefaultPageLimit,
	}

	if err := r.ParseForm(); err != nil {
		reqLog.WithError(err).Warnf("invalid request: %s", r.RequestURI)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	if err := schema.NewDecoder().Decode(&req, r.Form); err != nil {
		reqLog.WithError(err).Warnf("invalid request: %s", r.RequestURI)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)

		return
	}

	reqLog.WithField("cursor", req.Cursor).WithField("limit", req.Limit).Debug("request parameter")

	// Retrieve a page of keys from the store.
	keys, cursor, err := store.GetDocList(req.Cursor, req.Limit, r.Context())
	if err != nil {
		reqLog.WithError(err).Error("failed to get document list from redis store")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	var next string

	// If the cursor is not 0, there are more values to retrieve.
	// Format the URI of the next page.
	if cursor != 0 {
		queryValues := url.Query()
		queryValues.Set("q", strconv.FormatUint(cursor, 10))       //nolint: gomnd
		queryValues.Set("limit", strconv.FormatInt(req.Limit, 10)) //nolint: gomnd
		url.RawQuery = queryValues.Encode()
		next = url.String()
	}

	// We don't fail if we're unable to retrieve the number of keys in the store
	// since it is not the most important thing in the query.
	dbSize, err := store.GetSize(r.Context())
	if err != nil {
		reqLog.WithError(err).Error("failed to retrieve number of keys in store")
	}

	// Convert the list of keys to a map with the URI of each document.
	//TODO: retrieve the document names instead of using the ID.
	docs := make(map[string]string, len(keys))
	for _, k := range keys {
		docs[k] = path.Join(url.Path, k)
	}

	data, err := jsonMarshal(struct {
		Next  string            `json:"next,omitempty"`
		Count int               `json:"count"`
		Total int64             `json:"total,omitempty"`
		Docs  map[string]string `json:"docs"`
	}{
		Next:  next,
		Count: len(keys),
		Total: dbSize,
		Docs:  docs,
	})
	if err != nil {
		reqLog.WithError(err).Error("failed to marshal response to json")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(data); err != nil {
		reqLog.WithError(err).Error("failed to write http response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (s *Server) createDocument(w http.ResponseWriter, r *http.Request) {
	var (
		doc    canvas.Canvas
		reqLog = log.WithField("operation-id", "create-doc").WithField("request-id", s.getRequestID(r))
		store  = s.getStore(r)
	)

	reqLog.Debug("received create document request")

	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		reqLog.WithField("body", r.Body).WithError(err).Infof("failed to decode request body")
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	id := s.keygen.Generate()

	if err := store.SetDocument(id, &doc, r.Context()); err != nil {
		reqLog.WithError(err).Error("failed to set document in redis store")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	reqLog.WithField("doc-key", id).Infof("document created")

	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write([]byte(path.Join(r.RequestURI, id))); err != nil {
		reqLog.WithField("doc-key", id).WithError(err).Error("failed to write http response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (s *Server) getDocument(w http.ResponseWriter, r *http.Request) {
	var (
		url    = r.URL.String()
		store  = s.getStore(r)
		vars   = mux.Vars(r)
		docID  = vars["id"]
		reqLog = log.
			WithField("operation-id", "get-doc").
			WithField("request-id", s.getRequestID(r)).
			WithField("doc-id", docID)
	)

	reqLog.Debug("received get document request")

	doc, err := store.GetDocument(docID, r.Context())
	if err != nil {
		switch err {
		case datastore.NotFound:
			reqLog.Info("document not found")
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		case err:
			reqLog.WithError(err).Error("failed to marshal response to json")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	data, err := jsonMarshal(struct {
		Operations map[string]string `json:"operations"`
		Canvas     *canvas.Canvas
	}{
		Operations: map[string]string{
			"delete-doc":     url,
			"add-rect":       path.Join(url, "rect"),
			"add-flood-fill": path.Join(url, "fill"),
		},
		Canvas: doc,
	})

	if err != nil {
		reqLog.WithError(err).Error("failed to marshal response to json")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err := w.Write(data); err != nil {
		reqLog.WithError(err).Error("failed to write http response")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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

// getStore retrieves the data store connection from the context.
func (s *Server) getStore(r *http.Request) datastore.DataStore {
	store, ok := r.Context().Value(DatastoreContextKey).(datastore.DataStore)
	if !ok {
		log.Fatal("could not find datastore in request context")
	}

	return store
}

// getRequestID retrieves the id of the current request from the context.
func (s *Server) getRequestID(r *http.Request) int64 {
	id, ok := r.Context().Value(RequestIDContextKey).(int64)
	if !ok {
		log.Error("failed to retrieve request id")
	}

	return id
}

// jsonMarshal returns the JSON encoding of v without encoding HTML characters.
func jsonMarshal(v interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)

	return buffer.Bytes(), err
}
