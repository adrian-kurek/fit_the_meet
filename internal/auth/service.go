package auth

import (
	"context"
	"net/http"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"golang.org/x/crypto/bcrypt"
)

type (
	authRepository any
	userRepository interface {
		CheckDoesExistByEmail(ctx context.Context, email string) (bool, error)
		Insert(ctx context.Context, register Register) error
	}
)

type Service struct {
	logger         commoninterfaces.Logger
	authRepository authRepository
	userRepository userRepository
}

func NewService(authRepository authRepository, userRepository userRepository, logger commoninterfaces.Logger) *Service {
	return &Service{
		logger:         logger,
		authRepository: authRepository,
		userRepository: userRepository,
	}
}

func (as *Service) Register(ctx context.Context, registerData Register) error {
	doesUserExist, err := as.userRepository.CheckDoesExistByEmail(ctx, registerData.Email)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(registerData.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	registerData.Password = string(hashedPassword)

	return as.userRepository.Insert(ctx, registerData)
}
