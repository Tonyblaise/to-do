package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Tonyblaise/to-do/internal/middleware"
	"github.com/Tonyblaise/to-do/internal/models"
	"github.com/Tonyblaise/to-do/internal/repository"
	"github.com/Tonyblaise/to-do/internal/response"
	"github.com/Tonyblaise/to-do/internal/services"
	"github.com/gorilla/mux"
)

type TaskHandler struct {
	tasks *repository.TaskRepository
	hub   *Hub
}

func NewTaskHandler(tasks *repository.TaskRepository, hub *Hub) *TaskHandler {
	return &TaskHandler{tasks: tasks , hub: hub}
}

// Create godoc
// @Summary      Create a task
// @Description  Create a new task for the authenticated user
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      models.CreateTaskRequest  true  "Task payload"
// @Success      201   {object}  models.Task
// @Failure      400   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /tasks [post]
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	var req models.CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}
	if errs := services.ValidateCreateTask(&req); errs != nil {
		response.BadRequest(w, "validation failed", errs.Fields)
		return
	}
	task, err := h.tasks.Create(userID, &req)
	if err != nil {
		response.InternalError(w)
		return
	}

	h.hub.BroadcastToUser(userID, models.WSMessage{Type: "task.created", Payload: task})
	response.Created(w, task)

}

// List godoc
// @Summary      List tasks
// @Description  Get paginated tasks for the authenticated user with optional filters
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        search    query     string  false  "Search term"
// @Param        status    query     string  false  "Filter by status (pending|in_progress|completed|archived)"
// @Param        priority  query     string  false  "Filter by priority (low|medium|high)"
// @Param        cursor    query     string  false  "Pagination cursor"
// @Param        limit     query     int     false  "Page size"
// @Success      200       {object}  models.TaskListResponse
// @Failure      500       {object}  models.ErrorResponse
// @Router       /tasks [get]
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	q := r.URL.Query()

	filter := repository.TaskFilter{
		Search: q.Get("search"),
		Cursor: q.Get("cursor"),
	}

	if s := q.Get("status"); s != "" {
		status := models.TaskStatus(s)
		filter.Status = &status
	}
	if p := q.Get("priority"); p != "" {
		prio := models.Priority(p)
		filter.Priority = &prio
	}

	var limit int
	fmt.Sscanf(q.Get("limit"), "%d", &limit)
	filter.Limit = limit

	tasks, nextCursor, total, err := h.tasks.List(userID, filter)
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, models.TaskListResponse{
		Tasks:      tasks,
		NextCursor: nextCursor,
		Total:      total,
	})
}
// Get godoc
// @Summary      Get a task
// @Description  Retrieve a single task by ID
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Task ID"
// @Success      200  {object}  models.Task
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /tasks/{id} [get]
func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	taskID := mux.Vars(r)["id"]

	task, err := h.tasks.GetByID(taskID, userID)
	if err == repository.ErrUserNotFound {
		response.NotFound(w, "task not found")
		return
	}
	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, task)
}
// Update godoc
// @Summary      Update a task
// @Description  Update task fields (title, description, priority, due_date, tags)
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                   true  "Task ID"
// @Param        body  body      models.UpdateTaskRequest true  "Update payload"
// @Success      200   {object}  models.Task
// @Failure      400   {object}  models.ErrorResponse
// @Failure      404   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /tasks/{id} [patch]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	taskID := mux.Vars(r)["id"]

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	task, err := h.tasks.Update(taskID, userID, &req)

	if err == repository.ErrUserNotFound {
		response.NotFound(w, "task not found")
		return
	}
	if err != nil {
		response.InternalError(w)
		return
	}

	h.hub.BroadcastToUser(userID, models.WSMessage{Type: "task.updated", Payload: task})
	response.OK(w, task)
}

