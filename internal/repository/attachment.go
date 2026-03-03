package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Tonyblaise/to-do/internal/models"
	"github.com/google/uuid"
)

type AttachmentRepository struct {
	db *sql.DB
}

func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(taskID, userID, filename, mimeType, path string, size int64) (*models.Attachment, error) {
	a := &models.Attachment{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		UserID:    userID,
		Filename:  filename,
		MimeType:  mimeType,
		Size:      size,
		Path:      path,
		CreatedAt: time.Now(),
	}

	err := r.db.QueryRow(`
		INSERT INTO task_attachments (id, task_id, user_id, filename, mime_type, size, path, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, created_at`,
		a.ID, a.TaskID, a.UserID, a.Filename, a.MimeType, a.Size, a.Path, a.CreatedAt,
	).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting attachment: %w", err)
	}

	a.URL = "/api/v1/attachments/" + a.ID
	return a, nil
}

func (r *AttachmentRepository) GetByID(attachmentID, userID string) (*models.Attachment, error) {
	a := &models.Attachment{}
	err := r.db.QueryRow(`
		SELECT id, task_id, user_id, filename, mime_type, size, path, created_at
		FROM task_attachments WHERE id = $1 AND user_id = $2`,
		attachmentID, userID,
	).Scan(&a.ID, &a.TaskID, &a.UserID, &a.Filename, &a.MimeType, &a.Size, &a.Path, &a.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying attachment: %w", err)
	}

	a.URL = "/api/v1/attachments/" + a.ID
	return a, nil
}

func (r *AttachmentRepository) Delete(attachmentID, userID string) (string, error) {
	var path string
	err := r.db.QueryRow(`
		DELETE FROM task_attachments WHERE id = $1 AND user_id = $2 RETURNING path`,
		attachmentID, userID,
	).Scan(&path)

	if err == sql.ErrNoRows {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", fmt.Errorf("deleting attachment: %w", err)
	}

	return path, nil
}