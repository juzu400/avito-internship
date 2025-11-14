package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/juzu400/avito-internship/internal/domain"
)

func TestCreatePullRequest_InvalidJSON(t *testing.T) {
	h, _, _, _ := newTestHandler(t)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", strings.NewReader(`{`))

	h.CreatePullRequest(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	code, _ := decodeError(t, rr)
	if code != codeValidationErr {
		t.Fatalf("expected error code VALIDATION_ERROR, got %q", code)
	}
}

func TestReassignReviewer_MergedPR_ReturnsConflict(t *testing.T) {
	h, _, _, prRepo := newTestHandler(t)

	pr := &domain.PullRequest{
		ID:     "pr-1",
		Status: domain.PRStatusMerged,
	}
	prRepo.EXPECT().
		GetByID(gomock.Any(), domain.PullRequestID("pr-1")).
		Return(pr, nil)

	body := `{"pull_request_id": "pr-1", "old_user_id": "u2"}`
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", strings.NewReader(body))

	h.ReassignReviewer(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
	code, _ := decodeError(t, rr)
	if code != "PR_MERGED" {
		t.Fatalf("expected error code PR_MERGED, got %q", code)
	}
}
