package poll

import (
	"context"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/google/uuid"
)

type pollRepository interface {
	Create(ctx context.Context, userID int, code string, poll Input) error
	InsertAnswers(ctx context.Context, pollID, userID int, answers []Answer) error
}

type Service struct {
	pollRepository pollRepository
	logger         commoninterfaces.Logger
}

func NewService(pollRepository pollRepository, logger commoninterfaces.Logger) *Service {
	return &Service{
		pollRepository: pollRepository,
		logger:         logger,
	}
}

func (ps *Service) Create(ctx context.Context, userID int, poll Input) error {
	code := uuid.New().String()
	return ps.pollRepository.Create(ctx, userID, code, poll)
}

func (ps *Service) SubmitAnswers(ctx context.Context, userID, pollID int, answers SubmitAnswers) error {
	return ps.pollRepository.InsertAnswers(ctx, pollID, userID, answers.Answers)
}
