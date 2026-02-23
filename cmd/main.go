package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite" // registers the "sqlite" driver

	"user-management-api/internal/config"
	"user-management-api/internal/handler"
	"user-management-api/internal/middleware"
	"user-management-api/internal/repository"
	"user-management-api/internal/service"
)

func main() {
	cfg := config.Load()

	// Ensure the data directory exists before opening the DB file.
	if err := os.MkdirAll("data", 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Dependency wiring — pure constructor injection, no global state.
	userRepo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
	authHandler := handler.NewAuthHandler(userSvc)
	userHandler := handler.NewUserHandler(userSvc)

	r := gin.Default()
	r.SetTrustedProxies(nil) //nolint:errcheck

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/signin", authHandler.SignIn)
		}

		// All /users routes require a valid JWT.
		users := v1.Group("/users", middleware.JWTAuth(cfg.JWTSecret))
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
		}
	}

	log.Printf("server listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("run server: %v", err)
	}
}

// runMigrations creates the schema on first run. Idempotent.
func runMigrations(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id            TEXT PRIMARY KEY,
			name          TEXT NOT NULL,
			email         TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at    TEXT NOT NULL,
			updated_at    TEXT NOT NULL
		)
	`)
	return err
}
