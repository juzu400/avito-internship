package http

import (
	"encoding/json"
	"net/http"

	"github.com/juzu400/avito-internship/internal/domain"
)

func (h *Handler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req TeamDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	team := &domain.Team{
		Name: req.TeamName,
		// ...
	}

	if err := h.services.Teams.UpsertTeam(r.Context(), team); err != nil {
		h.log.Error("add team failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, req)
}
