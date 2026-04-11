package handler

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/backend/internal/middleware"
	"taskflow/backend/internal/model"
)

// TaskHandler handles task-related endpoints.
type TaskHandler struct {
	db *pgxpool.Pool
}

// NewTaskHandler creates a new TaskHandler.
func NewTaskHandler(db *pgxpool.Pool) *TaskHandler {
	return &TaskHandler{db: db}
}

// List returns tasks for a given project with optional status and assignee filters.
func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	// Check project exists
	var exists bool
	err = h.db.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&exists)
	if err != nil {
		slog.Error("failed to check project existence", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		Error(w, http.StatusNotFound, "not found")
		return
	}

	page, limit := parsePagination(r)
	offset := (page - 1) * limit

	// Build query with optional filters
	query := `SELECT id, title, description, status, priority, project_id, creator_id, assignee_id, due_date, created_at, updated_at
		 FROM tasks WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM tasks WHERE project_id = $1`
	args := []interface{}{projectID}
	countArgs := []interface{}{projectID}
	paramIdx := 2

	if status := r.URL.Query().Get("status"); status != "" {
		if !isValidStatus(status) {
			ValidationError(w, map[string]string{"status": "must be one of: todo, in_progress, done"})
			return
		}
		query += ` AND status = $` + itoa(paramIdx)
		countQuery += ` AND status = $` + itoa(paramIdx)
		args = append(args, status)
		countArgs = append(countArgs, status)
		paramIdx++
	}

	if assignee := r.URL.Query().Get("assignee"); assignee != "" {
		assigneeID, err := uuid.Parse(assignee)
		if err != nil {
			ValidationError(w, map[string]string{"assignee": "must be a valid UUID"})
			return
		}
		query += ` AND assignee_id = $` + itoa(paramIdx)
		countQuery += ` AND assignee_id = $` + itoa(paramIdx)
		args = append(args, assigneeID)
		countArgs = append(countArgs, assigneeID)
		paramIdx++
	}

	// Get total count
	var totalCount int
	err = h.db.QueryRow(r.Context(), countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		slog.Error("failed to count tasks", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	query += ` ORDER BY created_at DESC LIMIT $` + itoa(paramIdx) + ` OFFSET $` + itoa(paramIdx+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(r.Context(), query, args...)
	if err != nil {
		slog.Error("failed to query tasks", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	tasks := []model.Task{}
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.CreatorID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			slog.Error("failed to scan task", "error", err)
			Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		tasks = append(tasks, t)
	}

	JSON(w, http.StatusOK, model.PaginatedResponse{
		Data:       tasks,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(limit))),
	})
}

// Create creates a new task within a project.
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	// Check project exists
	var exists bool
	err = h.db.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1)", projectID).Scan(&exists)
	if err != nil {
		slog.Error("failed to check project existence", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !exists {
		Error(w, http.StatusNotFound, "not found")
		return
	}

	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	fields := make(map[string]string)
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		fields["title"] = "is required"
	}
	if req.Status == "" {
		req.Status = model.TaskStatusTodo
	} else if !isValidStatus(string(req.Status)) {
		fields["status"] = "must be one of: todo, in_progress, done"
	}
	if req.Priority == "" {
		req.Priority = model.TaskPriorityMedium
	} else if !isValidPriority(string(req.Priority)) {
		fields["priority"] = "must be one of: low, medium, high"
	}

	var dueDate *time.Time
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.DueDate)
		if err != nil {
			fields["due_date"] = "must be in YYYY-MM-DD format"
		} else {
			dueDate = &parsed
		}
	}

	if len(fields) > 0 {
		ValidationError(w, fields)
		return
	}

	var task model.Task
	err = h.db.QueryRow(
		r.Context(),
		`INSERT INTO tasks (title, description, status, priority, project_id, creator_id, assignee_id, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, title, description, status, priority, project_id, creator_id, assignee_id, due_date, created_at, updated_at`,
		req.Title, req.Description, req.Status, req.Priority, projectID, userID, req.AssigneeID, dueDate,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.ProjectID, &task.CreatorID, &task.AssigneeID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		slog.Error("failed to create task", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("task created", "task_id", task.ID, "project_id", projectID)
	JSON(w, http.StatusCreated, task)
}

// Update updates a task's fields.
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	// Check task exists
	var existingTask model.Task
	err = h.db.QueryRow(r.Context(),
		`SELECT id, title, description, status, priority, project_id, creator_id, assignee_id, due_date, created_at, updated_at
		 FROM tasks WHERE id = $1`, taskID,
	).Scan(&existingTask.ID, &existingTask.Title, &existingTask.Description, &existingTask.Status,
		&existingTask.Priority, &existingTask.ProjectID, &existingTask.CreatorID, &existingTask.AssigneeID,
		&existingTask.DueDate, &existingTask.CreatedAt, &existingTask.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			Error(w, http.StatusNotFound, "not found")
			return
		}
		slog.Error("failed to get task", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	fields := make(map[string]string)
	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			fields["title"] = "cannot be empty"
		}
		req.Title = &trimmed
	}
	if req.Status != nil && !isValidStatus(string(*req.Status)) {
		fields["status"] = "must be one of: todo, in_progress, done"
	}
	if req.Priority != nil && !isValidPriority(string(*req.Priority)) {
		fields["priority"] = "must be one of: low, medium, high"
	}

	var dueDate *time.Time
	dueDateProvided := false
	if req.DueDate != nil {
		dueDateProvided = true
		if *req.DueDate != "" {
			parsed, err := time.Parse("2006-01-02", *req.DueDate)
			if err != nil {
				fields["due_date"] = "must be in YYYY-MM-DD format"
			} else {
				dueDate = &parsed
			}
		}
		// if empty string, dueDate stays nil (clears the field)
	}

	if len(fields) > 0 {
		ValidationError(w, fields)
		return
	}

	// Build update with COALESCE for optional fields
	title := existingTask.Title
	if req.Title != nil {
		title = *req.Title
	}
	description := existingTask.Description
	if req.Description != nil {
		description = req.Description
	}
	status := existingTask.Status
	if req.Status != nil {
		status = *req.Status
	}
	priority := existingTask.Priority
	if req.Priority != nil {
		priority = *req.Priority
	}
	assigneeID := existingTask.AssigneeID
	if req.AssigneeID != nil {
		assigneeID = req.AssigneeID
	}
	if !dueDateProvided {
		dueDate = existingTask.DueDate
	}

	var task model.Task
	err = h.db.QueryRow(
		r.Context(),
		`UPDATE tasks SET title = $1, description = $2, status = $3, priority = $4,
		 assignee_id = $5, due_date = $6, updated_at = NOW()
		 WHERE id = $7
		 RETURNING id, title, description, status, priority, project_id, creator_id, assignee_id, due_date, created_at, updated_at`,
		title, description, status, priority, assigneeID, dueDate, taskID,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.ProjectID, &task.CreatorID, &task.AssigneeID, &task.DueDate, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		slog.Error("failed to update task", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("task updated", "task_id", task.ID)
	JSON(w, http.StatusOK, task)
}

// Delete deletes a task. Only the project owner or task creator can delete.
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	taskID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task ID")
		return
	}

	// Get task with project owner info
	var creatorID, projectOwnerID uuid.UUID
	err = h.db.QueryRow(r.Context(),
		`SELECT t.creator_id, p.owner_id
		 FROM tasks t
		 JOIN projects p ON p.id = t.project_id
		 WHERE t.id = $1`, taskID,
	).Scan(&creatorID, &projectOwnerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			Error(w, http.StatusNotFound, "not found")
			return
		}
		slog.Error("failed to get task for deletion", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if userID != creatorID && userID != projectOwnerID {
		Error(w, http.StatusForbidden, "only the task creator or project owner can delete this task")
		return
	}

	if _, err := h.db.Exec(r.Context(), "DELETE FROM tasks WHERE id = $1", taskID); err != nil {
		slog.Error("failed to delete task", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("task deleted", "task_id", taskID)
	w.WriteHeader(http.StatusNoContent)
}

func isValidStatus(s string) bool {
	return s == "todo" || s == "in_progress" || s == "done"
}

func isValidPriority(p string) bool {
	return p == "low" || p == "medium" || p == "high"
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
