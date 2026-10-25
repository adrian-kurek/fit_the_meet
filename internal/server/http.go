// Package server hold whole logic associated with server and router
package server

import (
	"context"
	"net/http"
	"time"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/internal/auth"
	"github.com/adrian-kurek/fit_the_meet/internal/event"
	"github.com/adrian-kurek/fit_the_meet/internal/group"
	"github.com/adrian-kurek/fit_the_meet/internal/poll"
)

const (
	defaultReadTimeout  = 50 * time.Second
	defaultWriteTimeout = 50 * time.Second
	defaultIdleTimeout  = 30 * time.Second
)

type eventHandler interface {
	Create(w http.ResponseWriter, r *http.Request) error
	CreateForGroup(w http.ResponseWriter, r *http.Request) error
	ToggleActive(w http.ResponseWriter, r *http.Request) error
	AddTimeBlocks(w http.ResponseWriter, r *http.Request) error
}

type pollHandler interface {
	Create(w http.ResponseWriter, r *http.Request) error
	SubmitAnswers(w http.ResponseWriter, r *http.Request) error
	ToggleActive(w http.ResponseWriter, r *http.Request) error
}

type authHandler interface {
	Register(w http.ResponseWriter, r *http.Request) error
	Login(w http.ResponseWriter, r *http.Request) error
	Verify(w http.ResponseWriter, r *http.Request) error
	RefreshAccessToken(w http.ResponseWriter, r *http.Request) error
	Logout(w http.ResponseWriter, r *http.Request) error
	LogoutFromAllDevices(w http.ResponseWriter, r *http.Request) error
}

type groupHandler interface {
	Create(w http.ResponseWriter, r *http.Request) error
	UpdateName(w http.ResponseWriter, r *http.Request) error
	Remove(w http.ResponseWriter, r *http.Request) error
	RemoveParticipants(w http.ResponseWriter, r *http.Request) error
	GenerateInviteCode(w http.ResponseWriter, r *http.Request) error
	Join(w http.ResponseWriter, r *http.Request) error
}

type DependencyConfig struct {
	port         string
	logger       commoninterfaces.Logger
	authHandler  authHandler
	groupHandler groupHandler
	eventHandler eventHandler
	pollHandler  pollHandler
}

func NewDependencyConfig(
	port string,
	logger commoninterfaces.Logger,
	authHandler authHandler,
	groupHandler groupHandler,
	eventHandler eventHandler,
	pollHandler pollHandler,
) *DependencyConfig {
	return &DependencyConfig{
		port:         port,
		logger:       logger,
		authHandler:  authHandler,
		groupHandler: groupHandler,
		eventHandler: eventHandler,
		pollHandler:  pollHandler,
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
	groupRoute := group.NewRoutes(h.config.groupHandler, h.config.logger)
	groupRoute.Setup(h.router)
	eventRoute := event.NewRoutes(h.config.eventHandler, h.config.logger)
	eventRoute.Setup(h.router)
	pollRoute := poll.NewRoutes(h.config.pollHandler, h.config.logger)
	pollRoute.Setup(h.router)
}

func (h *HTTP) Shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}
