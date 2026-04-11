package model

import (
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

// TaskPriority represents the priority level of a task.
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// User represents an application user.
type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// Project represents a project containing tasks.
type Project struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Task represents a task within a project.
type Task struct {
	ID          uuid.UUID    `json:"id"`
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	ProjectID   uuid.UUID    `json:"project_id"`
	CreatorID   uuid.UUID    `json:"creator_id"`
	AssigneeID  *uuid.UUID   `json:"assignee_id"`
	DueDate     *time.Time   `json:"due_date"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// RegisterRequest represents the payload for user registration.
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the response returned after successful auth.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateProjectRequest represents the payload for creating a project.
type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// UpdateProjectRequest represents the payload for updating a project.
type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// CreateTaskRequest represents the payload for creating a task.
type CreateTaskRequest struct {
	Title       string       `json:"title"`
	Description *string      `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	AssigneeID  *uuid.UUID   `json:"assignee_id"`
	DueDate     *string      `json:"due_date"`
}

// UpdateTaskRequest represents the payload for updating a task.
type UpdateTaskRequest struct {
	Title       *string       `json:"title"`
	Description *string       `json:"description"`
	Status      *TaskStatus   `json:"status"`
	Priority    *TaskPriority `json:"priority"`
	AssigneeID  *uuid.UUID    `json:"assignee_id"`
	DueDate     *string       `json:"due_date"`
}

// ProjectWithTasks holds a project and its tasks for detail views.
type ProjectWithTasks struct {
	Project
	Tasks []Task `json:"tasks"`
}

// ProjectStats holds aggregated statistics for a project.
type ProjectStats struct {
	StatusCounts   map[string]int            `json:"status_counts"`
	AssigneeCounts map[string]map[string]int `json:"assignee_counts"`
}

// PaginationParams holds pagination parameters.
type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// PaginatedResponse wraps a paginated list response.
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalCount int         `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}
