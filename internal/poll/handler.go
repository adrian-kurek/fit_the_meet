package poll

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

type pollService interface {
	Create(ctx context.Context, userID int, poll Input) error
	SubmitAnswers(ctx context.Context, userID, pollID int, answers SubmitAnswers) error
	ToggleActive(ctx context.Context, userID, pollID int) error
	Get(ctx context.Context, userID int) ([]Info, error)
	GetByID(ctx context.Context, userID, pollID int) ([]Info, error)
	DeleteAnswers(ctx context.Context, userID, pollID int) error
	Delete(ctx context.Context, userID, pollID int) error
	GetResults(ctx context.Context, userID, pollID int) (Result, error)
}

const timeout = 5 * time.Second

type authenticationMiddleware interface {
	VerifyToken(r *http.Request) (*http.Request, error)
}
type Handler struct {
	pollService              pollService
	authenticationMiddleware authenticationMiddleware
	logger                   commoninterfaces.Logger
}

func NewHandler(
	pollService pollService,
	authenticationMiddleware authenticationMiddleware,
	logger commoninterfaces.Logger,
) *Handler {
	return &Handler{
		pollService:              pollService,
		authenticationMiddleware: authenticationMiddleware,
		logger:                   logger,
	}
}

func (ph *Handler) handleTimeout(err error, path string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		ph.logger.Info("request timed out", "path", path)
		return commonerrors.RequestTimeout()
	}
	return err
}

func (ph *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := ph.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
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

	err = ph.pollService.Create(ctx, userID, *reqData)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusCreated, nil)
	return nil
}

func (ph *Handler) SubmitAnswers(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := ph.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	reqData, err := request.ReadBody[SubmitAnswers](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	pollID, err := strconv.Atoi(r.PathValue("pollID"))
	if err != nil || pollID <= 0 {
		return commonerrors.Validation("invalid poll ID")
	}

	err = ph.pollService.SubmitAnswers(ctx, userID, pollID, *reqData)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)
	return nil
}

func (ph *Handler) ToggleActive(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := ph.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	pollID, err := strconv.Atoi(r.PathValue("pollID"))
	if err != nil || pollID <= 0 {
		return commonerrors.Validation("invalid poll ID")
	}

	err = ph.pollService.ToggleActive(ctx, userID, pollID)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}

func (ph *Handler) Get(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := ph.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	polls, err := ph.pollService.Get(ctx, userID)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]any{
		"polls": polls,
	})

	return nil
}

func (ph *Handler) GetByID(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := ph.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	pollID, err := strconv.Atoi(r.PathValue("pollID"))
	if err != nil || pollID <= 0 {
		return commonerrors.Validation("invalid poll ID")
	}

	poll, err := ph.pollService.GetByID(ctx, userID, pollID)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]any{
		"poll": poll,
	})

	return nil
}

func (ph *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := ph.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	pollID, err := strconv.Atoi(r.PathValue("pollID"))
	if err != nil || pollID <= 0 {
		return commonerrors.Validation("invalid poll ID")
	}

	err = ph.pollService.Delete(ctx, userID, pollID)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}

func (ph *Handler) DeleteAnswers(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := ph.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	pollID, err := strconv.Atoi(r.PathValue("pollID"))
	if err != nil || pollID <= 0 {
		return commonerrors.Validation("invalid poll ID")
	}

	err = ph.pollService.DeleteAnswers(ctx, userID, pollID)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}

func (ph *Handler) GetResults(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	r, err := ph.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	pollID, err := strconv.Atoi(r.PathValue("pollID"))
	if err != nil || pollID <= 0 {
		return commonerrors.Validation("invalid poll ID")
	}

	res, err := ph.pollService.GetResults(ctx, userID, pollID)
	if err != nil {
		return ph.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]any{
		"result": res,
	})

	return nil
}
