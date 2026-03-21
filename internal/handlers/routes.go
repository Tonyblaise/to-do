package handlers

import (
	"database/sql"
	"net/http"
	"os"

	_ "github.com/Tonyblaise/to-do/docs"
	"github.com/Tonyblaise/to-do/internal/config"
	"github.com/Tonyblaise/to-do/internal/middleware"
	"github.com/Tonyblaise/to-do/internal/repository"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func RegisterRoutes(r *mux.Router, db *sql.DB, cfg *config.Config) {

	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	tagRepo := repository.NewTagRepository(db)
	attachRepo := repository.NewAttachmentRepository(db)

	hub := NewHub()

	authH := NewAuthHandler(userRepo, cfg)
	taskH := NewTaskHandler(taskRepo, hub)
	tagH := NewTagHandler(tagRepo)
	attachH := NewAttachmentHandler(attachRepo, taskRepo, cfg)

	authMW := middleware.Auth(cfg.JWTSecret)

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/health", HealthCheck(db)).Methods(http.MethodGet)

	api.HandleFunc("/auth/signup", authH.Signup).Methods(http.MethodPost)
	api.HandleFunc("/auth/login", authH.Login).Methods(http.MethodPost)

	protected := api.NewRoute().Subrouter{}

	protected.Use(authMW)

	protected.HandleFunc("/tasks", taskH.Create).Methods(http.MethodPost)
	protected.HandleFunc("/tasks", taskH.List).Methods(http.MethodGet)
	protected.HandleFunc("/tasks/export", taskH.ExportCSV).Methods(http.MethodGet)
	protected.HandleFunc("/tasks/sync", taskH.Sync).Methods(http.MethodGet)
	protected.HandleFunc("/tasks/bulk", taskH.BulkUpdate).Methods(http.MethodPatch)
	protected.HandleFunc("/tasks/bulk", taskH.BulkDelete).Methods(http.MethodDelete)
	protected.HandleFunc("/tasks/{id}", taskH.Get).Methods(http.MethodGet)
	protected.HandleFunc("/tasks/{id}", taskH.Update).Methods(http.MethodPatch)
	protected.HandleFunc("/tasks/{id}", taskH.Delete).Methods(http.MethodDelete)
	protected.HandleFunc("/tasks/{id}/status", taskH.UpdateStatus).Methods(http.MethodPatch)
	protected.HandleFunc("/tasks/{id}/attachments", attachH.Upload).Methods(http.MethodPost)

	protected.HandleFunc("/tags", tagH.Create).Methods(http.MethodPost)
	protected.HandleFunc("/tags", tagH.List).Methods(http.MethodGet)
	protected.HandleFunc("/tags/{id}", tagH.Delete).Methods(http.MethodDelete)

	protected.HandleFunc("/attachments/{id}", attachH.Download).Methods(http.MethodGet)
	protected.HandleFunc("/attachments/{id}", attachH.Delete).Methods(http.MethodDelete)

	r.HandleFunc("/ws", authMW(http.HandlerFunc(hub.WSHandler)).ServeHTTP)

	// Swagger UI
	if os.Getenv("ENV") != "production" {
		r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	}
}
