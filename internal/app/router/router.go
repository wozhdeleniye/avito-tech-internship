package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	"github.com/wozhdeleniye/avito-tech-internship/internal/app/handlers"
	"github.com/wozhdeleniye/avito-tech-internship/internal/services"
)

func NewApp(prService *services.PReqService, teamService *services.TeamService) http.Handler {
	r := chi.NewRouter()
	r.Use(CORSMiddleware())

	//middleware := middleware.NewAuthMiddleware(authService)

	mainRouter := chi.NewRouter()
	mainHandler := handlers.MainAPI{
		PRService:   prService,
		TeamService: teamService,
	}
	openapi.HandlerFromMux(mainHandler, mainRouter)

	r.Mount("/api", mainRouter)

	adminRouter := chi.NewRouter()
	adminHandler := handlers.AdminAPI{
		PRService:   prService,
		TeamService: teamService,
	}
	adminRouter.Get("/stats", adminHandler.GetAdminStats)
	adminRouter.Post("/team/deactivate", adminHandler.PostAdminTeamDeactivate)

	r.Mount("/api/admin", adminRouter)

	return r
}

func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, User-id")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
