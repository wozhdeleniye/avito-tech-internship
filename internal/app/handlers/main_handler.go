package handlers

import (
	"encoding/json"
	"net/http"

	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
)

type MainAPI struct{}

// Создать PR и автоматически назначить до 2 ревьюверов из команды автора
func (h MainAPI) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	// TODO: вызвать сервис создания PR
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "PR создан (заглушка)"})
}

// Пометить PR как MERGED (идемпотентная операция)
func (h MainAPI) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	// TODO: вызвать сервис merge PR
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "PR merged (заглушка)"})
}

// Переназначить конкретного ревьювера на другого из его команды
func (h MainAPI) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	// TODO: вызвать сервис переназначения ревьювера
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Reviewer reassigned (заглушка)"})
}

// Создать команду с участниками (создаёт/обновляет пользователей)
func (h MainAPI) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	// TODO: вызвать сервис создания команды
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Team created (заглушка)"})
}

// Получить команду с участниками
func (h MainAPI) GetTeamGet(w http.ResponseWriter, r *http.Request, params openapi.GetTeamGetParams) {
	// TODO: вызвать сервис получения команды
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Team info (заглушка)"})
}

// Получить PR'ы, где пользователь назначен ревьювером
func (h MainAPI) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params openapi.GetUsersGetReviewParams) {
	// TODO: вызвать сервис получения PR для ревьювера
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User reviews (заглушка)"})
}

// Установить флаг активности пользователя
func (h MainAPI) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	// TODO: вызвать сервис установки активности пользователя
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User activity set (заглушка)"})
}
