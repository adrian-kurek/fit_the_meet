package group

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
)

const lifespanOfActicationCode = 15 * time.Minute

type groupRepository interface {
	Insert(ctx context.Context, name string, userID int) error
	UpdateName(ctx context.Context, name string, groupID, userID int) error
	Delete(ctx context.Context, groupID, userID int) error
	RemoveParticipants(ctx context.Context, groupID, userID int, participants []int) error
	IsUserOwnerOfTheGroup(ctx context.Context, groupID, userID int) error
	InsertActivationCode(ctx context.Context, groupID int, code string, expiresAt time.Time) error
	InsertParticipantToGroup(ctx context.Context, groupID, userID int) error
	CheckActivationCode(ctx context.Context, code string) (int, error)
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

func (gs *Service) GenerateInviteCode(ctx context.Context, groupID, userID int) (string, error) {
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

	code := uuid.New().String()

	expiresAt := time.Now().Add(lifespanOfActicationCode)

	err = gs.groupRepository.InsertActivationCode(ctx, groupID, code, expiresAt)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (gs *Service) Join(ctx context.Context, userID int, code string) error {
	groupID, err := gs.groupRepository.CheckActivationCode(ctx, code)
	if err != nil {
		return err
	}
	return gs.groupRepository.InsertParticipantToGroup(ctx, groupID, userID)
}
