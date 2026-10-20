package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type (
	authRepository interface {
		InsertRefreshToken(
			ctx context.Context,
			ipAddress string,
			deviceInfo string,
			refreshToken string,
			userID int,
		) error
		GetRefreshTokenByTokenHash(ctx context.Context, tokenHash string) (TokenWithUser, error)
		UpdateLastTimeUsedToken(ctx context.Context, refreshToken string) error
		RemoveTokenByToken(ctx context.Context, refreshToken string) error
		RemoveTokensByUserID(ctx context.Context, userID int) error
	}
	userRepository interface {
		CheckDoesExistByEmail(ctx context.Context, email string) (bool, error)
		Insert(ctx context.Context, username, email, password string) error
		GetByEmail(ctx context.Context, email string) (user.Instance, error)
	}
)

type authenticationMiddleware interface {
	GenerateRefreshToken() ([]byte, error)
	HashToken(token []byte) string
	GenerateAccessToken(user user.Instance) (string, error)
	VerifyToken(r *http.Request) (*http.Request, error)
	BlacklistUser(r *http.Request) error
}

type Service struct {
	logger                   commoninterfaces.Logger
	authRepository           authRepository
	userRepository           userRepository
	authenticationMiddleware authenticationMiddleware
}

func NewService(
	authRepository authRepository,
	userRepository userRepository,
	authenticationMiddleware authenticationMiddleware,
	logger commoninterfaces.Logger,
) *Service {
	return &Service{
		logger:                   logger,
		authRepository:           authRepository,
		userRepository:           userRepository,
		authenticationMiddleware: authenticationMiddleware,
	}
}

func (as *Service) Register(ctx context.Context, register Register) error {
	doesUserExist, err := as.userRepository.CheckDoesExistByEmail(ctx, register.Email)
	if err != nil {
		return err
	}
	if doesUserExist {
		return commonerrors.NewAPI(
			http.StatusConflict,
			commonerrors.CategoryValidation,
			"user with this email already exists",
			true,
		)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(register.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	register.Password = string(hashedPassword)

	return as.userRepository.Insert(ctx, register.Username, register.Email, register.Password)
}

func (as *Service) Login(
	ctx context.Context,
	login Login,
	ipAddress string,
	deviceInfo string,
) (string, []byte, error) {
	user, err := as.userRepository.GetByEmail(ctx, login.Email)
	if err != nil {
		return "", nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		return "", nil, commonerrors.Unauthorized("provided invalid password")
	}

	if err = ctx.Err(); err != nil {
		return "", nil, err
	}

	accessToken, err := as.authenticationMiddleware.GenerateAccessToken(user)
	if err != nil {
		return "", nil, err
	}

	refreshToken, err := as.authenticationMiddleware.GenerateRefreshToken()
	if err != nil {
		return "", nil, err
	}

	hashedRefreshToken := as.authenticationMiddleware.HashToken(refreshToken)

	if err = ctx.Err(); err != nil {
		return "", nil, err
	}
	if err = ctx.Err(); err != nil {
		return "", nil, err
	}

	err = as.authRepository.InsertRefreshToken(ctx, ipAddress, deviceInfo, hashedRefreshToken, user.ID)
	if err != nil {
		return "", nil, err
	}

	return accessToken, refreshToken, nil
}

func (as *Service) validateToken(tokenWithUser TokenWithUser) error {
	if tokenWithUser.ID == 0 {
		err := errors.New("token not found")
		as.logger.Info(err.Error(), nil)
		return commonerrors.Unauthorized(err.Error())
	}

	if tokenWithUser.ExpiresAt.Before(time.Now()) {
		err := errors.New("refresh token expired")
		as.logger.Info(err.Error(), nil)
		return commonerrors.Unauthorized(err.Error())
	}
	return nil
}

func (as *Service) RefreshAccessToken(ctx context.Context, refreshToken []byte) (string, error) {
	hashedRefreshToken := as.authenticationMiddleware.HashToken(refreshToken)
	tokenWithUser, err := as.authRepository.GetRefreshTokenByTokenHash(ctx, hashedRefreshToken)
	if err != nil {
		return "", err
	}

	err = as.validateToken(tokenWithUser)
	if err != nil {
		return "", err
	}

	err = as.authRepository.UpdateLastTimeUsedToken(ctx, hashedRefreshToken)
	if err != nil {
		return "", err
	}

	user := user.Instance{
		ID:       tokenWithUser.ID,
		Email:    tokenWithUser.Email,
		Username: tokenWithUser.Username,
	}

	accessToken, err := as.authenticationMiddleware.GenerateAccessToken(user)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (as *Service) Logout(ctx context.Context, refreshToken []byte) error {
	hashedRefreshToken := as.authenticationMiddleware.HashToken(refreshToken)

	err := as.authRepository.RemoveTokenByToken(ctx, hashedRefreshToken)
	if err != nil {
		return err
	}

	return nil
}

func (as *Service) LogoutFromAllDevices(ctx context.Context, userID int) error {
	err := as.authRepository.RemoveTokensByUserID(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}
