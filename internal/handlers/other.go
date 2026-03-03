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

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r.Context())

	tags, err := h.tags.List(userID)

	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, tags)
}

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