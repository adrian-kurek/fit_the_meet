package poll

import (
	"context"
	"errors"
	"net/http"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
	"github.com/adrian-kurek/fit_the_meet/common/validation"
)

type pollService interface {
	Create(ctx context.Context, userID int, poll Input) error
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

func NewHandler(pollService pollService, authenticationMiddleware authenticationMiddleware, logger commoninterfaces.Logger) *Handler {
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

	return nil
}
