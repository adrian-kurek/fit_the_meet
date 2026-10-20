package group

import (
	"context"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

type groupRepository interface {
	Insert(ctx context.Context, name string, userID int) error
	UpdateName(ctx context.Context, name string, groupID, userID int) error
}

type Service struct {
	logger          commoninterfaces.Logger
	groupRepository groupRepository
}

func NewService(groupRepository groupRepository, logger commoninterfaces.Logger) *Service {
	return &Service{
		groupRepository: groupRepository,
		logger:          logger,
	}
}

func (gs *Service) Create(ctx context.Context, name string, userID int) error {
	return gs.groupRepository.Insert(ctx, name, userID)
}

func (gs *Service) UpdateName(ctx context.Context, name string, groupID, userID int) error {
	return gs.groupRepository.UpdateName(ctx, name, groupID, userID)
}
