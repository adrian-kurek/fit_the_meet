package event

import (
	"context"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/google/uuid"
)

type eventRepository interface {
	Insert(ctx context.Context, userID int, code string, event Input) (int, error)
	InsertForGroup(ctx context.Context, userID, groupID int, code string, event Input) error
	ToggleActive(ctx context.Context, eventID, userID int) error
	InsertTimeBlocks(ctx context.Context, userID, eventID int, timeBlocks SubmitAvailabilityRequest) error
}

type Service struct {
	eventRepository eventRepository
	logger          commoninterfaces.Logger
}

func NewService(eventRepository eventRepository, logger commoninterfaces.Logger) *Service {
	return &Service{
		eventRepository: eventRepository,
		logger:          logger,
	}
}

func (es *Service) Create(ctx context.Context, userID int, event Input) error {
	code := uuid.New().String()

	_, err := es.eventRepository.Insert(ctx, userID, code, event)
	return err
}

func (es *Service) CreateForGroup(ctx context.Context, userID, groupID int, event Input) error {
	code := uuid.New().String()

	return es.eventRepository.InsertForGroup(ctx, userID, groupID, code, event)
}

func (es *Service) ToggleActive(ctx context.Context, userID int, eventID int) error {
	return es.eventRepository.ToggleActive(ctx, eventID, userID)
}

func (es *Service) AddTimeBlocks(ctx context.Context, userID, eventID int, timeBlocks SubmitAvailabilityRequest) error {
	return es.eventRepository.InsertTimeBlocks(ctx, userID, eventID, timeBlocks)
}
