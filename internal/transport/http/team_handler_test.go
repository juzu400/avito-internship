package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/juzu400/avito-internship/internal/domain"
)

func TestAddTeam_InvalidJSON(t *testing.T) {
	h, _, _, _ := newTestHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader(`{invalid json`))

	h.AddTeam(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	code, _ := decodeError(t, rr)
	if code != codeValidationErr {
		t.Fatalf("expected error code VALIDATION_ERROR, got %q", code)
	}
}

func TestAddTeam_Success(t *testing.T) {
	h, _, teamRepo, _ := newTestHandler(t)

	body := `{
		"team_name": "backend",
		"members": [
		  {"user_id": "u1", "username": "Alice", "is_active": true}
		]
	}`

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/team/add", strings.NewReader(body))

	userID := domain.UserID("u1")

	gomock.InOrder(
		teamRepo.EXPECT().
			GetByMemberID(gomock.Any(), userID).
			Return(nil, domain.ErrNotFound),

		teamRepo.EXPECT().
			UpsertTeam(gomock.Any(), gomock.AssignableToTypeOf(&domain.Team{})).
			Return(nil),
	)

	h.AddTeam(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var resp TeamDTO
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.TeamName != "backend" {
		t.Fatalf("expected team_name %q, got %q", "backend", resp.TeamName)
	}
	if len(resp.Members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(resp.Members))
	}
	m := resp.Members[0]
	if m.UserID != "u1" || m.Username != "Alice" || !m.IsActive {
		t.Fatalf("unexpected member: %+v", m)
	}
}

func TestGetTeam_MissingTeamName(t *testing.T) {
	h, _, _, _ := newTestHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/team/get", nil)

	h.GetTeam(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	code, _ := decodeError(t, rr)
	if code != codeValidationErr {
		t.Fatalf("expected error code VALIDATION_ERROR, got %q", code)
	}
}

func TestGetTeam_NotFound(t *testing.T) {
	h, _, teamRepo, _ := newTestHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)

	teamRepo.EXPECT().
		GetByName(gomock.Any(), "backend").
		Return(nil, domain.ErrNotFound)

	h.GetTeam(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	code, _ := decodeError(t, rr)
	if code != codeNotFound {
		t.Fatalf("expected error code NOT_FOUND, got %q", code)
	}
}
