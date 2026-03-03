package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Tonyblaise/to-do/internal/models"
	"github.com/google/uuid"

	"github.com/lib/pq"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

type TaskFilter struct {
	Status   *models.TaskStatus
	Priority *models.Priority
	TagIDs   []string
	Search   string
	Cursor   string
	Limit    int
	ParentID string
}

func (r *TaskRepository) Create(userID string, req *models.CreateTaskRequest) (*models.Task, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	priority := req.Priority
	if priority == "" {
		priority = models.PriorityMedium
	}

	task := &models.Task{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      models.TaskStatusPending,

		Priority:  priority,
		DueDate:   req.DueDate,
		ParentID:  req.ParentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = tx.QueryRow(`
		INSERT INTO tasks (id, user_id, title, description, status, priority, due_date, parent_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at`,
		task.ID, task.UserID, task.Title, task.Description,
		task.Status, task.Priority, task.DueDate, task.ParentID,
		task.CreatedAt, task.UpdatedAt,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("inserting task: %w", err)
	}

	if len(req.TagIDs) > 0 {
		if err := assignTags(tx, task.ID, userID, req.TagIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("comming transaction: %w", err)
	}

	return r.GetByID(task.ID, userID)

}

func (r *TaskRepository) GetByID(taskID string, userID string) (*models.Task, error) {
	task := &models.Task{}
	err := r.db.QueryRow(`
		SELECT id, user_id, title, description, status, priority,
		       due_date, parent_id, deleted_at, created_at, updated_at
		FROM tasks
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		taskID, userID,
	).Scan(
		&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.Priority, &task.DueDate, &task.ParentID,
		&task.DeletedAt, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying task: %w", err)
	}

	// Load tags
	tags, err := r.getTagsForTask(task.ID)
	if err != nil {
		return nil, err
	}
	task.Tags = tags

	// Load subtasks
	subtasks, err := r.getSubtasks(task.ID, userID)
	if err != nil {
		return nil, err
	}
	task.Subtasks = subtasks

	// Load attachments
	attachments, err := r.getAttachments(task.ID)
	if err != nil {
		return nil, err
	}
	task.Attachments = attachments

	return task, nil

}

func (r *TaskRepository) List(userID string, f TaskFilter) ([]models.Task, *string, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 20
	}

	args := []interface{}{userID}
	conds := []string{"t.user_id = $1", "t.deleted_at IS NULL", "t.parent_id IS NULL"}
	i := 2

	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("t.status = $%d", i))
		args = append(args, *f.Status)
		i++
	}
	if f.Priority != nil {
		conds = append(conds, fmt.Sprintf("t.priority = $%d", i))
		args = append(args, *f.Priority)
		i++
	}
	if f.Search != "" {
		conds = append(conds, fmt.Sprintf("(t.title ILIKE $%d OR t.description ILIKE $%d)", i, i))
		args = append(args, "%"+f.Search+"%")
		i++
	}
	if f.Cursor != "" {
		conds = append(conds, fmt.Sprintf("t.id > $%d", i))
		args = append(args, f.Cursor)
		i++
	}
	if len(f.TagIDs) > 0 {
		conds = append(conds, fmt.Sprintf(`
			EXISTS (SELECT 1 FROM task_tags tt WHERE tt.task_id = t.id AND tt.tag_id = ANY($%d))`, i))
		args = append(args, pq.Array(f.TagIDs))
		i++
	}

	where := strings.Join(conds, " AND ")
	query := fmt.Sprintf(`
		SELECT t.id, t.user_id, t.title, t.description, t.status, t.priority,
		       t.due_date, t.parent_id, t.deleted_at, t.created_at, t.updated_at,
		       COUNT(*) OVER() as total
		FROM tasks t
		WHERE %s
		ORDER BY t.created_at DESC, t.id
		LIMIT $%d`, where, i)
	args = append(args, f.Limit+1)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("listing tasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	var total int
	for rows.Next() {
		var t models.Task
		err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.ParentID,
			&t.DeletedAt, &t.CreatedAt, &t.UpdatedAt, &total,
		)
		if err != nil {
			return nil, nil, 0, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, t)
	}

	var nextCursor *string
	if len(tasks) > f.Limit {
		cursor := tasks[f.Limit-1].ID
		nextCursor = &cursor
		tasks = tasks[:f.Limit]
	}

	if len(tasks) > 0 {
		ids := make([]string, len(tasks))
		for i, t := range tasks {
			ids[i] = t.ID
		}
		tagMap, err := r.getTagsForTasks(ids)
		if err != nil {
			return nil, nil, 0, err
		}
		for i := range tasks {
			tasks[i].Tags = tagMap[tasks[i].ID]
		}
	}

	return tasks, nextCursor, total, nil
}

func (r *TaskRepository) Update(taskID, userID string, req *models.UpdateTaskRequest) (*models.Task, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	i := 1

	if req.Title != nil {
		sets = append(sets, fmt.Sprintf("title = $%d", i))
		args = append(args, *req.Title)
		i++
	}
	if req.Description != nil {
		sets = append(sets, fmt.Sprintf("description = $%d", i))
		args = append(args, *req.Description)
		i++
	}
	if req.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", i))
		args = append(args, *req.Priority)
		i++
	}
	if req.DueDate != nil {
		sets = append(sets, fmt.Sprintf("due_date = $%d", i))
		args = append(args, *req.DueDate)
		i++
	}

	args = append(args, taskID, userID)
	query := fmt.Sprintf(`
		UPDATE tasks SET %s
		WHERE id = $%d AND user_id = $%d AND deleted_at IS NULL`,
		strings.Join(sets, ", "), i, i+1)

	result, err := tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, ErrUserNotFound
	}

	if req.TagIDs != nil {
		if _, err := tx.Exec(`DELETE FROM task_tags WHERE task_id = $1`, taskID); err != nil {
			return nil, fmt.Errorf("clearing tags: %w", err)
		}
		if len(req.TagIDs) > 0 {
			if err := assignTags(tx, taskID, userID, req.TagIDs); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing transaction: %w", err)
	}

	return r.GetByID(taskID, userID)
}

func (r *TaskRepository) UpdateStatus(taskID, userID string, status models.TaskStatus) (*models.Task, error) {
	result, err := r.db.Exec(`
		UPDATE tasks SET status = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL`,
		status, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("updating status: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return nil, ErrUserNotFound
	}
	return r.GetByID(taskID, userID)
}
func (r *TaskRepository) SoftDelete(taskID, userID string) error {
	result, err := r.db.Exec(`
		UPDATE tasks SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`,
		taskID, userID)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *TaskRepository) BulkUpdate(userID string, taskIDs []string, req *models.UpdateTaskRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	sets := []string{"updated_at = NOW()"}
	args := []interface{}{}
	i := 1

	if req.Priority != nil {
		sets = append(sets, fmt.Sprintf("priority = $%d", i))
		args = append(args, *req.Priority)
		i++
	}

	args = append(args, pq.Array(taskIDs), userID)
	query := fmt.Sprintf(`
		UPDATE tasks SET %s
		WHERE id = ANY($%d) AND user_id = $%d AND deleted_at IS NULL`,
		strings.Join(sets, ", "), i, i+1)

	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("bulk updating: %w", err)
	}

	return tx.Commit()
}
func (r *TaskRepository) BulkDelete(userID string, taskIDs []string) error {

	tx, err := r.db.Begin()

	if err != nil {
		return fmt.Errorf("starting transacitons %w", err)
	}

	defer tx.Rollback()

	result, err := r.db.Exec(`UPDATE tasks 
	SET(deleted_at=NOW(), updated_at=NOW())
	WHERE id = ANY($1) AND user_id = $3 AND deleted_at IS NUL`, pq.Array(taskIDs), userID)

	if err != nil {
		return fmt.Errorf("bulk deleting %w", err)
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("no entries found")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil

}
func (r *TaskRepository) Sync(userID string, lastSyncedAt time.Time) ([]models.SyncRecord, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, title, description, status, priority,
		       due_date, parent_id, deleted_at, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND updated_at > $2
		ORDER BY updated_at ASC`,
		userID, lastSyncedAt)
	if err != nil {
		return nil, fmt.Errorf("syncing tasks: %w", err)
	}
	defer rows.Close()

	var records []models.SyncRecord
	for rows.Next() {
		var t models.Task
		err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.ParentID,
			&t.DeletedAt, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning sync record: %w", err)
		}

		action := models.SyncActionModified
		if t.DeletedAt != nil {
			action = models.SyncActionDeleted
		} else if t.CreatedAt.After(lastSyncedAt) {
			action = models.SyncActionCreated
		}

		records = append(records, models.SyncRecord{Task: t, Action: action})
	}

	return records, nil
}

func (r *TaskRepository) GetUpcomingDue(window time.Duration) ([]models.Task, error) {
	now := time.Now()
	until := now.Add(window)

	rows, err := r.db.Query(`
		SELECT t.id, t.user_id, t.title, t.due_date,
		       u.email
		FROM tasks t
		JOIN users u ON u.id = t.user_id
		WHERE t.due_date BETWEEN $1 AND $2
		  AND t.status != 'completed'
		  AND t.deleted_at IS NULL`,
		now, until)
	if err != nil {
		return nil, fmt.Errorf("querying upcoming tasks: %w", err)
	}
	defer rows.Close()
	

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var email string
		if err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.DueDate, &email); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) getTagsForTask(taskID string) ([]models.Tag, error) {
	rows, err := r.db.Query(`
		SELECT tg.id, tg.user_id, tg.name, tg.color, tg.created_at
		FROM tags tg
		JOIN task_tags tt ON tt.tag_id = tg.id
		WHERE tt.task_id = $1`, taskID)
	if err != nil {
		return nil, fmt.Errorf("querying tags: %w", err)
	}
	defer rows.Close()
	var tags []models.Tag
	for rows.Next() {
		var tg models.Tag
		if err := rows.Scan(&tg.ID, &tg.UserID, &tg.Name, &tg.Color, &tg.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tg)
	}
	return tags, nil


}
func (r *TaskRepository) getTagsForTasks(taskIDs []string) (map[string][]models.Tag, error) {
	rows, err := r.db.Query(`
		SELECT tt.task_id, tg.id, tg.user_id, tg.name, tg.color, tg.created_at
		FROM tags tg
		JOIN task_tags tt ON tt.tag_id = tg.id
		WHERE tt.task_id = ANY($1)`, pq.Array(taskIDs))

	if err != nil {
		return nil, fmt.Errorf("batch querying tags: %w", err)
	}
	defer rows.Close()

	result := make(map[string][]models.Tag)

	for rows.Next() {

		var taskID string
		var tg models.Tag
		if err := rows.Scan(&taskID, &tg.ID, &tg.UserID, &tg.Name, &tg.Color, &tg.CreatedAt); err != nil {
			return nil, err
		}
		result[taskID] = append(result[taskID], tg)
	}
	return result, nil
}
func (r *TaskRepository) getSubtasks(parentID, userID string) ([]models.Task, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, title, description, status, priority,
		       due_date, parent_id, deleted_at, created_at, updated_at
		FROM tasks
		WHERE parent_id = $1 AND user_id = $2 AND deleted_at IS NULL
		ORDER BY created_at`, parentID, userID)
	if err != nil {
		return nil, fmt.Errorf("querying subtasks: %w", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.Title, &t.Description,
			&t.Status, &t.Priority, &t.DueDate, &t.ParentID,
			&t.DeletedAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *TaskRepository) getAttachments(taskID string) ([]models.Attachment, error) {
	rows, err := r.db.Query(`
	SELECT id, task_id, user_id, filename, mimetype, size, path, created_at
	FROM task_attachments WHERE task_id=$1
	`, taskID)

	if err != nil {
		return nil, fmt.Errorf("querying attachments: %w", err)
	}
	defer rows.Close()
	var attachments []models.Attachment
	for rows.Next() {
		var a models.Attachment
		if err := rows.Scan(&a.ID, &a.TaskID, &a.UserID, &a.Filename,
			&a.MimeType, &a.Size, &a.Path, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.URL = "/api/v1/attachments/" + a.ID
		attachments = append(attachments, a)
	}
	return attachments, nil

}

func assignTags(tx *sql.Tx, taskID, UserID string, tagIDs []string) error {
	for _, tagID := range tagIDs {
		var exists bool
		err := tx.QueryRow(`SELECT EXISTS (SELECT 1 FROM tags WHERE id=$1 and user_id=$2)`, tagID, UserID).Scan(&exists)
		if err != nil || !exists {
			continue
		}

		_, err = tx.Exec(`INSERT INTO task_tags (task_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, taskID, tagID)
		if err != nil {
			return fmt.Errorf("assigning tag %s: %w", tagID, err)
		}

	}
	return nil
}
