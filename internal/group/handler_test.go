package group_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	commonerrors "github.com/adrian-kurek/fit_the_meet/common/errors"
	"github.com/adrian-kurek/fit_the_meet/common/request"
	"github.com/adrian-kurek/fit_the_meet/internal/group"
)

type groupServiceStub struct {
	group group.Info
}

func (stub groupServiceStub) Create(context.Context, string, int) error { return nil }

func (stub groupServiceStub) UpdateName(context.Context, string, int, int) error { return nil }

func (stub groupServiceStub) Remove(context.Context, int, int) error { return nil }

func (stub groupServiceStub) RemoveParticipants(context.Context, int, int, []int) error { return nil }

func (stub groupServiceStub) GenerateInviteCode(context.Context, int, int) (string, error) {
	return "", nil
}

func (stub groupServiceStub) Join(context.Context, int, string) error { return nil }

func (stub groupServiceStub) GetGroupNamesForUser(context.Context, int) ([]string, error) {
	return nil, nil
}

func (stub groupServiceStub) GetByID(context.Context, int, int) (group.Info, error) {
	return stub.group, nil
}

type authenticationStub struct {
	userID int
}

func (stub authenticationStub) VerifyToken(r *http.Request) (*http.Request, error) {
	return request.SetContext(r, "id", stub.userID), nil
}

type loggerStub struct{}

func (loggerStub) Info(string, ...any)  {}
func (loggerStub) Debug(string, ...any) {}
func (loggerStub) Error(string, ...any) {}

func TestHandlerGetByID(t *testing.T) {
	t.Parallel()

	handler := group.NewHandler(
		groupServiceStub{group: group.Info{ID: 8, Name: "Project group", Role: "owner", MembersCount: 1}},
		authenticationStub{userID: 12},
		loggerStub{},
	)
	httpRequest := httptest.NewRequest(http.MethodGet, "/groups/8", nil)
	httpRequest.SetPathValue("groupID", "8")
	response := httptest.NewRecorder()

	if err := handler.GetByID(response, httpRequest); err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	var body struct {
		Group group.Info `json:"group"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body.Group.ID != 8 || body.Group.Role != "owner" {
		t.Fatalf("group = %+v, want ID 8 and owner role", body.Group)
	}
}

func TestHandlerGetByIDRejectsInvalidID(t *testing.T) {
	t.Parallel()

	handler := group.NewHandler(groupServiceStub{}, authenticationStub{userID: 12}, loggerStub{})
	httpRequest := httptest.NewRequest(http.MethodGet, "/groups/invalid", nil)
	httpRequest.SetPathValue("groupID", "invalid")

	err := handler.GetByID(httptest.NewRecorder(), httpRequest)
	var apiErr *commonerrors.API
	if !errors.As(err, &apiErr) {
		t.Fatalf("GetByID() error = %T, want *commonerrors.API", err)
	}
	if apiErr.Category != commonerrors.CategoryValidation {
		t.Fatalf("error category = %q, want %q", apiErr.Category, commonerrors.CategoryValidation)
	}
}
