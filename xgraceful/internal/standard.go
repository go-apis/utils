package internal

import (
	"context"
	"net/http"
)

type standard struct {
	server *http.Server
}

func (s *standard) Start(ctx context.Context) error {
	// Run the server
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *standard) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewStandard(addr string, h http.Handler) Startable {
	server := &http.Server{Addr: addr, Handler: h}
	return &standard{server}
}
