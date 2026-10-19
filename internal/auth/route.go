package auth

import (
	"fmt"
	"net/http"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
)

type authHandler interface {
	Register(w http.ResponseWriter, r *http.Request) error
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
}
