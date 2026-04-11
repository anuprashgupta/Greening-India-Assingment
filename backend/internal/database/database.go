package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Connect establishes a connection pool to the PostgreSQL database.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	slog.Info("connected to database")
	return pool, nil
}

// RunMigrations runs all pending database migrations from the migrations directory.
func RunMigrations(databaseURL string, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	slog.Info("database migrations applied successfully")
	return nil
}

// RunSeed inserts test data if it doesn't already exist.
// Uses Go's bcrypt at runtime to guarantee the password hash is correct.
func RunSeed(ctx context.Context, pool *pgxpool.Pool) error {
	// Check if seed user already exists
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = 'test@example.com')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check seed user: %w", err)
	}
	if exists {
		slog.Info("seed data already exists, skipping")
		return nil
	}

	// Hash password at runtime with Go's bcrypt — guaranteed to work
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	if err != nil {
		return fmt.Errorf("failed to hash seed password: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin seed transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert test user
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, name, email, password, created_at)
		VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Test User', 'test@example.com', $1, NOW())
		ON CONFLICT (email) DO NOTHING
	`, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to seed user: %w", err)
	}

	// Insert test project
	_, err = tx.Exec(ctx, `
		INSERT INTO projects (id, name, description, owner_id, created_at)
		VALUES ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Demo Project', 'A sample project for testing TaskFlow features', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NOW())
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed project: %w", err)
	}

	// Insert test tasks
	_, err = tx.Exec(ctx, `
		INSERT INTO tasks (id, title, description, status, priority, project_id, creator_id, assignee_id, due_date, created_at, updated_at)
		VALUES
		('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a31', 'Set up project repository', 'Initialize the Git repository and set up the project structure', 'done', 'high', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2025-01-15', NOW(), NOW()),
		('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a32', 'Design database schema', 'Create the database schema with proper relations and constraints', 'in_progress', 'medium', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '2025-02-01', NOW(), NOW()),
		('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'Implement authentication', 'Add JWT-based authentication to the API endpoints', 'todo', 'low', 'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL, '2025-03-01', NOW(), NOW())
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("failed to seed tasks: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit seed transaction: %w", err)
	}

	slog.Info("seed data inserted successfully")
	return nil
}
