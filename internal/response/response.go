package response

import (
	"encoding/json"
	"github.com/Tonyblaise/to-do/internal/models"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)

}

func OK(w http.ResponseWriter, v interface{}) {
	JSON(w, http.StatusCreated, v)
}

func Created(w http.ResponseWriter, v interface{}) {
	JSON(w, http.StatusCreated, v)
}
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
func BadRequest(w http.ResponseWriter, msg string, details ...map[string]string) {
	e := models.ErrorResponse{Error: msg, Code: "BAD_REQUEST"}
	if len(details) > 0 {
		e.Details = details[0]
	}
	JSON(w, http.StatusBadRequest, e)
}
func Unauthorized(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: msg, Code: "UNAUTHORIZED"})
}
func Forbidden(w http.ResponseWriter) {
	JSON(w, http.StatusForbidden, models.ErrorResponse{Error: "forbidden", Code: "FORBIDDEN"})
}
func NotFound(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusNotFound, models.ErrorResponse{Error: msg, Code: "NOT_FOUND"})
}

func Conflict(w http.ResponseWriter, msg string) {
	JSON(w, http.StatusConflict, models.ErrorResponse{Error: msg, Code: "CONFLICT"})
}
func TooManyRequests(w http.ResponseWriter) {
	JSON(w, http.StatusTooManyRequests, models.ErrorResponse{Error: "rate limit exceeded", Code: "RATE_LIMITED"})
}
func InternalError(w http.ResponseWriter) {
	JSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "internal server error", Code: "INTERNAL_ERROR"})
}