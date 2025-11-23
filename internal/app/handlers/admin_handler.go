package handlers

import (
	"encoding/json"
	"net/http"

	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	"github.com/wozhdeleniye/avito-tech-internship/internal/services"
)

type AdminAPI struct {
	PRService   *services.PReqService
	TeamService *services.TeamService
}

// GET /admin/stats
// Статистика по количеству назначений ревьюером на пользователя и количество ревьюеров на PR
func (h AdminAPI) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	userCounts, serr := h.PRService.CountAssignmentsPerUser(r.Context())
	if serr != nil {
		w.Header().Set("Content-Type", "application/json")
		status := serr.HTTPCode
		if status == 0 {
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serr.Code, serr.Message)})
		return
	}

	prCounts, serr := h.PRService.CountAssignmentsPerPR(r.Context())
	if serr != nil {
		w.Header().Set("Content-Type", "application/json")
		status := serr.HTTPCode
		if status == 0 {
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serr.Code, serr.Message)})
		return
	}

	resp := map[string]interface{}{
		"assignments_per_user": userCounts,
		"assignments_per_pr":   prCounts,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// POST /admin/team/deactivate
// Деактивирует всех участников выбранной команды и переназначает все PR на участников новой команды
func (h AdminAPI) PostAdminTeamDeactivate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OldTeamName string `json:"old_team_name"`
		NewTeamName string `json:"new_team_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.OldTeamName == "" || req.NewTeamName == "" {
		http.Error(w, "old_team_name and new_team_name are required", http.StatusBadRequest)
		return
	}

	result, serr := h.TeamService.MassDeactivateTeam(r.Context(), req.OldTeamName, req.NewTeamName)
	if serr != nil {
		w.Header().Set("Content-Type", "application/json")
		status := serr.HTTPCode
		if status == 0 {
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serr.Code, serr.Message)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}
