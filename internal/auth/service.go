package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type (
	authRepository interface {
		SaveRefreshToken(
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
		SavePasswordResetToken(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
		ResetPassword(ctx context.Context, tokenHash, password string) error
	}
	userRepository interface {
		ExistsByEmail(ctx context.Context, email string) (bool, error)
		Insert(ctx context.Context, username, email, password string) error
		GetByEmail(ctx context.Context, email string) (user.Account, error)
	}
	passwordResetMailer interface {
		SendPasswordReset(ctx context.Context, email, resetURL string) error
	}
)

type authenticationMiddleware interface {
	GenerateRefreshToken() ([]byte, error)
	HashToken(token []byte) string
	GenerateAccessToken(user user.Account) (string, error)
	VerifyToken(r *http.Request) (*http.Request, error)
	RevokeAccessToken(r *http.Request) error
}

type Service struct {
	logger                   commoninterfaces.Logger
	authRepository           authRepository
	userRepository           userRepository
	authenticationMiddleware authenticationMiddleware
	passwordResetMailer      passwordResetMailer
	frontendURL              string
}

func NewService(
	authRepository authRepository,
	userRepository userRepository,
	authenticationMiddleware authenticationMiddleware,
	passwordResetMailer passwordResetMailer,
	frontendURL string,
	logger commoninterfaces.Logger,
) *Service {
	if strings.TrimSpace(frontendURL) == "" {
		frontendURL = "http://localhost:5173"
	}

	return &Service{
		logger:                   logger,
		authRepository:           authRepository,
		userRepository:           userRepository,
		authenticationMiddleware: authenticationMiddleware,
		passwordResetMailer:      passwordResetMailer,
		frontendURL:              strings.TrimRight(strings.TrimSpace(frontendURL), "/"),
	}
}

const passwordResetExpiration = 30 * time.Minute

func (as *Service) Register(ctx context.Context, register RegisterRequest) error {
	doesUserExist, err := as.userRepository.ExistsByEmail(ctx, register.Email)
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
	login LoginRequest,
	ipAddress string,
	deviceInfo string,
) (accessToken string, refreshToken []byte, id int, username string, err error) {
	user, err := as.userRepository.GetByEmail(ctx, login.Email)
	if err != nil {
		return "", nil, 0, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		return "", nil, 0, "", commonerrors.Unauthorized("provided invalid password")
	}

	if err = ctx.Err(); err != nil {
		return "", nil, 0, "", err
	}

	accessToken, err = as.authenticationMiddleware.GenerateAccessToken(user)
	if err != nil {
		return "", nil, 0, "", err
	}

	refreshToken, err = as.authenticationMiddleware.GenerateRefreshToken()
	if err != nil {
		return "", nil, 0, "", err
	}

	hashedRefreshToken := as.authenticationMiddleware.HashToken(refreshToken)

	if err = ctx.Err(); err != nil {
		return "", nil, 0, "", err
	}
	if err = ctx.Err(); err != nil {
		return "", nil, 0, "", err
	}

	err = as.authRepository.SaveRefreshToken(ctx, ipAddress, deviceInfo, hashedRefreshToken, user.ID)
	if err != nil {
		return "", nil, 0, "", err
	}

	return accessToken, refreshToken, user.ID, user.Username, nil
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

	user := user.Account{
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

func (as *Service) RequestPasswordReset(ctx context.Context, email string) error {
	account, err := as.userRepository.GetByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		apiErr := &commonerrors.API{}
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}

	token := make([]byte, 32)
	if _, err = rand.Read(token); err != nil {
		return err
	}
	tokenValue := base64.RawURLEncoding.EncodeToString(token)
	tokenHash := hashPasswordResetToken(tokenValue)
	if err = as.authRepository.SavePasswordResetToken(
		ctx,
		account.ID,
		tokenHash,
		time.Now().Add(passwordResetExpiration),
	); err != nil {
		return err
	}

	resetURL, err := url.Parse(as.frontendURL)
	if err != nil {
		return err
	}
	query := resetURL.Query()
	query.Set("reset_token", tokenValue)
	resetURL.RawQuery = query.Encode()

	return as.passwordResetMailer.SendPasswordReset(ctx, account.Email, resetURL.String())
}

func (as *Service) ResetPassword(ctx context.Context, reset ResetPasswordRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reset.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return as.authRepository.ResetPassword(
		ctx,
		hashPasswordResetToken(strings.TrimSpace(reset.Token)),
		string(hashedPassword),
	)
}

func hashPasswordResetToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
