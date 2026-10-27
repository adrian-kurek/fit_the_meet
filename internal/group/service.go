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
	Insert(ctx context.Context, name string, userID int) (int, error)
	UpdateName(ctx context.Context, name string, groupID, userID int) error
	DeleteOwnedGroup(ctx context.Context, groupID, userID int) error
	RemoveParticipants(ctx context.Context, groupID, userID int, participants []int) error
	VerifyOwnership(ctx context.Context, groupID, userID int) error
	SaveInviteCode(ctx context.Context, groupID int, code string, expiresAt time.Time) error
	AddMember(ctx context.Context, groupID, userID int) error
	FindGroupIDByInviteCode(ctx context.Context, code string) (int, error)
	GetByID(ctx context.Context, userID, groupID int) (Info, error)
	GetMany(ctx context.Context, userID int) ([]InfoWithoutParticipants, error)
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

func (gs *Service) Create(ctx context.Context, name string, userID int) (int, error) {
	return gs.groupRepository.Insert(ctx, name, userID)
}

func (gs *Service) UpdateName(ctx context.Context, name string, groupID, userID int) error {
	return gs.groupRepository.UpdateName(ctx, name, groupID, userID)
}

func (gs *Service) Remove(ctx context.Context, groupID, userID int) error {
	return gs.groupRepository.DeleteOwnedGroup(ctx, groupID, userID)
}

func (gs *Service) RemoveParticipants(ctx context.Context, groupID, userID int, participants []int) error {
	return gs.groupRepository.RemoveParticipants(ctx, groupID, userID, participants)
}

func (gs *Service) GenerateInviteCode(ctx context.Context, groupID, userID int) (string, error) {
	err := gs.groupRepository.VerifyOwnership(ctx, groupID, userID)
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

	err = gs.groupRepository.SaveInviteCode(ctx, groupID, code, expiresAt)
	if err != nil {
		return "", err
	}

	return code, nil
}

func (gs *Service) Join(ctx context.Context, userID int, code string) (Joined, error) {
	groupID, err := gs.groupRepository.FindGroupIDByInviteCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Joined{}, commonerrors.NotFound("activation code not found or expired")
		}
		return Joined{}, err
	}

	if err = gs.groupRepository.AddMember(ctx, groupID, userID); err != nil {
		return Joined{}, err
	}

	group, err := gs.groupRepository.GetByID(ctx, userID, groupID)
	if err != nil {
		return Joined{}, err
	}

	return Joined{
		ID:           group.ID,
		Name:         group.Name,
		Role:         group.Role,
		MembersCount: group.MembersCount,
	}, nil
}

func (gs *Service) GetByID(ctx context.Context, userID int, groupID int) (Info, error) {
	return gs.groupRepository.GetByID(ctx, userID, groupID)
}

func (gs *Service) GetMany(ctx context.Context, userID int) ([]InfoWithoutParticipants, error) {
	return gs.groupRepository.GetMany(ctx, userID)
}
