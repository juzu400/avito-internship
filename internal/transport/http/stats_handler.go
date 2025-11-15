package http

import (
	"log/slog"
	"net/http"
)

// ReviewerStatsItemDTO represents statistics for a single reviewer.
func (h *Handler) GetReviewerStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.services.PullRequests.GetReviewerAssignmentStats(ctx)
	if err != nil {
		status, code := mapErrorToHTTP(err)
		h.log.Error("GetReviewerStats failed",
			slog.String("handler", "GetReviewerStats"),
			slog.String("error_code", code),
			slog.Any("err", err),
		)
		writeError(w, status, code, "failed to get reviewer statistics")
		return
	}

	resp := ReviewerStatsResponse{
		Items: make([]ReviewerStatsItemDTO, 0, len(stats)),
	}

	for _, s := range stats {
		resp.Items = append(resp.Items, ReviewerStatsItemDTO{
			ReviewerID:  string(s.ReviewerID),
			Assignments: s.AssignmentsCount,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// PullRequestStatsItemDTO represents statistics for pull requests.
func (h *Handler) GetPullRequestStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.services.PullRequests.GetPullRequestReviewerStats(ctx)
	if err != nil {
		status, code := mapErrorToHTTP(err)
		h.log.Error("GetPullRequestStats failed",
			slog.String("handler", "GetPullRequestStats"),
			slog.String("error_code", code),
			slog.Any("err", err),
		)
		writeError(w, status, code, "failed to get pull request statistics")
		return
	}

	resp := PullRequestStatsResponse{
		Items: make([]PullRequestStatsItemDTO, 0, len(stats)),
	}

	for _, s := range stats {
		resp.Items = append(resp.Items, PullRequestStatsItemDTO{
			PullRequestID: string(s.PullRequestID),
			Reviewers:     s.ReviewersCount,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
