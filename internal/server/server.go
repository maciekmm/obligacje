package server

import (
	"net/http"
	"time"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/calculator"
)

type Server struct {
	repo    bond.Repository
	calc    *calculator.Calculator
	handler *http.ServeMux
}

func NewServer(repo bond.Repository) *Server {
	server := &Server{
		repo:    repo,
		calc:    calculator.NewCalculator(),
		handler: http.NewServeMux(),
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	s.handler.HandleFunc("GET /v1/bond/{name}/valuation", s.handleValuation)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
