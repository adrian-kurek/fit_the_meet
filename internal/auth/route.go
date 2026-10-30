package auth

import (
	"fmt"
	"net/http"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
)

type authHandler interface {
	Register(w http.ResponseWriter, r *http.Request) error
	Login(w http.ResponseWriter, r *http.Request) error
	Verify(w http.ResponseWriter, r *http.Request) error
	RefreshAccessToken(w http.ResponseWriter, r *http.Request) error
	Logout(w http.ResponseWriter, r *http.Request) error
	LogoutFromAllDevices(w http.ResponseWriter, r *http.Request) error
	RequestPasswordReset(w http.ResponseWriter, r *http.Request) error
	ResetPassword(w http.ResponseWriter, r *http.Request) error
}

type Routes struct {
	authHandler authHandler
	logger      commoninterfaces.Logger
}

func NewRoutes(authHandler authHandler, logger commoninterfaces.Logger) *Routes {
	return &Routes{
		authHandler: authHandler,
		logger:      logger,
	}
}

func (ar *Routes) Setup(router *http.ServeMux) {
	prefix := "/auth"
	router.Handle(fmt.Sprintf("POST %s/register", prefix), request.Make(ar.authHandler.Register, ar.logger))
	router.Handle(fmt.Sprintf("POST %s/login", prefix), request.Make(ar.authHandler.Login, ar.logger))
	router.Handle(fmt.Sprintf("GET %s/verify", prefix), request.Make(ar.authHandler.Verify, ar.logger))
	router.Handle(fmt.Sprintf("POST %s/refresh", prefix), request.Make(ar.authHandler.RefreshAccessToken, ar.logger))
	router.Handle(
		fmt.Sprintf("POST %s/password-reset/request", prefix),
		request.Make(ar.authHandler.RequestPasswordReset, ar.logger),
	)
	router.Handle(
		fmt.Sprintf("POST %s/password-reset/confirm", prefix),
		request.Make(ar.authHandler.ResetPassword, ar.logger),
	)
	router.Handle(fmt.Sprintf("DELETE %s/logout", prefix), request.Make(ar.authHandler.Logout, ar.logger))
	router.Handle(
		fmt.Sprintf("DELETE %s/logoutFromAll", prefix),
		request.Make(ar.authHandler.LogoutFromAllDevices, ar.logger),
	)
}
