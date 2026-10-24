package poll

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type routeTestHandler struct {
	createCalls        int
	submitAnswersCalls int
	pollID             string
}

func (handler *routeTestHandler) Create(w http.ResponseWriter, _ *http.Request) error {
	handler.createCalls++
	w.WriteHeader(http.StatusCreated)
	return nil
}

func (handler *routeTestHandler) SubmitAnswers(w http.ResponseWriter, r *http.Request) error {
	handler.submitAnswersCalls++
	handler.pollID = r.PathValue("pollID")
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type routeTestLogger struct{}

func (routeTestLogger) Info(string, ...any)  {}
func (routeTestLogger) Debug(string, ...any) {}
func (routeTestLogger) Error(string, ...any) {}

func TestRoutesSetup(t *testing.T) {
	handler := &routeTestHandler{}
	router := http.NewServeMux()
	NewRoutes(handler, routeTestLogger{}).Setup(router)

	createRequest := httptest.NewRequest(http.MethodPost, "/polls/", nil)
	createResponse := httptest.NewRecorder()
	router.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResponse.Code, http.StatusCreated)
	}
	if handler.createCalls != 1 || handler.submitAnswersCalls != 0 {
		t.Fatalf(
			"after create: Create calls = %d, SubmitAnswers calls = %d",
			handler.createCalls,
			handler.submitAnswersCalls,
		)
	}

	answersRequest := httptest.NewRequest(http.MethodPost, "/polls/42/answers", nil)
	answersResponse := httptest.NewRecorder()
	router.ServeHTTP(answersResponse, answersRequest)

	if answersResponse.Code != http.StatusNoContent {
		t.Fatalf("answers status = %d, want %d", answersResponse.Code, http.StatusNoContent)
	}
	if handler.submitAnswersCalls != 1 {
		t.Fatalf("SubmitAnswers calls = %d, want 1", handler.submitAnswersCalls)
	}
	if handler.pollID != "42" {
		t.Fatalf("pollID = %q, want %q", handler.pollID, "42")
	}
}
