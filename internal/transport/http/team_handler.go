package http

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/juzu400/avito-internship/internal/domain"
	"github.com/juzu400/avito-internship/internal/service"
)

func (h *Handler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req TeamDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Warn("AddTeam: invalid json", slog.Any("err", err))
		writeError(w, http.StatusBadRequest, service.ErrCodeValidation, "invalid json")
		return
	}

	team := &domain.Team{
		Name:    req.TeamName,
		Members: make([]domain.User, 0, len(req.Members)),
	}
	for _, m := range req.Members {
		team.Members = append(team.Members, domain.User{
			ID:       domain.UserID(m.UserID),
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	if err := h.services.Teams.UpsertTeam(r.Context(), team); err != nil {
		status, code := mapErrorToHTTP(err)
		writeError(w, status, code, err.Error())
		return
	}

	resp := TeamDTO{
		TeamName: team.Name,
		Members:  make([]TeamMemberDTO, 0, len(team.Members)),
	}

	for _, m := range team.Members {
		resp.Members = append(resp.Members, TeamMemberDTO{
			UserID:   string(m.ID),
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeError(w, http.StatusBadRequest, service.ErrCodeValidation, "team_name is required")
		return
	}

	team, err := h.services.Teams.GetByName(r.Context(), teamName)
	if err != nil {
		status, code := mapErrorToHTTP(err)
		writeError(w, status, code, err.Error())
		return
	}

	resp := TeamDTO{
		TeamName: team.Name,
		Members:  make([]TeamMemberDTO, 0, len(team.Members)),
	}
	for _, m := range team.Members {
		resp.Members = append(resp.Members, TeamMemberDTO{
			UserID:   string(m.ID),
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
