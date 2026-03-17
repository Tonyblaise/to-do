package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"


	"github.com/Tonyblaise/to-do/internal/config"
	"github.com/Tonyblaise/to-do/internal/middleware"
	"github.com/Tonyblaise/to-do/internal/models"
	"github.com/Tonyblaise/to-do/internal/repository"
	"github.com/Tonyblaise/to-do/internal/response"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type TagHandler struct {
	tags *repository.TagRepository
}

func NewTagHandler(tags *repository.TagRepository) *TagHandler {
	return &TagHandler{tags: tags}
}
// Create godoc
// @Summary      Create a tag
// @Description  Create a new tag for the authenticated user
// @Tags         tags
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.CreateTagRequest  true  "Tag payload"
// @Success      201   {object}  models.Tag
// @Failure      400   {object}  models.ErrorResponse
// @Failure      409   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /tags [post]
func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		response.BadRequest(w, "name is required")
		return
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}

	tag, err := h.tags.Create(userID, &req)

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			response.Conflict(w, "tag name already exists")
			return
		}
		response.InternalError(w)
		return
	}

	response.Created(w, tag)
}

// List godoc
// @Summary      List tags
// @Description  Get all tags for the authenticated user
// @Tags         tags
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   models.Tag
// @Failure      500  {object}  models.ErrorResponse
// @Router       /tags [get]
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r.Context())

	tags, err := h.tags.List(userID)

	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, tags)
}

// Delete godoc
// @Summary      Delete a tag
// @Description  Delete a tag by ID
// @Tags         tags
// @Security     BearerAuth
// @Param        id   path  string  true  "Tag ID"
// @Success      204
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /tags/{id} [delete]
func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	tagID := mux.Vars(r)["id"]

	if err := h.tags.Delete(tagID, userID); err == repository.ErrUserNotFound {
		response.NotFound(w, "tag not found")
		return
	} else if err != nil {
		response.InternalError(w)
		return
	}

	response.NoContent(w)
}

type AttachmentHandler struct {
	attachments *repository.AttachmentRepository
	tasks       *repository.TaskRepository
	cfg         *config.Config
}

func NewAttachmentHandler(attachments *repository.AttachmentRepository, tasks *repository.TaskRepository, cfg *config.Config) *AttachmentHandler {
	return &AttachmentHandler{attachments: attachments, tasks: tasks, cfg: cfg}
}
// Upload godoc
// @Summary      Upload attachment
// @Description  Upload a file attachment to a task (jpeg, png, gif, pdf; max size from config)
// @Tags         attachments
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Task ID"
// @Param        file  formData  file    true  "File to upload"
// @Success      201   {object}  models.Attachment
// @Failure      400   {object}  models.ErrorResponse
// @Failure      404   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /tasks/{id}/attachments [post]
func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	taskID := mux.Vars(r)["id"]
	


	if _, err := h.tasks.GetByID(taskID, userID); err != nil {
		response.NotFound(w, "task not found")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.MaxUploadSize)

	if err := r.ParseMultipartForm(h.cfg.MaxUploadSize); err != nil {
		response.BadRequest(w, fmt.Sprintf("file too large (max %d bytes)", h.cfg.MaxUploadSize))
		return
	}

	file, header, err := r.FormFile("file")

	if err != nil {
		response.BadRequest(w, "file field is required")
		return
	}

	defer file.Close()

	
	mimeType := header.Header.Get("Content-Type")
	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"application/pdf": true,
	}
	if !allowed[mimeType] {
		response.BadRequest(w, "unsupported file type")
		return
	}

	
	storagePath := filepath.Join(h.cfg.StoragePath, userID)
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		response.InternalError(w)
		return
	}

	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(storagePath, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		response.InternalError(w)
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		response.InternalError(w)
		return
	}

	attachment, err := h.attachments.Create(taskID, userID, header.Filename, mimeType, filePath, size)
	if err != nil {
		os.Remove(filePath)
		response.InternalError(w)
		return
	}

	response.Created(w, attachment)
}

// Download godoc
// @Summary      Download attachment
// @Description  Download a file attachment by ID
// @Tags         attachments
// @Produce      octet-stream
// @Security     BearerAuth
// @Param        id   path  string  true  "Attachment ID"
// @Success      200  {file}    binary
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /attachments/{id} [get]
func (h *AttachmentHandler) Download(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	attachmentID := mux.Vars(r)["id"]

	attachment, err := h.attachments.GetByID(attachmentID, userID)
	if err == repository.ErrUserNotFound {
		response.NotFound(w, "attachment not found")
		return
	}
	if err != nil {
		response.InternalError(w)
		return
	}

	w.Header().Set("Content-Type", attachment.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, attachment.Filename))
	http.ServeFile(w, r, attachment.Path)
}
// Delete godoc
// @Summary      Delete attachment
// @Description  Delete a file attachment by ID
// @Tags         attachments
// @Security     BearerAuth
// @Param        id   path  string  true  "Attachment ID"
// @Success      204
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /attachments/{id} [delete]
func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	attachmentID := mux.Vars(r)["id"]

	path, err := h.attachments.Delete(attachmentID, userID)
	if err == repository.ErrUserNotFound {
		response.NotFound(w, "attachment not found")
		return
	}
	if err != nil {
		response.InternalError(w)
		return
	}

	os.Remove(path)
	response.NoContent(w)
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns the health status of the API and database
// @Tags         health
// @Produce      json
// @Success      200  {object}  models.HealthResponse
// @Router       /health [get]
func HealthCheck(db interface{ Ping() error }) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if err := db.Ping(); err != nil {
			dbStatus = "error: " + err.Error()
		}

		status := "ok"
		if dbStatus != "ok" {
			status = "degraded"
		}

		response.OK(w, models.HealthResponse{
			Status:   status,
			Database: dbStatus,
			Version:  "1.0.0",
		})
	}
}