package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KirillShip/user-service/internal/config"
)

type Server struct {
	htpServer *http.Server
}

func NewServer(config *config.Config, handler http.Handler) *Server {
	return &Server{
		&http.Server{
			Addr:         fmt.Sprintf("%s:%s", config.Address, config.Port),
			Handler:      handler,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	log.Printf("Server is running on %s \n", s.htpServer.Addr)
	return s.htpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	log.Printf("Server is shutting down \n")
	return s.htpServer.Shutdown(ctx)
}
