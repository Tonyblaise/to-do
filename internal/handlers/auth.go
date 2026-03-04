package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Tonyblaise/to-do/internal/auth"
	"github.com/Tonyblaise/to-do/internal/config"
	"github.com/Tonyblaise/to-do/internal/models"
	"github.com/Tonyblaise/to-do/internal/repository"
	"github.com/Tonyblaise/to-do/internal/response"
	"github.com/Tonyblaise/to-do/internal/services"
)

type AuthHandler struct {
	users *repository.UserRepository
	cfg   *config.Config
}

func NewAuthHandler(users *repository.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg}
}


func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	if errs := services.ValidateSignUp(&req); errs != nil {
		response.BadRequest(w, "validation failed", errs.Fields)
		return
	}

	user, err := h.users.Create(req.Email, req.Password, req.FirstName, req.LastName)
	if err == repository.ErrEmailTaken {
		response.Conflict(w, "email already registered")
		return
	}
	if err != nil {
		response.InternalError(w)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.cfg.JWTSecret, h.cfg.JWTExpiry)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.Created(w, models.AuthResponse{Token: token, User: *user})
}


func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	user, err := h.users.GetByEmail(req.Email)
	if err != nil {
		response.Unauthorized(w, "invalid credentials")
		return
	}

	if !h.users.VerifyPassword(user, req.Password) {
		response.Unauthorized(w, "invalid credentials")
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Email, h.cfg.JWTSecret, h.cfg.JWTExpiry)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, models.AuthResponse{Token: token, User: *user})
}