-- Seed data for TaskFlow
-- Password for test user is "password123" (bcrypt cost 12)
-- Hash generated with: bcrypt.GenerateFromPassword([]byte("password123"), 12)

INSERT INTO users (id, name, email, password, created_at)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'Test User',
    'test@example.com',
    '$2a$12$LJ3m4ys3LkBsRGNwpRFzheWCzRqMd/GVTUBqH8G.tJGQBfwnLTbGa',
    NOW()
) ON CONFLICT (email) DO NOTHING;

INSERT INTO projects (id, name, description, owner_id, created_at)
VALUES (
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'Demo Project',
    'A sample project for testing',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    NOW()
) ON CONFLICT DO NOTHING;

INSERT INTO tasks (id, title, description, status, priority, project_id, creator_id, assignee_id, due_date, created_at, updated_at)
VALUES
(
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a31',
    'Set up project repository',
    'Initialize the Git repository and set up the project structure',
    'done',
    'high',
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    '2025-01-15',
    NOW(),
    NOW()
),
(
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a32',
    'Design database schema',
    'Create the database schema for the application',
    'in_progress',
    'medium',
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    '2025-02-01',
    NOW(),
    NOW()
),
(
    'c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
    'Implement authentication',
    'Add JWT-based authentication to the API',
    'todo',
    'low',
    'b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    NULL,
    '2025-03-01',
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;
