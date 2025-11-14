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

func TestSetUserActive_InvalidJSON(t *testing.T) {
	h, _, _, _ := newTestHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(`{`))

	h.SetUserActive(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	code, _ := decodeError(t, rr)
	if code != codeValidationErr {
		t.Fatalf("expected error code VALIDATION_ERROR, got %q", code)
	}
}

func TestSetUserActive_Success(t *testing.T) {
	h, userRepo, _, _ := newTestHandler(t)

	body := `{"user_id": "u1", "is_active": true}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", strings.NewReader(body))

	userRepo.EXPECT().
		SetIsActive(gomock.Any(), domain.UserID("u1"), true).
		Return(nil)

	h.SetUserActive(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp struct {
		User struct {
			UserID   string `json:"user_id"`
			IsActive bool   `json:"is_active"`
		} `json:"user"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if resp.User.UserID != "u1" || !resp.User.IsActive {
		t.Fatalf("unexpected body: %+v", resp.User)
	}
}

func TestGetUserReview_MissingUserID(t *testing.T) {
	h, _, _, _ := newTestHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/getReview", nil)

	h.GetUserReview(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	code, _ := decodeError(t, rr)
	if code != codeValidationErr {
		t.Fatalf("expected error code VALIDATION_ERROR, got %q", code)
	}
}

func TestGetUserReview_UserNotFound(t *testing.T) {
	h, userRepo, _, _ := newTestHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u10", nil)

	userRepo.EXPECT().
		GetByID(gomock.Any(), domain.UserID("u10")).
		Return(nil, domain.ErrUserNotFound)

	h.GetUserReview(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	code, _ := decodeError(t, rr)
	if code != codeNotFound {
		t.Fatalf("expected error code NOT_FOUND, got %q", code)
	}
}
