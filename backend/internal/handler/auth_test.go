package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskflow/backend/internal/handler"
	"taskflow/backend/internal/middleware"
)

var testDB *pgxpool.Pool

const testJWTSecret = "test-secret-key-for-testing"

func TestMain(m *testing.M) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://taskflow:taskflow@localhost:5432/taskflow_test?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	testDB, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to test database: %v\n", err)
		fmt.Fprintf(os.Stderr, "Skipping integration tests. Set TEST_DATABASE_URL or run PostgreSQL.\n")
		os.Exit(0)
	}

	if err := testDB.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Cannot ping test database: %v\n", err)
		os.Exit(0)
	}

	// Set up schema
	setupTestDB(ctx)

	code := m.Run()

	// Cleanup
	teardownTestDB(ctx)
	testDB.Close()
	os.Exit(code)
}

func setupTestDB(ctx context.Context) {
	queries := []string{
		`CREATE TYPE IF NOT EXISTS task_status AS ENUM ('todo', 'in_progress', 'done')`,
		`CREATE TYPE IF NOT EXISTS task_priority AS ENUM ('low', 'medium', 'high')`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status task_status NOT NULL DEFAULT 'todo',
			priority task_priority NOT NULL DEFAULT 'medium',
			project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
			due_date DATE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
	}

	for _, q := range queries {
		if _, err := testDB.Exec(ctx, q); err != nil {
			// Types might already exist — ignore errors for CREATE TYPE IF NOT EXISTS
			continue
		}
	}
}

func teardownTestDB(ctx context.Context) {
	testDB.Exec(ctx, "DELETE FROM tasks")
	testDB.Exec(ctx, "DELETE FROM projects")
	testDB.Exec(ctx, "DELETE FROM users")
}

func cleanTables(ctx context.Context) {
	testDB.Exec(ctx, "DELETE FROM tasks")
	testDB.Exec(ctx, "DELETE FROM projects")
	testDB.Exec(ctx, "DELETE FROM users")
}

// ============================
// Test 1: User Registration
// ============================
func TestRegister(t *testing.T) {
	ctx := context.Background()
	cleanTables(ctx)

	authHandler := handler.NewAuthHandler(testDB, testJWTSecret)

	body := `{"name":"Test User","email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authHandler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, ok := resp["token"]; !ok {
		t.Error("response missing 'token' field")
	}

	user, ok := resp["user"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing 'user' object")
	}

	if user["email"] != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%v'", user["email"])
	}
	if user["name"] != "Test User" {
		t.Errorf("expected name 'Test User', got '%v'", user["name"])
	}

	t.Log("PASS: User registration returns 201 with token and user")
}

// ============================
// Test 2: User Registration with duplicate email
// ============================
func TestRegisterDuplicateEmail(t *testing.T) {
	ctx := context.Background()
	cleanTables(ctx)

	authHandler := handler.NewAuthHandler(testDB, testJWTSecret)

	// Register first user
	body := `{"name":"User One","email":"dup@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("first register failed: %d", w.Code)
	}

	// Try to register with same email
	body2 := `{"name":"User Two","email":"dup@example.com","password":"password456"}`
	req2 := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	authHandler.Register(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for duplicate email, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp)

	if resp["error"] != "validation failed" {
		t.Errorf("expected error 'validation failed', got '%v'", resp["error"])
	}

	t.Log("PASS: Duplicate email registration returns 400 validation error")
}

// ============================
// Test 3: Login with correct credentials
// ============================
func TestLoginSuccess(t *testing.T) {
	ctx := context.Background()
	cleanTables(ctx)

	authHandler := handler.NewAuthHandler(testDB, testJWTSecret)

	// Register a user first
	regBody := `{"name":"Login User","email":"login@example.com","password":"mypassword"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	authHandler.Register(regW, regReq)

	if regW.Code != http.StatusCreated {
		t.Fatalf("registration failed: %d", regW.Code)
	}

	// Login
	loginBody := `{"email":"login@example.com","password":"mypassword"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	authHandler.Login(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", loginW.Code, loginW.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &resp)

	if _, ok := resp["token"]; !ok {
		t.Error("response missing 'token' field")
	}

	user := resp["user"].(map[string]interface{})
	if user["email"] != "login@example.com" {
		t.Errorf("expected email 'login@example.com', got '%v'", user["email"])
	}

	t.Log("PASS: Login with correct credentials returns 200 with token")
}

// ============================
// Test 4: Login with wrong password
// ============================
func TestLoginWrongPassword(t *testing.T) {
	ctx := context.Background()
	cleanTables(ctx)

	authHandler := handler.NewAuthHandler(testDB, testJWTSecret)

	// Register
	regBody := `{"name":"Wrong PW User","email":"wrong@example.com","password":"correctpassword"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	authHandler.Register(regW, regReq)

	// Login with wrong password
	loginBody := `{"email":"wrong@example.com","password":"wrongpassword"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	authHandler.Login(loginW, loginReq)

	if loginW.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", loginW.Code, loginW.Body.String())
	}

	t.Log("PASS: Login with wrong password returns 401")
}

// ============================
// Test 5: Create project (authenticated)
// ============================
func TestCreateProject(t *testing.T) {
	ctx := context.Background()
	cleanTables(ctx)

	authHandler := handler.NewAuthHandler(testDB, testJWTSecret)
	projectHandler := handler.NewProjectHandler(testDB)

	// Register user and get token
	regBody := `{"name":"Project User","email":"proj@example.com","password":"password123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	authHandler.Register(regW, regReq)

	var authResp map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &authResp)
	token := authResp["token"].(string)
	user := authResp["user"].(map[string]interface{})
	userID := user["id"].(string)

	// Set up router with auth middleware
	r := chi.NewRouter()
	r.Use(middleware.Auth(testJWTSecret))
	r.Post("/projects", projectHandler.Create)

	// Create project
	body := `{"name":"My Project","description":"A test project"}`
	req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var project map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &project)

	if project["name"] != "My Project" {
		t.Errorf("expected name 'My Project', got '%v'", project["name"])
	}
	if project["owner_id"] != userID {
		t.Errorf("expected owner_id '%s', got '%v'", userID, project["owner_id"])
	}

	t.Log("PASS: Authenticated project creation returns 201 with correct owner")
}

// ============================
// Test 6: Validation errors return proper format
// ============================
func TestRegisterValidation(t *testing.T) {
	ctx := context.Background()
	cleanTables(ctx)

	authHandler := handler.NewAuthHandler(testDB, testJWTSecret)

	// Empty body — missing all fields
	body := `{"name":"","email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["error"] != "validation failed" {
		t.Errorf("expected error 'validation failed', got '%v'", resp["error"])
	}

	fields, ok := resp["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("response missing 'fields' object")
	}

	if _, ok := fields["name"]; !ok {
		t.Error("missing validation error for 'name'")
	}
	if _, ok := fields["email"]; !ok {
		t.Error("missing validation error for 'email'")
	}
	if _, ok := fields["password"]; !ok {
		t.Error("missing validation error for 'password'")
	}

	t.Log("PASS: Validation errors return 400 with structured fields object")
}
