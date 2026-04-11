package handler

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/backend/internal/middleware"
	"taskflow/backend/internal/model"
)

// ProjectHandler handles project-related endpoints.
type ProjectHandler struct {
	db *pgxpool.Pool
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(db *pgxpool.Pool) *ProjectHandler {
	return &ProjectHandler{db: db}
}

// List returns all projects the current user owns or has tasks in.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	page, limit := parsePagination(r)
	offset := (page - 1) * limit

	// Count total
	var totalCount int
	err := h.db.QueryRow(r.Context(),
		`SELECT COUNT(DISTINCT p.id) FROM projects p
		 LEFT JOIN tasks t ON t.project_id = p.id
		 WHERE p.owner_id = $1 OR t.assignee_id = $1 OR t.creator_id = $1`,
		userID,
	).Scan(&totalCount)
	if err != nil {
		slog.Error("failed to count projects", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	rows, err := h.db.Query(r.Context(),
		`SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
		 FROM projects p
		 LEFT JOIN tasks t ON t.project_id = p.id
		 WHERE p.owner_id = $1 OR t.assignee_id = $1 OR t.creator_id = $1
		 ORDER BY p.created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		slog.Error("failed to query projects", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	projects := []model.Project{}
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			slog.Error("failed to scan project", "error", err)
			Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		projects = append(projects, p)
	}

	JSON(w, http.StatusOK, model.PaginatedResponse{
		Data:       projects,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(limit))),
	})
}

// Create creates a new project owned by the current user.
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		ValidationError(w, map[string]string{"name": "is required"})
		return
	}

	var project model.Project
	err := h.db.QueryRow(
		r.Context(),
		`INSERT INTO projects (name, description, owner_id) VALUES ($1, $2, $3)
		 RETURNING id, name, description, owner_id, created_at`,
		req.Name, req.Description, userID,
	).Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		slog.Error("failed to create project", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("project created", "project_id", project.ID, "owner_id", userID)
	JSON(w, http.StatusCreated, project)
}

// Get returns a project with its tasks.
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	var project model.Project
	err = h.db.QueryRow(
		r.Context(),
		"SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1",
		projectID,
	).Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			Error(w, http.StatusNotFound, "not found")
			return
		}
		slog.Error("failed to get project", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Fetch tasks for this project
	rows, err := h.db.Query(
		r.Context(),
		`SELECT id, title, description, status, priority, project_id, creator_id, assignee_id, due_date, created_at, updated_at
		 FROM tasks WHERE project_id = $1 ORDER BY created_at DESC`,
		projectID,
	)
	if err != nil {
		slog.Error("failed to query tasks for project", "error", err)
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

	JSON(w, http.StatusOK, model.ProjectWithTasks{
		Project: project,
		Tasks:   tasks,
	})
}

// Update updates project name and/or description. Only the owner can update.
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	// Check ownership
	var ownerID uuid.UUID
	err = h.db.QueryRow(r.Context(), "SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			Error(w, http.StatusNotFound, "not found")
			return
		}
		slog.Error("failed to check project ownership", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if ownerID != userID {
		Error(w, http.StatusForbidden, "only the project owner can update this project")
		return
	}

	var req model.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			ValidationError(w, map[string]string{"name": "cannot be empty"})
			return
		}
		req.Name = &trimmed
	}

	var project model.Project
	err = h.db.QueryRow(
		r.Context(),
		`UPDATE projects
		 SET name = COALESCE($1, name), description = COALESCE($2, description)
		 WHERE id = $3
		 RETURNING id, name, description, owner_id, created_at`,
		req.Name, req.Description, projectID,
	).Scan(&project.ID, &project.Name, &project.Description, &project.OwnerID, &project.CreatedAt)
	if err != nil {
		slog.Error("failed to update project", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("project updated", "project_id", project.ID)
	JSON(w, http.StatusOK, project)
}

// Delete deletes a project and all its tasks. Only the owner can delete.
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid project ID")
		return
	}

	// Check ownership
	var ownerID uuid.UUID
	err = h.db.QueryRow(r.Context(), "SELECT owner_id FROM projects WHERE id = $1", projectID).Scan(&ownerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			Error(w, http.StatusNotFound, "not found")
			return
		}
		slog.Error("failed to check project ownership", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if ownerID != userID {
		Error(w, http.StatusForbidden, "only the project owner can delete this project")
		return
	}

	// Delete tasks first, then project
	tx, err := h.db.Begin(r.Context())
	if err != nil {
		slog.Error("failed to begin transaction", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer tx.Rollback(r.Context())

	if _, err := tx.Exec(r.Context(), "DELETE FROM tasks WHERE project_id = $1", projectID); err != nil {
		slog.Error("failed to delete tasks", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if _, err := tx.Exec(r.Context(), "DELETE FROM projects WHERE id = $1", projectID); err != nil {
		slog.Error("failed to delete project", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		slog.Error("failed to commit transaction", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("project deleted", "project_id", projectID)
	w.WriteHeader(http.StatusNoContent)
}

// Stats returns task counts grouped by status and by assignee for a project.
func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
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

	// Status counts
	statusCounts := map[string]int{"todo": 0, "in_progress": 0, "done": 0}
	rows, err := h.db.Query(r.Context(),
		"SELECT status, COUNT(*) FROM tasks WHERE project_id = $1 GROUP BY status", projectID)
	if err != nil {
		slog.Error("failed to query status counts", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			slog.Error("failed to scan status count", "error", err)
			Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		statusCounts[status] = count
	}

	// Assignee counts
	assigneeCounts := map[string]map[string]int{}
	rows2, err := h.db.Query(r.Context(),
		`SELECT COALESCE(u.name, 'unassigned'), t.status, COUNT(*)
		 FROM tasks t
		 LEFT JOIN users u ON u.id = t.assignee_id
		 WHERE t.project_id = $1
		 GROUP BY u.name, t.status
		 ORDER BY u.name`, projectID)
	if err != nil {
		slog.Error("failed to query assignee counts", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		var assignee, status string
		var count int
		if err := rows2.Scan(&assignee, &status, &count); err != nil {
			slog.Error("failed to scan assignee count", "error", err)
			Error(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if _, ok := assigneeCounts[assignee]; !ok {
			assigneeCounts[assignee] = map[string]int{}
		}
		assigneeCounts[assignee][status] = count
	}

	JSON(w, http.StatusOK, model.ProjectStats{
		StatusCounts:   statusCounts,
		AssigneeCounts: assigneeCounts,
	})
}

// parsePagination extracts page and limit from query parameters with defaults.
func parsePagination(r *http.Request) (int, int) {
	page := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	return page, limit
}
