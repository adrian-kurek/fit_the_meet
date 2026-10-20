package group

import (
	"fmt"
	"net/http"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
)

type groupHandler interface {
	Create(w http.ResponseWriter, r *http.Request) error
	UpdateName(w http.ResponseWriter, r *http.Request) error
	Remove(w http.ResponseWriter, r *http.Request) error
}

type Routes struct {
	groupHandler groupHandler
	logger       commoninterfaces.Logger
}

func NewRoutes(groupHandler groupHandler, logger commoninterfaces.Logger) *Routes {
	return &Routes{
		groupHandler: groupHandler,
		logger:       logger,
	}
}

func (gr *Routes) Setup(router *http.ServeMux) {
	prefix := "/groups"

	router.Handle(fmt.Sprintf("POST %s/", prefix), request.Make(gr.groupHandler.Create, gr.logger))
	router.Handle(fmt.Sprintf("PATCH %s/{groupID}", prefix), request.Make(gr.groupHandler.UpdateName, gr.logger))
	router.Handle(fmt.Sprintf("DELETE %s/{groupID}", prefix), request.Make(gr.groupHandler.Remove, gr.logger))
}
