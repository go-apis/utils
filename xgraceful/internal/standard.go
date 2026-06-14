package internal

import (
	"context"
	"errors"
	"net/http"
)

type standard struct {
	certFile string
	keyFile  string
	server   *http.Server
}

func (s *standard) Start(ctx context.Context) error {
	// if we have a certFile and keyFile, use TLS
	var err error
	if s.certFile != "" && s.keyFile != "" {
		err = s.server.ListenAndServeTLS(s.certFile, s.keyFile)
	} else {
		err = s.server.ListenAndServe()
	}
	// a graceful Shutdown causes ListenAndServe to return ErrServerClosed;
	// that's the expected, successful path — not a failure to bubble up.
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
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
