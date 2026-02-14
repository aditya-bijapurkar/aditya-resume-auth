package router

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/aditya-bijapurkar/aditya-resume-auth/handlers"
	"github.com/aditya-bijapurkar/aditya-resume-auth/middleware/auth"
	"github.com/aditya-bijapurkar/aditya-resume-auth/models"
	"github.com/aditya-bijapurkar/aditya-resume-auth/services"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func setHandlers(db *sql.DB) {
	handlers.SetUserStore(models.NewUserStore(db))
	handlers.SetEmailService(services.NewEmailService())
}

func setMiddleware(router *chi.Mux) {
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{os.Getenv("REDIRECT_URL")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func setRoutes(router *chi.Mux) {
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router.Route("/auth", func(r chi.Router) {
		r.Post("/signup", auth.AddAuthTokenContext(handlers.Signup))
		r.Post("/login", handlers.Login)
		r.Post("/logout", handlers.Logout)
		r.Get("/me", auth.RequireAuth(handlers.GetMe))
		r.Get("/verify", handlers.VerifyEmail)
		r.Post("/session/user", handlers.CreateSessionUser)
	})

	router.Route("/internal", func(r chi.Router) {
		r.Post("/create/users", auth.VerifyTokenSignature(handlers.CreateAnonymousUsers))
	})
}

func NewRouter(db *sql.DB) http.Handler {
	setHandlers(db)

	router := chi.NewRouter()

	setMiddleware(router)
	setRoutes(router)

	return router
}
