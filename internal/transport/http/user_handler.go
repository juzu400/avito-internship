package http

import (
	"encoding/json"
	"errors"
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

	userID := domain.UserID(req.UserID)

	if err := h.services.Users.SetIsActive(r.Context(), userID, req.IsActive); err != nil {
		status, code := mapErrorToHTTP(err)
		writeError(w, status, code, err.Error())
		return
	}

	u, err := h.services.Users.GetByID(r.Context(), userID)
	if err != nil {
		status, code := mapErrorToHTTP(err)
		writeError(w, status, code, err.Error())
		return
	}

	var teamName string
	team, err := h.services.Teams.GetByMemberID(r.Context(), userID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			status, code := mapErrorToHTTP(err)
			writeError(w, status, code, err.Error())
			return
		}
	} else {
		teamName = team.Name
	}

	resp := UserResponse{
		User: UserDTO{
			UserID:   string(u.ID),
			Username: u.Username,
			TeamName: teamName,
			IsActive: u.IsActive,
		},
	}

	writeJSON(w, http.StatusOK, resp)
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
