package auth

import (
	"encoding/json"
	"net/http"

	"github.com/Tonyblaise/to-do/internal/config"
	"github.com/Tonyblaise/to-do/internal/models"
	"github.com/Tonyblaise/to-do/internal/repository"
	"github.com/Tonyblaise/to-do/internal/response"
)

type AuthHandler struct{
	user *repository.UserRepository
	cfg *config.Config
}


func NewAuthHandler(users *repository.UserRepository, cfg *config.Config) *AuthHandler{
	return &AuthHandler{
		user: users, cfg: cfg,
	}
}

func (h *AuthHandler)SignUp(w http.ResponseWriter, r *http.Request){
	var req models.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err!=nil{
		response.BadRequest(w, "invalid JSON body")
		return
	}

}