// UpdateStatus godoc
// @Summary      Update task status
// @Description  Change the status of a task
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                     true  "Task ID"
// @Param        body  body      models.UpdateStatusRequest true  "Status payload"
// @Success      200   {object}  models.Task
// @Failure      400   {object}  models.ErrorResponse
// @Failure      404   {object}  models.ErrorResponse
// @Failure      500   {object}  models.ErrorResponse
// @Router       /tasks/{id}/status [patch]
func (h *TaskHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	taskID := mux.Vars(r)["id"]

	var req models.UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	if !services.ValidateStatus(req.Status) {
		response.BadRequest(w, "invalid status value")
		return
	}

	task, err := h.tasks.UpdateStatus(taskID, userID, req.Status)
	if err == repository.ErrUserNotFound {
		response.NotFound(w, "task not found")
		return
	}
	if err != nil {
		response.InternalError(w)
		return
	}

	h.hub.BroadcastToUser(userID, models.WSMessage{Type: "task.status_changed", Payload: task})
	response.OK(w, task)
}
// Delete godoc
// @Summary      Delete a task
// @Description  Soft-delete a task by ID
// @Tags         tasks
// @Security     BearerAuth
// @Param        id   path  string  true  "Task ID"
// @Success      204
// @Failure      404  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	taskID := mux.Vars(r)["id"]

	if err := h.tasks.SoftDelete(taskID, userID); err == repository.ErrUserNotFound {
		response.NotFound(w, "task not found")
		return
	} else if err != nil {
		response.InternalError(w)
		return
	}

	h.hub.BroadcastToUser(userID, models.WSMessage{Type: "task.deleted", Payload: map[string]string{"id": taskID}})
	response.NoContent(w)
}

// BulkUpdate godoc
// @Summary      Bulk update tasks
// @Description  Update multiple tasks at once
// @Tags         tasks
// @Accept       json
// @Security     BearerAuth
// @Param        body  body  models.BulkUpdateRequest  true  "Bulk update payload"
// @Success      204
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /tasks/bulk [patch]
func (h *TaskHandler) BulkUpdate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req models.BulkUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}
	if len(req.TaskIDs) == 0 {
		response.BadRequest(w, "task_ids is required")
		return
	}

	if err := h.tasks.BulkUpdate(userID, req.TaskIDs, &req.Update); err != nil {
		response.InternalError(w)
		return
	}
	response.NoContent(w)
}

// BulkDelete godoc
// @Summary      Bulk delete tasks
// @Description  Soft-delete multiple tasks at once
// @Tags         tasks
// @Accept       json
// @Security     BearerAuth
// @Param        body  body  models.BulkDeleteRequest  true  "Bulk delete payload"
// @Success      204
// @Failure      400  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /tasks/bulk [delete]
func (h *TaskHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req models.BulkDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}
	if len(req.TaskIDs) == 0 {
		response.BadRequest(w, "task_ids is required")
		return
	}

	if err := h.tasks.BulkDelete(userID, req.TaskIDs); err != nil {
		response.InternalError(w)
		return
	}
	response.NoContent(w)
}

// ExportCSV godoc
// @Summary      Export tasks as CSV
// @Description  Download all tasks as a CSV file
// @Tags         tasks
// @Produce      text/csv
// @Security     BearerAuth
// @Success      200  {file}    binary
// @Failure      500  {object}  models.ErrorResponse
// @Router       /tasks/export [get]
func (h *TaskHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	tasks, _, _, err := h.tasks.List(userID, repository.TaskFilter{Limit: 10000})
	if err != nil {
		response.InternalError(w)
		return
	}

	csv, err := services.ExportTasksCSV(tasks)
	if err != nil {
		response.InternalError(w)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=tasks.csv")
	w.WriteHeader(http.StatusOK)

	buf := make([]byte, 32*1024)
	for {
		n, err := csv.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

// Sync godoc
// @Summary      Sync tasks
// @Description  Get tasks created/modified/deleted since a given timestamp
// @Tags         tasks
// @Produce      json
// @Security     BearerAuth
// @Param        last_synced_at  query     string  true  "RFC3339 timestamp of last sync"
// @Success      200             {object}  models.SyncResponse
// @Failure      400             {object}  models.ErrorResponse
// @Failure      500             {object}  models.ErrorResponse
// @Router       /tasks/sync [get]
func (h *TaskHandler) Sync(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	lastSyncStr := r.URL.Query().Get("last_synced_at")
	

	lastSync, err := time.Parse(time.RFC3339, lastSyncStr)

	if err != nil {
		response.BadRequest(w, "last_synced_at must be a valid RFC3339 timestamp")
		return
	}

	records, err := h.tasks.Sync(userID, lastSync)

	if err != nil {
		response.InternalError(w)
		return
	}

	response.OK(w, models.SyncResponse{
		Records:      records,
		LastSyncedAt: time.Now().UTC(),
	})
}
