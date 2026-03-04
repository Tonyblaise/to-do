package models

import (
	"time"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusArchived   TaskStatus = "archived"
)

type SyncAction string

const (
	SyncActionCreated  SyncAction = "created"
	SyncActionModified SyncAction = "modified"
	SyncActionDeleted  SyncAction = "deleted"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}



type Task struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      TaskStatus   `json:"status"`
	Priority    Priority     `json:"priority"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	ParentID    *string      `json:"parent_id,omitempty"`
	Tags        []Tag        `json:"tags,omitempty"`
	Subtasks    []Task       `json:"subtasks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}



type Tag struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}


type Attachment struct {
	ID        string    `json:"id"`
	TaskID    string    `json:"task_id"`
	UserID    string    `json:"user_id"`
	Filename  string    `json:"filename"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	Path      string    `json:"-"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type SignupRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    Priority   `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
	ParentID    *string    `json:"parent_id"`
	TagIDs      []string   `json:"tag_ids"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Priority    *Priority  `json:"priority"`
	DueDate     *time.Time `json:"due_date"`
	TagIDs      []string   `json:"tag_ids"`
}

type UpdateStatusRequest struct {
	Status TaskStatus `json:"status"`
}

type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type BulkUpdateRequest struct {
	TaskIDs []string          `json:"task_ids"`
	Update  UpdateTaskRequest `json:"update"`
}

type BulkDeleteRequest struct {
	TaskIDs []string `json:"task_ids"`
}

type TaskListResponse struct {
	Tasks      []Task  `json:"tasks"`
	NextCursor *string `json:"next_cursor,omitempty"`
	Total      int     `json:"total"`
}

type SyncRecord struct {
	Task   Task       `json:"task"`
	Action SyncAction `json:"action"`
}

type SyncResponse struct {
	Records      []SyncRecord `json:"records"`
	LastSyncedAt time.Time    `json:"last_synced_at"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Version  string `json:"version"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}
