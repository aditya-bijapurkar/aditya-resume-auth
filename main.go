package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/aditya-bijapurkar/aditya-resume-auth/handlers"
	"github.com/aditya-bijapurkar/aditya-resume-auth/middleware/auth"
	"github.com/aditya-bijapurkar/aditya-resume-auth/models"
	"github.com/aditya-bijapurkar/aditya-resume-auth/services"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := initDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := models.InitDB(db); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	userStore := models.NewUserStore(db)
	handlers.SetUserStore(userStore)

	emailService := services.NewEmailService()
	handlers.SetEmailService(emailService)

	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/signup", handlers.Signup)
		r.Post("/login", handlers.Login)
		r.Post("/logout", auth.RequireAuth(handlers.Logout))
		r.Get("/me", auth.RequireAuth(handlers.GetMe))
		r.Get("/verify", handlers.VerifyEmail)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func initDB() (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
		dbPort := os.Getenv("DB_PORT")
		if dbPort == "" {
			dbPort = "5432"
		}
		dbUser := os.Getenv("DB_USER")
		if dbUser == "" {
			dbUser = "postgres"
		}
		dbPassword := os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			dbPassword = "postgres"
		}
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "resume_auth"
		}
		dbSchema := os.Getenv("DB_SCHEMA")
		if dbSchema == "" {
			dbSchema = "auth"
		}

		dbURL = "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable&search_path=" + dbSchema
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}
