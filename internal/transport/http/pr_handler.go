package http

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/juzu400/avito-internship/internal/domain"
	"github.com/juzu400/avito-internship/internal/service"
)

func (h *Handler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	var req CreatePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("CreatePullRequest: invalid json", slog.Any("err", err))
		writeError(w, http.StatusBadRequest, service.ErrCodeValidation, "invalid json")
		return
	}

	pr, err := h.services.PullRequests.Create(
		r.Context(),
		domain.PullRequestID(req.PullRequestID),
		req.PullRequestName,
		domain.UserID(req.AuthorID),
	)
	if err != nil {
		status, code := mapErrorToHTTP(err)
		h.log.Error("CreatePullRequest failed",
			slog.String("pull_request_id", req.PullRequestID),
			slog.String("author_id", req.AuthorID),
			slog.String("error_code", code),
			slog.Any("err", err),
		)
		writeError(w, status, code, err.Error())
		return
	}

	resp := toPullRequestDTO(pr)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	var req MergePullRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("MergePullRequest: invalid json", slog.Any("err", err))
		writeError(w, http.StatusBadRequest, service.ErrCodeValidation, "invalid json")
		return
	}

	pr, err := h.services.PullRequests.Merge(r.Context(), domain.PullRequestID(req.PullRequestID))
	if err != nil {
		status, code := mapErrorToHTTP(err)
		h.log.Error("MergePullRequest failed",
			slog.String("pull_request_id", req.PullRequestID),
			slog.String("error_code", code),
			slog.Any("err", err),
		)
		writeError(w, status, code, err.Error())
		return
	}

	resp := toPullRequestDTO(pr)
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("ReassignReviewer: invalid json", slog.Any("err", err))
		writeError(w, http.StatusBadRequest, service.ErrCodeValidation, "invalid json")
		return
	}

	pr, newReviewer, err := h.services.PullRequests.ReassignReviewer(
		r.Context(),
		domain.PullRequestID(req.PullRequestID),
		domain.UserID(req.OldUserID),
	)
	if err != nil {
		status, code := mapErrorToHTTP(err)
		h.log.Error("ReassignReviewer failed",
			slog.String("pull_request_id", req.PullRequestID),
			slog.String("old_user_id", req.OldUserID),
			slog.String("error_code", code),
			slog.Any("err", err),
		)
		writeError(w, status, code, err.Error())
		return
	}

	resp := ReassignReviewerResponse{
		PullRequest: toPullRequestDTO(pr),
		ReplacedBy: TeamMemberDTO{
			UserID:   string(newReviewer.ID),
			Username: newReviewer.Username,
			IsActive: newReviewer.IsActive,
		},
	}

	writeJSON(w, http.StatusOK, resp)
}

func toPullRequestDTO(pr *domain.PullRequest) PullRequestDTO {
	dto := PullRequestDTO{
		PullRequestID:     string(pr.ID),
		PullRequestName:   pr.Name,
		AuthorID:          string(pr.AuthorID),
		Status:            string(pr.Status),
		AssignedReviewers: make([]string, 0, len(pr.AssignedReviewers)),
		CreatedAt:         pr.CreatedAt,
		MergedAt:          pr.MergedAt,
	}
	for _, rid := range pr.AssignedReviewers {
		dto.AssignedReviewers = append(dto.AssignedReviewers, string(rid))
	}
	return dto
}
