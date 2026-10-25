package poll

import (
	"fmt"
	"net/http"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
)

type pollHandler interface {
	Create(w http.ResponseWriter, r *http.Request) error
	SubmitAnswers(w http.ResponseWriter, r *http.Request) error
	ToggleActive(w http.ResponseWriter, r *http.Request) error
}

type Routes struct {
	pollHandler pollHandler
	logger      commoninterfaces.Logger
}

func NewRoutes(pollHandler pollHandler, logger commoninterfaces.Logger) *Routes {
	return &Routes{
		pollHandler: pollHandler,
		logger:      logger,
	}
}

func (pr *Routes) Setup(router *http.ServeMux) {
	prefix := "/polls"

	router.Handle(fmt.Sprintf("POST %s/", prefix), request.Make(pr.pollHandler.Create, pr.logger))
	router.Handle(
		fmt.Sprintf("POST %s/{pollID}/answers", prefix),
		request.Make(pr.pollHandler.SubmitAnswers, pr.logger),
	)

	router.Handle(
		fmt.Sprintf("PATCH %s/{pollID}", prefix),
		request.Make(pr.pollHandler.ToggleActive, pr.logger),
	)
}
