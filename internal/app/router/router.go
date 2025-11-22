package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	"github.com/wozhdeleniye/avito-tech-internship/internal/app/handlers"
	"github.com/wozhdeleniye/avito-tech-internship/internal/app/middleware"
	"github.com/wozhdeleniye/avito-tech-internship/internal/services"
)

func NewAuthRouter(authHandler *handlers.AuthHandler, authMiddleware *middleware.AuthMiddleware) chi.Router {
	r := chi.NewRouter()

	r.Post("/login", authHandler.Login)
	r.Post("/register", authHandler.Register)

	r.Group(func(protected chi.Router) {
		protected.Use(authMiddleware.Authenticate)
		protected.Post("/refresh", authHandler.Refresh)
		protected.Post("/logout", authHandler.Logout)
	})

	return r
}

func NewApp(authService *services.AuthService) http.Handler {
	r := chi.NewRouter()

	middleware := middleware.NewAuthMiddleware(authService)

	mainRouter := chi.NewRouter()
	mainHandler := handlers.MainAPI{}
	mainRouter.Use(middleware.Authenticate)
	openapi.HandlerFromMux(mainHandler, mainRouter)

	authHandler := handlers.NewAuthHandler(authService)
	authRouter := NewAuthRouter(authHandler, middleware)

	r.Mount("/api", mainRouter)
	r.Mount("/api/auth", authRouter)

	return r
}
