package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/juzu400/avito-internship/internal/repository"
	"github.com/juzu400/avito-internship/internal/repository/mocks"
	"github.com/juzu400/avito-internship/internal/service"
)

const (
	contentTypeJSON   = "application/json"
	codeValidationErr = "VALIDATION_ERROR"
	codeNotFound      = "NOT_FOUND"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestHandler(t *testing.T) (*Handler,
	*mocks.MockUserRepository,
	*mocks.MockTeamRepository,
	*mocks.MockPullRequestRepository,
) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	userRepo := mocks.NewMockUserRepository(ctrl)
	teamRepo := mocks.NewMockTeamRepository(ctrl)
	prRepo := mocks.NewMockPullRequestRepository(ctrl)

	repos := &repository.Repositories{
		Users:        userRepo,
		Teams:        teamRepo,
		PullRequests: prRepo,
	}

	log := newTestLogger()
	services := service.NewServices(log, repos)

	h := &Handler{
		log:      log.With(slog.String("layer", "http")),
		services: services,
	}

	return h, userRepo, teamRepo, prRepo
}

func decodeError(t *testing.T, rr *httptest.ResponseRecorder) (code, message string) {
	t.Helper()

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}

	return body.Error.Code, body.Error.Message
}

func TestWriteJSON_SetsStatusAndContentType(t *testing.T) {
	rr := httptest.NewRecorder()

	payload := map[string]string{"foo": "bar"}
	writeJSON(rr, http.StatusCreated, payload)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	if got := rr.Header().Get("Content-Type"); got != contentTypeJSON {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var decoded map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&decoded); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if decoded["foo"] != "bar" {
		t.Fatalf("expected body %v, got %v", payload, decoded)
	}
}

func TestWriteError_ProducesErrorEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()

	writeError(rr, http.StatusBadRequest, codeValidationErr, "invalid json")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != contentTypeJSON {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	code, msg := decodeError(t, rr)
	if code != codeValidationErr {
		t.Fatalf("expected code %q, got %q", codeValidationErr, code)
	}
	if msg != "invalid json" {
		t.Fatalf("expected message %q, got %q", "invalid json", msg)
	}
}
