package handlers

import (
	"encoding/json"
	"net/http"

	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	serverrors "github.com/wozhdeleniye/avito-tech-internship/internal/pkg/errors"
	"github.com/wozhdeleniye/avito-tech-internship/internal/repo/models"
	"github.com/wozhdeleniye/avito-tech-internship/internal/services"
)

type MainAPI struct {
	PRService   *services.PReqService
	TeamService *services.TeamService
	AuthService *services.AuthService
}

func ErrorConstructor(code openapi.ErrorResponseErrorCode, message string) (Error struct {
	Code    openapi.ErrorResponseErrorCode `json:"code"`
	Message string                         `json:"message"`
}) {
	return struct {
		Code    openapi.ErrorResponseErrorCode `json:"code"`
		Message string                         `json:"message"`
	}{Code: code, Message: message}
}

// Создать PR и автоматически назначить до 2 ревьюверов из команды автора
func (h MainAPI) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var req openapi.PostPullRequestCreateJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	pr, serr := h.PRService.CreatePullRequest(r.Context(), req)
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

	if pr == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serverrors.ErrUserNotFound.Code, "author or team not found")})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]*openapi.PullRequest{"pr": pr})
}

// Пометить PR как MERGED (идемпотентная операция)
func (h MainAPI) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var req openapi.PostPullRequestMergeJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.PullRequestId == "" {
		http.Error(w, "pull_request_id is required", http.StatusBadRequest)
		return
	}

	pr, serr := h.PRService.MarkPullReqAsMerged(r.Context(), req.PullRequestId)
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

	if pr == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serverrors.ErrPRNotFound.Code, "pull request not found")})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]*openapi.PullRequest{"pr": pr})
}

// Переназначить конкретного ревьювера на другого из его команды
func (h MainAPI) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var req openapi.PostPullRequestReassignJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.PullRequestId == "" || req.OldUserId == "" {
		http.Error(w, "pull_request_id and old_user_id are required", http.StatusBadRequest)
		return
	}

	resp, serr := h.PRService.ReassignReviewer(r.Context(), req.PullRequestId, req.OldUserId)
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

	if resp == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serverrors.ErrPRNotFound.Code, "pull request or user not found")})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Создать команду с участниками (создаёт/обновляет пользователей)
func (h MainAPI) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var req openapi.Team
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	team, serr := h.TeamService.CreateTeam(r.Context(), req)
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

	if team == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serverrors.ErrPRNotFound.Code, "can`t create team")})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]*openapi.Team{"team": team})
}

// Получить команду с участниками
func (h MainAPI) GetTeamGet(w http.ResponseWriter, r *http.Request, params openapi.GetTeamGetParams) {
	team, serr := h.TeamService.GetTeamQuery(r.Context(), params.TeamName)
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

	if team == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serverrors.ErrTeamNotFound.Code, "team not found")})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(team)
}

// Получить PR'ы, где пользователь назначен ревьювером
func (h MainAPI) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params openapi.GetUsersGetReviewParams) {
	prSearch, serr := h.PRService.GetPullReqsByReviever(r.Context(), params.UserId)
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

	if prSearch == nil {
		prSearch = &models.PullRequestSearch{
			PullRequest: make([]*openapi.PullRequestShort, 0),
			Author:      params.UserId,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(prSearch)
}

// Установить флаг активности пользователя
func (h MainAPI) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	user, serr := h.AuthService.SetUserActive(r.Context(), req.UserID, req.IsActive)
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

	if user == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(openapi.ErrorResponse{Error: ErrorConstructor(serverrors.ErrUserNotFound.Code, "user not found")})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]*models.User{"user": user})
}
