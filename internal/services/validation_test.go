package services

import (
	"strings"
	"testing"

	"github.com/Tonyblaise/to-do/internal/models"
)

func TestValidateSignUp(t *testing.T) {
	tests := []struct {
		name    string
		req     models.SignupRequest
		wantErr bool
		field   string
	}{
		{
			name:    "valid signup",
			req:     models.SignupRequest{Email: "jane@example.com", Password: "password123"},
			wantErr: false,
		},
		{
			name:    "missing email",
			req:     models.SignupRequest{Email: "", Password: "password123"},
			wantErr: true,
			field:   "email",
		},
		{
			name:    "whitespace-only email",
			req:     models.SignupRequest{Email: "   ", Password: "password123"},
			wantErr: true,
			field:   "email",
		},
		{
			name:    "email missing @ symbol",
			req:     models.SignupRequest{Email: "notanemail", Password: "password123"},
			wantErr: true,
			field:   "email",
		},
		{
			name:    "password too short",
			req:     models.SignupRequest{Email: "a@b.com", Password: "short"},
			wantErr: true,
			field:   "password",
		},
		{
			name:    "password exactly 8 characters",
			req:     models.SignupRequest{Email: "a@b.com", Password: "12345678"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSignUp(&tt.req)
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got fields: %v", err.Fields)
				return
			}
			if tt.wantErr && tt.field != "" {
				if _, ok := err.Fields[tt.field]; !ok {
					t.Errorf("expected field %q in error fields, got %v", tt.field, err.Fields)
				}
			}
		})
	}
}

func TestValidateCreateTask(t *testing.T) {
	longTitle := strings.Repeat("a", 256)

	tests := []struct {
		name    string
		req     models.CreateTaskRequest
		wantErr bool
		field   string
	}{
		{
			name:    "valid task",
			req:     models.CreateTaskRequest{Title: "My task"},
			wantErr: false,
		},
		{
			name:    "valid task with priority",
			req:     models.CreateTaskRequest{Title: "My task", Priority: models.PriorityHigh},
			wantErr: false,
		},
		{
			name:    "missing title",
			req:     models.CreateTaskRequest{Title: ""},
			wantErr: true,
			field:   "title",
		},
		{
			name:    "whitespace-only title",
			req:     models.CreateTaskRequest{Title: "   "},
			wantErr: true,
			field:   "title",
		},
		{
			name:    "title exceeds 255 characters",
			req:     models.CreateTaskRequest{Title: longTitle},
			wantErr: true,
			field:   "title",
		},
		{
			name:    "invalid priority",
			req:     models.CreateTaskRequest{Title: "Task", Priority: "urgent"},
			wantErr: true,
			field:   "priority",
		},
		{
			name:    "empty priority is allowed",
			req:     models.CreateTaskRequest{Title: "Task", Priority: ""},
			wantErr: false,
		},
		{
			name:    "priority low",
			req:     models.CreateTaskRequest{Title: "Task", Priority: models.PriorityLow},
			wantErr: false,
		},
		{
			name:    "priority medium",
			req:     models.CreateTaskRequest{Title: "Task", Priority: models.PriorityMedium},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateTask(&tt.req)
			if tt.wantErr && err == nil {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got fields: %v", err.Fields)
				return
			}
			if tt.wantErr && tt.field != "" {
				if _, ok := err.Fields[tt.field]; !ok {
					t.Errorf("expected field %q in error fields, got %v", tt.field, err.Fields)
				}
			}
		})
	}
}

func TestValidateStatus(t *testing.T) {
	valid := []models.TaskStatus{
		models.TaskStatusPending,
		models.TaskStatusInProgress,
		models.TaskStatusCompleted,
		models.TaskStatusArchived,
	}
	for _, s := range valid {
		if !ValidateStatus(s) {
			t.Errorf("expected status %q to be valid", s)
		}
	}

	invalid := []models.TaskStatus{"done", "cancelled", "PENDING", ""}
	for _, s := range invalid {
		if ValidateStatus(s) {
			t.Errorf("expected status %q to be invalid", s)
		}
	}
}

func TestValidationError_Error(t *testing.T) {
	e := &ValidationError{Fields: map[string]string{"email": "required"}}
	if e.Error() != "validation error" {
		t.Errorf("Error() = %q, want %q", e.Error(), "validation error")
	}
}
