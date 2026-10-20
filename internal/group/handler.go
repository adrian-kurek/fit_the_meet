package group

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

const groupTimeout = 5 * time.Second

type (
	groupService interface {
		Create(ctx context.Context, name string, userID int) error
		UpdateName(ctx context.Context, name string, groupID, userID int) error
	}
	authenticationMiddleware interface {
		VerifyToken(r *http.Request) (*http.Request, error)
	}
)

type Handler struct {
	logger                   commoninterfaces.Logger
	groupService             groupService
	authenticationMiddleware authenticationMiddleware
}

func NewHandler(
	groupService groupService,
	authenticationMiddleware authenticationMiddleware,
	logger commoninterfaces.Logger,
) *Handler {
	return &Handler{
		groupService:             groupService,
		authenticationMiddleware: authenticationMiddleware,
		logger:                   logger,
	}
}

func (gh *Handler) handleTimeout(err error, path string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		gh.logger.Info("request timed out", "path", path)
		return commonerrors.RequestTimeout()
	}
	return err
}

func (gh *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	r, err := gh.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	reqData, err := request.ReadBody[Upsert](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	err = gh.groupService.Create(ctx, reqData.Name, userID)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusCreated, nil)

	return nil
}

func (gh *Handler) UpdateName(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	r, err := gh.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	groupID, err := strconv.Atoi(r.PathValue("groupID"))
	if err != nil {
		return err
	}

	reqData, err := request.ReadBody[Upsert](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	err = gh.groupService.UpdateName(ctx, reqData.Name, groupID, userID)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}
