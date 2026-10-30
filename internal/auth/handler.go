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
	Register(ctx context.Context, register RegisterRequest) error
	Login(ctx context.Context,
		login LoginRequest,
		ipAddress string,
		deviceInfo string,
	) (string, []byte, int, string, error)
	RefreshAccessToken(ctx context.Context, refreshToken []byte) (string, error)
	Logout(ctx context.Context, refreshToken []byte) error
	LogoutFromAllDevices(ctx context.Context, userID int) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, reset ResetPasswordRequest) error
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

func (ah *Handler) authenticateAndGetUserID(r *http.Request) (int, error) {
	r, err := ah.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return 0, ah.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (ah *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	reqData, err := request.ReadBody[RegisterRequest](r)
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

	reqData, err := request.ReadBody[LoginRequest](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	ipAddress := r.RemoteAddr
	deviceInfo := r.UserAgent()

	accessToken, refreshToken, userID, username, err := ah.authService.Login(ctx, *reqData, ipAddress, deviceInfo)
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
	response.Send(w, http.StatusOK, map[string]any{
		"token": accessToken,
		"user": map[string]any{
			"id":       userID,
			"email":    reqData.Email,
			"username": username,
		},
	})

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

	err := ah.authenticationMiddleware.RevokeAccessToken(r)
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

	userID, err := ah.authenticateAndGetUserID(r)
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

func (ah *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	reqData, err := request.ReadAndValidateBody[RequestPasswordResetRequest](r)
	if err != nil {
		return err
	}

	if err = ah.authService.RequestPasswordReset(ctx, reqData.Email); err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusAccepted, nil)
	return nil
}

func (ah *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), authTimeout)
	defer cancel()

	reqData, err := request.ReadAndValidateBody[ResetPasswordRequest](r)
	if err != nil {
		return err
	}

	if err = ah.authService.ResetPassword(ctx, *reqData); err != nil {
		return ah.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)
	return nil
}
