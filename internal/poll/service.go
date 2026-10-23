package poll

import (
	"context"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

type pollRepository interface {
	Insert(ctx context.Context, userID, eventID int) error
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
	return nil
}
