package internal

import (
	"context"
	"net/http"
)

type standard struct {
	certFile string
	keyFile  string
	server   *http.Server
}

func (s *standard) Start(ctx context.Context) error {
	// if we have a certFile and keyFile, use TLS
	if s.certFile != "" && s.keyFile != "" {
		return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	}
	return s.server.ListenAndServe()
}

func (s *standard) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func NewStandard(addr string, certFile string, keyFile string, h http.Handler) Startable {
	server := &http.Server{Addr: addr, Handler: h}
	return &standard{
		certFile: certFile,
		keyFile:  keyFile,
		server:   server,
	}
}
