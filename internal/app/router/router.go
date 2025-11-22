package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	openapi "github.com/wozhdeleniye/avito-tech-internship/api"
	"github.com/wozhdeleniye/avito-tech-internship/internal/app/handlers"
	"github.com/wozhdeleniye/avito-tech-internship/internal/services"
)

func NewAuthRouter(authHandler *handlers.AuthHandler) chi.Router {
	r := chi.NewRouter()

	r.Post("/login", authHandler.Login)
	r.Post("/register", authHandler.Register)

	//r.Group(func(protected chi.Router) {
	//	protected.Use(authMiddleware.Authenticate)
	//	protected.Post("/refresh", authHandler.Refresh)
	//	protected.Post("/logout", authHandler.Logout)
	//})

	return r
}

// Обновлённая сигнатура NewApp: принимаем сервисы, необходимые для main API.
func NewApp(authService *services.AuthService, prService *services.PReqService, teamService *services.TeamService) http.Handler {
	r := chi.NewRouter()
	r.Use(CORSMiddleware())

	//middleware := middleware.NewAuthMiddleware(authService)

	mainRouter := chi.NewRouter()
	mainHandler := handlers.MainAPI{
		PRService:   prService,
		TeamService: teamService,
		AuthService: authService,
	}
	//mainRouter.Use(middleware.Authenticate) //сначала подумал, что понадобится аутентификация, но потом прочитал документацию и таску и понял, что вроде бы не требуется
	openapi.HandlerFromMux(mainHandler, mainRouter)

	authHandler := handlers.NewAuthHandler(authService)
	authRouter := NewAuthRouter(authHandler)

	r.Mount("/api", mainRouter)
	r.Mount("/api/auth", authRouter)

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
