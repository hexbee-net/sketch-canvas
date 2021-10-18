package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/apex/log"
	"github.com/gorilla/mux"
	"github.com/hexbee-net/sketch-canvas/pkg/datastore"
)

type Server struct {
	port         int
	srv          *http.Server
	router       *mux.Router
	storeOptions *datastore.Options
}

func New(port int, storeOptions *datastore.Options) *Server {
	s := Server{
		port:         port,
		storeOptions: storeOptions,
		router:       mux.NewRouter(),
	}

	s.router.HandleFunc("/", s.getVersions).Methods("GET")
	s.router.HandleFunc("/v1", s.getDocumentList).Methods("GET")
	s.router.HandleFunc("/v1/", s.createDocument).Methods("POST")
	s.router.HandleFunc("/v1/{id}", s.getDocuments).Methods("GET")
	s.router.HandleFunc("/v1/{id}", s.deleteDocument).Methods("DELETE")
	s.router.HandleFunc("/v1/{id}/rect", s.addRectangle).Methods("POST")
	s.router.HandleFunc("/v1/{id}/fill", s.addFloodFill).Methods("POST")

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", s.port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.router,
	}

	return &s
}

func (s *Server) Start(ctx context.Context) {
	if _, err := datastore.New(s.storeOptions, context.Background()); err != nil {
		log.Warnf("failed to connect to Redis instance at %s", s.storeOptions.Addr)
	}

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		log.Infof("Canvas server running on port %d", s.port)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("failed to start the http server")
		}
	}()

	// Block until we receive our signal.
	<-c

	log.Infof("Shutting down...")

	// Create a deadline to wait for.
	wait, _ := ctx.Value("wait").(time.Duration)
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
	log.Info("TODO: createDocument")
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
