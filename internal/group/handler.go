package group

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	commoninterfaces "github.com/adrian-kurek/fit_the_meet/common/interfaces"
	"github.com/adrian-kurek/fit_the_meet/common/request"
	"github.com/adrian-kurek/fit_the_meet/common/response"
)

const groupTimeout = 5 * time.Second

type (
	groupService interface {
		Create(ctx context.Context, name string, userID int) (int, error)
		UpdateName(ctx context.Context, name string, groupID, userID int) error
		Remove(ctx context.Context, groupID, userID int) error
		RemoveParticipants(ctx context.Context, groupID, userID int, participants []int) error
		GenerateInviteCode(ctx context.Context, groupID, userID int) (string, error)
		Join(ctx context.Context, userID int, code string) (Joined, error)
		GetByID(ctx context.Context, userID, groupID int) (Info, error)
		GetMany(ctx context.Context, userID int) ([]InfoWithoutParticipants, error)
	}
	authenticationMiddleware interface {
		VerifyToken(r *http.Request) (*http.Request, error)
	}
)

type Handler struct {
	logger                   commoninterfaces.Logger
	groupService             groupService
	authenticationMiddleware authenticationMiddleware
}

func NewHandler(
	groupService groupService,
	authenticationMiddleware authenticationMiddleware,
	logger commoninterfaces.Logger,
) *Handler {
	return &Handler{
		groupService:             groupService,
		authenticationMiddleware: authenticationMiddleware,
		logger:                   logger,
	}
}

func (gh *Handler) handleTimeout(err error, path string) error {
	if errors.Is(err, context.DeadlineExceeded) {
		gh.logger.Info("request timed out", "path", path)
		return commonerrors.RequestTimeout()
	}
	return err
}

func (gh *Handler) authenticateAndGetUserID(r *http.Request) (int, error) {
	r, err := gh.authenticationMiddleware.VerifyToken(r)
	if err != nil {
		return 0, gh.handleTimeout(err, r.URL.Path)
	}

	userID, err := request.ReadUserIDFromToken(r)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (gh *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	userID, err := gh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	reqData, err := request.ReadAndValidateBody[Input](r)
	if err != nil {
		return err
	}

	id, err := gh.groupService.Create(ctx, reqData.Name, userID)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusCreated, map[string]any{
		"group": map[string]any{
			"id":   id,
			"name": reqData.Name,
		},
	})

	return nil
}

func (gh *Handler) UpdateName(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	userID, err := gh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	groupID, err := strconv.Atoi(r.PathValue("groupID"))
	if err != nil || groupID <= 0 {
		return commonerrors.Validation("invalid group ID")
	}

	reqData, err := request.ReadAndValidateBody[Input](r)
	if err != nil {
		return err
	}

	err = gh.groupService.UpdateName(ctx, reqData.Name, groupID, userID)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}

func (gh *Handler) Remove(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	userID, err := gh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	groupID, err := strconv.Atoi(r.PathValue("groupID"))
	if err != nil {
		return err
	}

	err = gh.groupService.Remove(ctx, groupID, userID)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)

	return nil
}

func (gh *Handler) RemoveParticipants(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	userID, err := gh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	groupID, err := strconv.Atoi(r.PathValue("groupID"))
	if err != nil {
		return err
	}

	reqData, err := request.ReadAndValidateBody[RemoveParticipantsRequest](r)
	if err != nil {
		return err
	}

	err = gh.groupService.RemoveParticipants(ctx, groupID, userID, reqData.UserIDs)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusNoContent, nil)
	return nil
}

func (gh *Handler) GenerateInviteCode(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	userID, err := gh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	groupID, err := strconv.Atoi(r.PathValue("groupID"))
	if err != nil {
		return err
	}

	code, err := gh.groupService.GenerateInviteCode(ctx, groupID, userID)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]string{
		"code": code,
	})

	return nil
}

func (gh *Handler) Join(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	userID, err := gh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	activationCode := strings.TrimSpace(r.URL.Query().Get("code"))
	if activationCode == "" {
		return commonerrors.Validation("activation code is required")
	}

	group, err := gh.groupService.Join(ctx, userID, activationCode)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]any{
		"group": group,
	})

	return nil
}

func (gh *Handler) GetByID(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	userID, err := gh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	groupID, err := strconv.Atoi(r.PathValue("groupID"))
	if err != nil || groupID <= 0 {
		return commonerrors.Validation("invalid group ID")
	}

	group, err := gh.groupService.GetByID(ctx, userID, groupID)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]any{
		"group": group,
	})

	return nil
}

func (gh *Handler) GetMany(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := context.WithTimeout(r.Context(), groupTimeout)
	defer cancel()

	userID, err := gh.authenticateAndGetUserID(r)
	if err != nil {
		return err
	}

	groups, err := gh.groupService.GetMany(ctx, userID)
	if err != nil {
		return gh.handleTimeout(err, r.URL.Path)
	}

	response.Send(w, http.StatusOK, map[string]any{
		"groups": groups,
	})

	return nil
}
