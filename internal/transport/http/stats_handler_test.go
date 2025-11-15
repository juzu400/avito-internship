package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/juzu400/avito-internship/internal/domain"
	"github.com/juzu400/avito-internship/internal/service"
)

func TestGetReviewerStats_Success(t *testing.T) {
	h, _, _, prRepo := newTestHandler(t)

	stats := []domain.ReviewerAssignmentStat{
		{ReviewerID: domain.UserID("u1"), AssignmentsCount: 3},
		{ReviewerID: domain.UserID("u2"), AssignmentsCount: 1},
	}

	prRepo.EXPECT().
		GetReviewerAssignmentStats(gomock.Any()).
		Return(stats, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats/reviewers", nil)

	h.GetReviewerStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp ReviewerStatsResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Items) != len(stats) {
		t.Fatalf("expected %d items, got %d", len(stats), len(resp.Items))
	}

	if resp.Items[0].ReviewerID != string(stats[0].ReviewerID) ||
		resp.Items[0].Assignments != stats[0].AssignmentsCount {
		t.Fatalf("unexpected first item: %+v", resp.Items[0])
	}
}

func TestGetReviewerStats_InternalError(t *testing.T) {
	h, _, _, prRepo := newTestHandler(t)

	prRepo.EXPECT().
		GetReviewerAssignmentStats(gomock.Any()).
		Return(nil, errors.New("db error"))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats/reviewers", nil)

	h.GetReviewerStats(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	code, _ := decodeError(t, rr)
	if code != service.ErrCodeInternal {
		t.Fatalf("expected error code %q, got %q", service.ErrCodeInternal, code)
	}
}

func TestGetPullRequestStats_Success(t *testing.T) {
	h, _, _, prRepo := newTestHandler(t)

	stats := []domain.PullRequestReviewersStat{
		{PullRequestID: domain.PullRequestID("pr-1"), ReviewersCount: 2},
		{PullRequestID: domain.PullRequestID("pr-2"), ReviewersCount: 3},
	}

	prRepo.EXPECT().
		GetPullRequestReviewerStats(gomock.Any()).
		Return(stats, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats/pullRequests", nil)

	h.GetPullRequestStats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp PullRequestStatsResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Items) != len(stats) {
		t.Fatalf("expected %d items, got %d", len(stats), len(resp.Items))
	}

	if resp.Items[0].PullRequestID != string(stats[0].PullRequestID) ||
		resp.Items[0].Reviewers != stats[0].ReviewersCount {
		t.Fatalf("unexpected first item: %+v", resp.Items[0])
	}
}

func TestGetPullRequestStats_Error(t *testing.T) {
	h, _, _, prRepo := newTestHandler(t)

	prRepo.EXPECT().
		GetPullRequestReviewerStats(gomock.Any()).
		Return(nil, errors.New("db error"))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/stats/pullRequests", nil)

	h.GetPullRequestStats(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	code, _ := decodeError(t, rr)
	if code != service.ErrCodeInternal {
		t.Fatalf("expected error code %q, got %q", service.ErrCodeInternal, code)
	}
}
