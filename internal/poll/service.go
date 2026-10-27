package poll

import (
	"context"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/google/uuid"
)

type pollRepository interface {
	Create(ctx context.Context, userID int, code string, poll Input) error
	CreateForGroup(ctx context.Context, userID, groupID int, code string, poll Input) error
	InsertAnswers(ctx context.Context, pollID, userID int, answers []Answer) error
	ToggleActive(ctx context.Context, pollID, userID int) error
	Get(ctx context.Context, userID int) ([]Info, error)
	GetByID(ctx context.Context, userID, pollID int) (Detail, error)
	Delete(ctx context.Context, userID, pollID int) error
	DeleteAnswers(ctx context.Context, userID, pollID int) error
	GetResults(ctx context.Context, userID, pollID int) (Result, error)
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

func (ps *Service) CreateForGroup(ctx context.Context, userID, groupID int, poll Input) error {
	code := uuid.New().String()
	return ps.pollRepository.CreateForGroup(ctx, userID, groupID, code, poll)
}

func (ps *Service) SubmitAnswers(ctx context.Context, userID, pollID int, answers SubmitAnswers) error {
	return ps.pollRepository.InsertAnswers(ctx, pollID, userID, answers.Answers)
}

func (ps *Service) ToggleActive(ctx context.Context, userID, pollID int) error {
	return ps.pollRepository.ToggleActive(ctx, userID, pollID)
}

func (ps *Service) Get(ctx context.Context, userID int) ([]Info, error) {
	return ps.pollRepository.Get(ctx, userID)
}

func (ps *Service) GetByID(ctx context.Context, userID, pollID int) (Detail, error) {
	return ps.pollRepository.GetByID(ctx, userID, pollID)
}

func (ps *Service) Delete(ctx context.Context, userID, pollID int) error {
	return ps.pollRepository.Delete(ctx, userID, pollID)
}

func (ps *Service) DeleteAnswers(ctx context.Context, userID, pollID int) error {
	return ps.pollRepository.DeleteAnswers(ctx, userID, pollID)
}

func (ps *Service) GetResults(ctx context.Context, userID, pollID int) (Result, error) {
	return ps.pollRepository.GetResults(ctx, userID, pollID)
}
