package event

import (
	"context"

	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/google/uuid"
)

type eventRepository interface {
	Insert(ctx context.Context, userID int, code string, event Upsert) error
	ChangeStatus(ctx context.Context, eventID, userID int) error
	InsertTimeBlocks(ctx context.Context, userID, eventID int, timeBlocks UpsertTimeBlocks) error
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

func (es *Service) Create(ctx context.Context, userID int, event Upsert) error {
	code := uuid.New().String()

	return es.eventRepository.Insert(ctx, userID, code, event)
}

func (es *Service) ChangeStatus(ctx context.Context, userID int, eventID int) error {
	return es.eventRepository.ChangeStatus(ctx, eventID, userID)
}

func (es *Service) AddTimeBlocks(ctx context.Context, userID, eventID int, timeBlocks UpsertTimeBlocks) error {
	return es.eventRepository.InsertTimeBlocks(ctx, userID, eventID, timeBlocks)
}
