// Package server hold whole logic associated with server and router
package server

import (
	"context"
	"net/http"
	"time"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/internal/auth"
)

const (
	defaultReadTimeout  = 50 * time.Second
	defaultWriteTimeout = 50 * time.Second
	defaultIdleTimeout  = 30 * time.Second
)

type authHandler interface {
	Register(w http.ResponseWriter, r *http.Request) error
	Login(w http.ResponseWriter, r *http.Request) error
	Verify(w http.ResponseWriter, r *http.Request) error
}

type DependencyConfig struct {
	port        string
	logger      commoninterfaces.Logger
	authHandler authHandler
}

func NewDependencyConfig(
	port string,
	logger commoninterfaces.Logger,
	authHandler authHandler,
) *DependencyConfig {
	return &DependencyConfig{
		port:        port,
		logger:      logger,
		authHandler: authHandler,
	}
}

type HTTP struct {
	config *DependencyConfig
	server *http.Server
	router *http.ServeMux
}

func NewHTTP(config *DependencyConfig) *HTTP {
	return &HTTP{
		config: config,
		router: http.NewServeMux(),
	}
}

func (h *HTTP) Start() error {
	h.SetupRoutes()
	h.server = &http.Server{
		Addr:         ":" + h.config.port,
		Handler:      h.router,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}
	return h.server.ListenAndServe()
}

func (h *HTTP) SetupRoutes() {
	authRoute := auth.NewRoutes(h.config.authHandler, h.config.logger)
	authRoute.Setup(h.router)
}

func (h *HTTP) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}
