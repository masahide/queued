package queued

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	Config *Config
	Router *mux.Router
	Store  Store
	App    *Application
	Addr   string
}

func NewServer(config *Config) *Server {
	router := mux.NewRouter()
	store := config.CreateStore()
	app := NewApplication(store)
	addr := fmt.Sprintf(":%d", config.Port)

	s := &Server{config, router, store, app, addr}

	s.HandleFunc("/", s.ListQueuesHandler).Methods("GET")
	s.HandleFunc("/", s.CreateQueueHandler).Methods("POST")
	s.HandleFunc("/{queue}", s.EnqueueHandler).Methods("POST")
	s.HandleFunc("/{queue}", s.StatsHandler).Methods("GET")
	s.HandleFunc("/{queue}/dequeue", s.DequeueHandler).Methods("POST")
	s.HandleFunc("/{queue}/{id}", s.InfoHandler).Methods("GET")
	s.HandleFunc("/{queue}/{id}", s.CompleteHandler).Methods("DELETE")

	return s
}

func (s *Server) HandleFunc(route string, fn http.HandlerFunc) *mux.Route {
	return s.Router.Handle(route, auth(s.Config, fn))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.Router.ServeHTTP(w, req)
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.Addr)

	if err != nil {
		return err
	}

	srv := http.Server{Handler: s}
	go srv.Serve(listener)

	return nil
}
