package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Server struct {
	port   int
	srv    *http.Server
	router *mux.Router
}

func New(port int) *Server {
	s := Server{
		port:   port,
		router: mux.NewRouter(),
	}

	s.router.HandleFunc("/", getVersions).Methods("GET")
	s.router.HandleFunc("/v1", getDocumentList).Methods("GET")
	s.router.HandleFunc("/v1/", createDocument).Methods("POST")
	s.router.HandleFunc("/v1/{id}", getDocuments).Methods("GET")
	s.router.HandleFunc("/v1/{id}", deleteDocument).Methods("DELETE")
	s.router.HandleFunc("/v1/{id}/rect", addRectangle).Methods("POST")
	s.router.HandleFunc("/v1/{id}/fill", addFloodFill).Methods("POST")

	s.srv = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", s.port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.router,
	}
	return &s
}

func (s *Server) Start(wait time.Duration) {
	go func() {
		log.Printf("Canvas server running on port %d", s.port)
		fmt.Println()
		if err := s.srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	log.Println("Shutting down...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait until the timeout deadline.
	_ = s.srv.Shutdown(ctx)
}

func getVersions(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: getVersions")
}

func getDocumentList(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: getDocumentList")
}

func createDocument(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: createDocument")
}

func getDocuments(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: getDocuments")
}

func deleteDocument(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: deleteDocument")
}

func addRectangle(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: addRectangle")
}

func addFloodFill(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: addFloodFill")
}
