package auth

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
	"github.com/adrian-kurek/fit_the_meet/common/response"
	"github.com/adrian-kurek/fit_the_meet/common/validation"
)

const authTimeout = 2 * time.Second

type authService interface {
	Register(ctx context.Context, register Register) error
	Login(ctx context.Context,
		login Login,
		ipAddress string,
		deviceInfo string,
	) (string, []byte, error)
}

type Handler struct {
	logger      commoninterfaces.Logger
	authService authService
}

func NewHandler(authService authService, logger commoninterfaces.Logger) *Handler {
	return &Handler{
		logger:      logger,
		authService: authService,
	}
}

func (ah *Handler) handleTimeout(err error, path string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		ah.logger.Info("request timed out", "path", path)
		return commonerrors.RequestTimeout()
	}
	return err
}

func (ah *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	reqData, err := request.ReadBody[Register](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	err = ah.authService.Register(ctx, *reqData)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusCreated, nil)
	return nil
}

func (ah *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	reqData, err := request.ReadBody[Login](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	ipAddress := r.RemoteAddr
	deviceInfo := r.UserAgent()

	accessToken, refreshToken, err := ah.authService.Login(ctx, *reqData, ipAddress, deviceInfo)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	expiration := time.Now().Add(7 * 24 * time.Hour)

	cookie := http.Cookie{
		Name:     "refreshToken",
		Value:    hex.EncodeToString(refreshToken),
		Expires:  expiration,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		HttpOnly: true,
	}

	http.SetCookie(w, &cookie)
	response.Send(w, http.StatusOK, map[string]string{"token": accessToken})

	return nil
}
