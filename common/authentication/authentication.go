package middleware

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
	"github.com/adrian-kurek/fit_the_meet/internal/user"
	"github.com/golang-jwt/jwt/v4"
)

const accessTokenExpiration = 5 * time.Minute

type Middleware struct {
	accessTokenSecret  string
	refreshTokenSecret string
	logger             commoninterfaces.Logger
	cache              commoninterfaces.CacheService
}

func NewMiddleware(
	accessTokenSecret string,
	refreshTokenSecret string,
	loggerService commoninterfaces.Logger,
	cacheService commoninterfaces.CacheService,
) *Middleware {
	return &Middleware{
		accessTokenSecret:  accessTokenSecret,
		refreshTokenSecret: refreshTokenSecret,
		logger:             loggerService,
		cache:              cacheService,
	}
}

func (am *Middleware) GenerateRefreshToken() ([]byte, error) {
	lengthOfRefreshToken := 64
	bytes := make([]byte, lengthOfRefreshToken)

	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (am *Middleware) HashToken(token []byte) string {
	hasher := sha256.New()
	hasher.Write(token)
	return hex.EncodeToString(hasher.Sum(nil))
}

func (am *Middleware) GenerateAccessToken(user user.Instance) (string, error) {
	am.logger.Info("started signing a new access token", nil)
	tokenWithData := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       user.ID,
		"email":    user.Email,
		"username": user.Username,
		"exp":      time.Now().Add(accessTokenExpiration).Unix(),
	})

	tokenString, err := tokenWithData.SignedString([]byte(am.accessTokenSecret))
	if err != nil {
		errMsg := errors.New("failed to sign access token properly")
		am.logger.Error(errMsg.Error(), err.Error())
		return "", commonerrors.Unauthorized(errMsg.Error())
	}

	am.logger.Info("successfully signed a new access token", nil)
	return tokenString, nil
}

func (am *Middleware) ParseClaimsFromToken(tokenString string) (*jwt.Token, user.Claims, error) {
	var claims user.Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.accessTokenSecret), nil
	})
	if err != nil {
		return nil, user.Claims{}, err
	}
	return token, claims, nil
}

func (am *Middleware) readTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		am.logger.Info("token is missing", nil)
		return "", commonerrors.Unauthorized("failed to authorize a user")
	}

	return strings.Split(authHeader, " ")[1], nil
}

func (am *Middleware) isTokenBlackListed(ctx context.Context, token string) error {
	cacheKey := "tokenBlackList-" + token
	result, err := am.cache.Exists(ctx, cacheKey)
	if err != nil {
		am.logger.Info("failed to check blacklist", err)
		return err
	}

	if result > 0 {
		err = errors.New("token blacklisted")
		am.logger.Info(err.Error(), nil)
		return commonerrors.Unauthorized(err.Error())
	}
	return nil
}

func (am *Middleware) checkIsTokenValid(token *jwt.Token) error {
	if !token.Valid {
		err := errors.New("provided token is invalid")
		am.logger.Info(err.Error(), nil)
		return commonerrors.Unauthorized(err.Error())
	}
	return nil
}

func (am *Middleware) VerifyToken(r *http.Request) (*http.Request, error) {
	ctx := r.Context()
	if err := ctx.Err(); err != nil {
		return r, err
	}

	token, err := am.readTokenFromRequest(r)
	if err != nil {
		return r, err
	}

	err = am.isTokenBlackListed(ctx, token)
	if err != nil {
		return r, err
	}

	if err = ctx.Err(); err != nil {
		return r, err
	}

	tokenWithData, user, err := am.ParseClaimsFromToken(token)
	if err != nil {
		am.logger.Info("failed to read data properly", err.Error())
		return r, commonerrors.Unauthorized("provided token is invalid")
	}

	if err = ctx.Err(); err != nil {
		return r, err
	}

	err = am.checkIsTokenValid(tokenWithData)
	if err != nil {
		return r, err
	}

	r = request.SetContext(r, "id", user.ID)
	r = request.SetContext(r, "email", user.Email)
	r = request.SetContext(r, "username", user.Username)
	return r, nil
}

func (am *Middleware) BlacklistUser(r *http.Request) error {
	ctx := r.Context()

	token, err := am.readTokenFromRequest(r)
	if err != nil {
		return err
	}

	tokenWithData, user, err := am.ParseClaimsFromToken(token)
	if err != nil {
		am.logger.Info("failed to read data properly", nil)
		return commonerrors.Unauthorized("failed to read token")
	}

	err = am.checkIsTokenValid(tokenWithData)
	if err != nil {
		return err
	}
	err = am.isTokenBlackListed(ctx, token)
	if err != nil {
		return err
	}

	expirationTime := time.Until(user.ExpiresAt.Time)

	cacheKey := "tokenBlackList-" + token
	err = am.cache.Set(ctx, cacheKey, "true", expirationTime)
	if err != nil {
		am.logger.Info("failed to set data in cache", err)
		return err
	}

	return nil
}
