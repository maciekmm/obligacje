package server

import (
	"log/slog"
	"net/http"

	"github.com/maciekmm/obligacje/bond"
	"github.com/maciekmm/obligacje/calculator"
)

type Server struct {
	repo    bond.Repository
	calc    *calculator.Calculator
	handler *http.ServeMux
	log     *slog.Logger
}

func NewServer(repo bond.Repository, logger *slog.Logger) *Server {
	server := &Server{
		repo:    repo,
		calc:    calculator.NewCalculator(),
		handler: http.NewServeMux(),
		log:     logger,
	}

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	s.handler.HandleFunc("GET /v1/bond/{name}/valuation", s.handleValuation)
	s.handler.HandleFunc("GET /v1/bond/{name}/historical", s.handleHistorical)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
