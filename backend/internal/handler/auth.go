package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"taskflow/backend/internal/model"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	db        *pgxpool.Pool
	jwtSecret string
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(db *pgxpool.Pool, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

// Register handles user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate fields
	fields := make(map[string]string)
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Name == "" {
		fields["name"] = "is required"
	}
	if req.Email == "" {
		fields["email"] = "is required"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	} else if len(req.Password) < 6 {
		fields["password"] = "must be at least 6 characters"
	}
	if len(fields) > 0 {
		ValidationError(w, fields)
		return
	}

	// Check if email already exists
	var exists bool
	err := h.db.QueryRow(r.Context(), "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		slog.Error("failed to check email existence", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if exists {
		ValidationError(w, map[string]string{"email": "already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		slog.Error("failed to hash password", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Insert user
	var user model.User
	err = h.db.QueryRow(
		r.Context(),
		`INSERT INTO users (name, email, password) VALUES ($1, $2, $3)
		 RETURNING id, name, email, created_at`,
		req.Name, req.Email, string(hashedPassword),
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		slog.Error("failed to insert user", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Generate JWT
	token, err := h.generateToken(user)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("user registered", "user_id", user.ID, "email", user.Email)
	JSON(w, http.StatusCreated, model.AuthResponse{Token: token, User: user})
}

// Login handles user login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate fields
	fields := make(map[string]string)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Email == "" {
		fields["email"] = "is required"
	}
	if req.Password == "" {
		fields["password"] = "is required"
	}
	if len(fields) > 0 {
		ValidationError(w, fields)
		return
	}

	// Find user by email
	var user model.User
	err := h.db.QueryRow(
		r.Context(),
		"SELECT id, name, email, password, created_at FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt)
	if err != nil {
		slog.Warn("login attempt for unknown email", "email", req.Email)
		Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		slog.Warn("login attempt with wrong password", "email", req.Email)
		Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// Clear password before response
	user.Password = ""

	// Generate JWT
	token, err := h.generateToken(user)
	if err != nil {
		slog.Error("failed to generate token", "error", err)
		Error(w, http.StatusInternalServerError, "internal server error")
		return
	}

	slog.Info("user logged in", "user_id", user.ID, "email", user.Email)
	JSON(w, http.StatusOK, model.AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) generateToken(user model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// GetUserByID fetches a user by ID from the database.
func GetUserByID(ctx context.Context, db *pgxpool.Pool, userID string) (*model.User, error) {
	var user model.User
	err := db.QueryRow(
		ctx,
		"SELECT id, name, email, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
