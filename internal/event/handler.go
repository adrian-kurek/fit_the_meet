package event

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
	"github.com/adrian-kurek/fit_the_meet/common/response"
	"github.com/adrian-kurek/fit_the_meet/common/validation"
)

type eventService interface {
	Create(ctx context.Context, userID int, event Input) error
	CreateForGroup(ctx context.Context, userID, groupID int, event Input) error
	ToggleActive(ctx context.Context, userID int, eventID int) error
	AddTimeBlocks(ctx context.Context, userID int, eventID int, timeBlocks SubmitAvailabilityRequest) error
	GetMany(ctx context.Context, userID int, options ListOptions) ([]Info, error)
}

type authenticationMiddleware interface {
	VerifyToken(r *http.Request) (*http.Request, error)
}

const (
	timeout              = 5 * time.Second
	defaultEventPageSize = 20
	maxEventPageSize     = 100
)

type Handler struct {
	eventService             eventService
	authenticationMiddleware authenticationMiddleware
	logger                   commoninterfaces.Logger
}

func NewHandler(
	eventService eventService,
	authenticationMiddleware authenticationMiddleware,
	logger commoninterfaces.Logger,
) *Handler {
	return &Handler{
		eventService:             eventService,
		authenticationMiddleware: authenticationMiddleware,
		logger:                   logger,
	}
}

func (eh *Handler) handleTimeout(err error, path string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		eh.logger.Info("request timed out", "path", path)
		return commonerrors.RequestTimeout()
	}
	return err
}

func (eh *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	return eh.create(w, r, nil)
}

func (eh *Handler) authenticateAndGetUserID(r *http.Request) (int, error) {
	r, err := eh.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return 0, eh.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (eh *Handler) CreateForGroup(w http.ResponseWriter, r *http.Request) error {
	groupID, err := strconv.Atoi(r.PathValue("groupID"))
	if err != nil || groupID <= 0 {
		return commonerrors.Validation("invalid group ID")
	}

	return eh.create(w, r, &groupID)
}

func (eh *Handler) create(w http.ResponseWriter, r *http.Request, groupID *int) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	userID, err := eh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	reqData, err := request.ReadBody[Input](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}

	if groupID == nil {
		err = eh.eventService.Create(ctx, userID, *reqData)
	} else {
		err = eh.eventService.CreateForGroup(ctx, userID, *groupID, *reqData)
	}
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusCreated, nil)
	return nil
}

func (eh *Handler) ToggleActive(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	userID, err := eh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	eventID, err := strconv.Atoi(r.PathValue("eventID"))
	if err != nil {
		return err
	}

	err = eh.eventService.ToggleActive(ctx, userID, eventID)
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusCreated, nil)
	return nil
}

func (eh *Handler) AddTimeBlocks(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	userID, err := eh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	eventID, err := strconv.Atoi(r.PathValue("eventID"))
	if err != nil {
		return err
	}

	reqData, err := request.ReadBody[SubmitAvailabilityRequest](r)
	if err != nil {
		return commonerrors.Validation("provided invalid json format")
	}

	err = validation.ValidateRequestData(reqData)
	if err != nil {
		return err
	}
	if err = reqData.Validate(); err != nil {
		return err
	}

	err = eh.eventService.AddTimeBlocks(ctx, userID, eventID, *reqData)
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}

func (eh *Handler) GetMany(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	userID, err := eh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	page, pageSize, options, err := readListOptions(r)
	if err != nil {
		return err
	}

	events, err := eh.eventService.GetMany(ctx, userID, options)
	if err != nil {
		return eh.handleTimeout(err, r.URL.Path)
	}
	if events == nil {
		events = make([]Info, 0)
	}

	hasMore := len(events) > pageSize
	if hasMore {
		events = events[:pageSize]
	}

	response.Send(w, http.StatusOK, map[string]any{
		"events": events,
		"pagination": map[string]any{
			"page":      page,
			"page_size": pageSize,
			"has_more":  hasMore,
		},
	})

	return nil
}

func readListOptions(r *http.Request) (int, int, ListOptions, error) {
	page, err := readPositiveQueryInt(r, "page", 1)
	if err != nil {
		return 0, 0, ListOptions{}, err
	}

	pageSize, err := readPositiveQueryInt(r, "page_size", defaultEventPageSize)
	if err != nil {
		return 0, 0, ListOptions{}, err
	}
	if pageSize > maxEventPageSize {
		return 0, 0, ListOptions{}, commonerrors.Validation("page_size must not exceed 100")
	}

	// Keep the offset calculation bounded on all supported architectures.
	if page > 1_000_000 {
		return 0, 0, ListOptions{}, commonerrors.Validation("page must not exceed 1000000")
	}

	options := ListOptions{
		Limit:  pageSize + 1,
		Offset: (page - 1) * pageSize,
	}

	if rawIsActive := r.URL.Query().Get("is_active"); rawIsActive != "" {
		isActive, parseErr := strconv.ParseBool(rawIsActive)
		if parseErr != nil {
			return 0, 0, ListOptions{}, commonerrors.Validation("is_active must be true or false")
		}
		options.IsActive = &isActive
	}

	if rawGroupID := r.URL.Query().Get("group_id"); rawGroupID != "" {
		groupID, parseErr := strconv.Atoi(rawGroupID)
		if parseErr != nil || groupID <= 0 {
			return 0, 0, ListOptions{}, commonerrors.Validation("group_id must be a positive integer")
		}
		options.GroupID = &groupID
	}

	return page, pageSize, options, nil
}

func readPositiveQueryInt(r *http.Request, name string, fallback int) (int, error) {
	rawValue := r.URL.Query().Get(name)
	if rawValue == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(rawValue)
	if err != nil || value <= 0 {
		return 0, commonerrors.Validation(name + " must be a positive integer")
	}

	return value, nil
}
