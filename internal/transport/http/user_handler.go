package http

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/juzu400/avito-internship/internal/domain"
	"github.com/juzu400/avito-internship/internal/service"
)

func (h *Handler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	var req SetUserActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("SetUserActive: invalid json", slog.Any("err", err))
		writeError(w, http.StatusBadRequest, service.ErrCodeValidation, "invalid json")
		return
	}

	err := h.services.Users.SetIsActive(r.Context(), domain.UserID(req.UserID), req.IsActive)
	if err != nil {
		status, code := mapErrorToHTTP(err)
		h.log.Error("SetUserActive failed",
			slog.String("user_id", req.UserID),
			slog.String("error_code", code),
			slog.Any("err", err),
		)
		writeError(w, status, code, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":   req.UserID,
		"is_active": req.IsActive,
	})
}

func (h *Handler) GetUserReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, service.ErrCodeValidation, "user_id is required")
		return
	}

	prs, err := h.services.Users.GetReviews(r.Context(), domain.UserID(userID))
	if err != nil {
		status, code := mapErrorToHTTP(err)
		h.log.Error("GetUserReview failed",
			slog.String("user_id", userID),
			slog.String("error_code", code),
			slog.Any("err", err),
		)
		writeError(w, status, code, err.Error())
		return
	}

	resp := GetUserReviewResponse{
		UserID:       userID,
		PullRequests: make([]PullRequestShortDTO, 0, len(prs)),
	}

	for _, pr := range prs {
		resp.PullRequests = append(resp.PullRequests, PullRequestShortDTO{
			PullRequestID:   string(pr.ID),
			PullRequestName: pr.Name,
			AuthorID:        string(pr.AuthorID),
			Status:          string(pr.Status),
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
