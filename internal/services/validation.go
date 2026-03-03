package services

import (
	"strings"

	"github.com/Tonyblaise/to-do/internal/models"
)

type ValidationError struct{
	Fields map[string]string
}

func(e *ValidationError) Error()string{
	return "validation error"
}

func ValidateSignUp(req *models.SignupRequest)*ValidationError{
	errs := make(map[string]string)
	if strings.TrimSpace(req.Email)==""{
		errs["email"]="email is required"
	}else if !strings.Contains(req.Email, "@") {
		errs["email"] = "email is invalid"
	}
	if len(req.Password) < 8 {
		errs["password"] = "password must be at least 8 characters"
	}
	if len(errs) > 0 {
		return &ValidationError{Fields: errs}
	}
	return nil


}
func ValidateCreateTask(req *models.CreateTaskRequest) *ValidationError {
	errs := make(map[string]string)
	

	if strings.TrimSpace(req.Title) == "" {
		errs["title"] = "title is required"
	}
	if len(req.Title) > 255 {
		errs["title"] = "title must not exceed 255 characters"
	}

	if req.Priority != "" {
		switch req.Priority {
		case models.PriorityLow, models.PriorityMedium, models.PriorityHigh:
		default:
			errs["priority"] = "priority must be one of: low, medium, high"
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Fields: errs}
	}
	return nil
}
func ValidateStatus(status models.TaskStatus) bool {
	switch status {
	case models.TaskStatusPending, models.TaskStatusInProgress,
		models.TaskStatusCompleted, models.TaskStatusArchived:
		return true
	}
	return false
}