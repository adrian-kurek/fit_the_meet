package event

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
	"github.com/adrian-kurek/fit_the_meet/common/response"
	"github.com/adrian-kurek/fit_the_meet/common/validation"
)

type eventService interface {
	Create(ctx context.Context, userID int, event Input) error
	CreateForGroup(ctx context.Context, userID, groupID int, event Input) error
	ToggleActive(ctx context.Context, userID int, eventID int) error
	AddTimeBlocks(ctx context.Context, userID int, eventID int, timeBlocks SubmitAvailabilityRequest) error
}

type authenticationMiddleware interface {
	VerifyToken(r *http.Request) (*http.Request, error)
}

const timeout = 5 * time.Second

type Handler struct {
	eventService             eventService
	authenticationMiddleware authenticationMiddleware
	logger                   commoninterfaces.Logger
}

func NewHandler(
	eventService eventService,
	authenticationMiddleware authenticationMiddleware,
	logger commoninterfaces.Logger,
) *Handler {
	return &Handler{
		eventService:             eventService,
		authenticationMiddleware: authenticationMiddleware,
		logger:                   logger,
	}
}

func (eh *Handler) handleTimeout(err error, path string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		eh.logger.Info("request timed out", "path", path)
		return commonerrors.RequestTimeout()
	}
	return err
}

func (eh *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	return eh.create(w, r, nil)
}

func (eh *Handler) CreateForGroup(w http.ResponseWriter, r *http.Request) error {
	groupID, err := strconv.Atoi(r.PathValue("groupID"))
	if err != nil || groupID <= 0 {
		return commonerrors.Validation("invalid group ID")
	}

	return eh.create(w, r, &groupID)
}

func (eh *Handler) create(w http.ResponseWriter, r *http.Request, groupID *int) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := eh.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	reqData, err := request.ReadBody[Input](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	if groupID == nil {
		err = eh.eventService.Create(ctx, userID, *reqData)
	} else {
		err = eh.eventService.CreateForGroup(ctx, userID, *groupID, *reqData)
	}
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusCreated, nil)
	return nil
}

func (eh *Handler) ToggleActive(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := eh.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	eventID, err := strconv.Atoi(r.PathValue("eventID"))
	if err != nil {
		return err
	}

	err = eh.eventService.ToggleActive(ctx, userID, eventID)
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusCreated, nil)
	return nil
}

func (eh *Handler) AddTimeBlocks(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := eh.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	eventID, err := strconv.Atoi(r.PathValue("eventID"))
	if err != nil {
		return err
	}

	reqData, err := request.ReadBody[SubmitAvailabilityRequest](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}
	if err = reqData.Validate(); err != nil {
		return err
	}

	err = eh.eventService.AddTimeBlocks(ctx, userID, eventID, *reqData)
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}
