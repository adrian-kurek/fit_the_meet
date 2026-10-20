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
	RefreshAccessToken(ctx context.Context, refreshToken []byte) (string, error)
	Logout(ctx context.Context, refreshToken []byte) error
	LogoutFromAllDevices(ctx context.Context, userID int) error
}

type Handler struct {
	logger                   commoninterfaces.Logger
	authService              authService
	authenticationMiddleware authenticationMiddleware
}

func NewHandler(
	authService authService,
	authenticationMiddleware authenticationMiddleware,
	logger commoninterfaces.Logger,
) *Handler {
	return &Handler{
		logger:                   logger,
		authService:              authService,
		authenticationMiddleware: authenticationMiddleware,
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

func (ah *Handler) Verify(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	r = r.WithContext(ctx)

	r, err := ah.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}

func (ah *Handler) readRefreshToken(r *http.Request) ([]byte, error) {
	refreshToken, err := r.Cookie("refreshToken")
	if err != nil {
		ah.logger.Error("failed to read cookie from request", "err", err.Error())
		return nil, err
	}

	tokenBytes, err := hex.DecodeString(refreshToken.Value)
	if err != nil {
		ah.logger.Error("failed to decode string into bytes", "err", err.Error())
		return nil, err
	}
	return tokenBytes, nil
}

func (ah *Handler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	refreshToken, err := ah.readRefreshToken(r)
	if err != nil {
		return err
	}

	newAccessToken, err := ah.authService.RefreshAccessToken(ctx, refreshToken)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]string{"token": newAccessToken})

	return nil
}

func (ah *Handler) Logout(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	err := ah.authenticationMiddleware.BlacklistUser(r)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	refreshToken, err := ah.readRefreshToken(r)
	if err != nil {
		return err
	}

	err = ah.authService.Logout(ctx, refreshToken)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)
	return nil
}

func (ah *Handler) LogoutFromAllDevices(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	r, err := ah.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return err
	}

	err = ah.authService.LogoutFromAllDevices(ctx, userID)
	if err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}
