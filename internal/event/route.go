package event

import (
	"fmt"
	"net/http"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
)

type eventHandler interface {
	Create(w http.ResponseWriter, r *http.Request) error
	CreateForGroup(w http.ResponseWriter, r *http.Request) error
	ToggleActive(w http.ResponseWriter, r *http.Request) error
	AddTimeBlocks(w http.ResponseWriter, r *http.Request) error
	GetMany(w http.ResponseWriter, r *http.Request) error
}

type Routes struct {
	eventHandler eventHandler
	logger       commoninterfaces.Logger
}

func NewRoutes(eventHandler eventHandler, logger commoninterfaces.Logger) *Routes {
	return &Routes{
		eventHandler: eventHandler,
		logger:       logger,
	}
}

func (er *Routes) Setup(router *http.ServeMux) {
	prefix := "/events"

	router.Handle(
		fmt.Sprintf("POST /groups/{groupID}%s/", prefix),
		request.Make(er.eventHandler.CreateForGroup, er.logger),
	)
	router.Handle(fmt.Sprintf("POST %s/", prefix), request.Make(er.eventHandler.Create, er.logger))
	router.Handle(fmt.Sprintf("GET %s/", prefix), request.Make(er.eventHandler.GetMany, er.logger))
	router.Handle(fmt.Sprintf("PATCH %s/{eventID}", prefix), request.Make(er.eventHandler.ToggleActive, er.logger))
	router.Handle(fmt.Sprintf("PUT %s/{eventID}", prefix), request.Make(er.eventHandler.AddTimeBlocks, er.logger))
}
