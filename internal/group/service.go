package group

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

type groupRepository interface {
	Insert(ctx context.Context, name string, userID int) error
	UpdateName(ctx context.Context, name string, groupID, userID int) error
	Delete(ctx context.Context, groupID, userID int) error
	RemoveParticipants(ctx context.Context, groupID, userID int, participants []int) error
	IsUserOwnerOfTheGroup(ctx context.Context, groupID, userID int) error
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

func (gs *Service) Remove(ctx context.Context, groupID, userID int) error {
	return gs.groupRepository.Delete(ctx, groupID, userID)
}

func (gs *Service) RemoveParticipants(ctx context.Context, groupID, userID int, participants []int) error {
	return gs.groupRepository.RemoveParticipants(ctx, groupID, userID, participants)
}

func (gs *Service) GenerateInviteLink(ctx context.Context, groupID, userID int) (string, error) {
	err := gs.groupRepository.IsUserOwnerOfTheGroup(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", commonerrors.NewAPI(
				http.StatusForbidden,
				commonerrors.CategoryForbidden,
				"user is not allowed to perform this action",
				true,
			)
		}
		return "", err
	}

	// code := uuid.New().String()

	return "", nil
}
