# TaskFlow

A full-stack task management system built with Go, React, and PostgreSQL. Users can register, log in, create projects, add tasks, and assign them to team members.

## Tech Stack

| Layer     | Technology                                                    |
|-----------|---------------------------------------------------------------|
| Backend   | Go 1.23, Chi router, pgx (PostgreSQL driver), JWT, bcrypt    |
| Frontend  | React 19, TypeScript, Vite, Tailwind CSS v4, React Router v7 |
| Database  | PostgreSQL 16                                                 |
| Infra     | Docker, Docker Compose, nginx (frontend proxy)                |

## Architecture Decisions

**Backend structure** — The backend uses a layered architecture (`handler` → `model`) with Chi as the HTTP router. Chi was chosen for its lightweight, stdlib-compatible design. I skipped a full service/repository layer since the scope is small enough that handlers can query the database directly without becoming unwieldy. For a larger app, I'd add those layers.

**Authentication** — JWT with bcrypt (cost 12). Tokens are issued at login/register and validated via middleware. The JWT secret is loaded from environment variables (never hardcoded). I chose stateless JWT over sessions for simplicity and because the API is consumed by a SPA.

**Database** — PostgreSQL with golang-migrate for schema management. Custom enum types (`task_status`, `task_priority`) enforce valid values at the database level. All migrations have both up and down files. Indexes on all foreign keys and the status column for query performance.

**Frontend** — React with TypeScript and Tailwind CSS v4 (no component library). I built custom components to keep the bundle lean and demonstrate component design. State management uses React Context for auth and theme, with local state for everything else — no Redux needed at this scale.

**API design** — RESTful with proper HTTP status codes (400 validation, 401 unauth, 403 forbidden, 404 not found). List endpoints return paginated responses. Task update/delete endpoints use `/tasks/:id` (not project-scoped) since task IDs are globally unique.

**Docker** — Multi-stage builds for both backend (Go build → Alpine runtime) and frontend (Node build → nginx). The frontend nginx proxies `/api/` to the backend, so the SPA uses relative URLs in production.

**Tradeoffs made:**
- No WebSocket/SSE for real-time updates — would add complexity without matching the scope
- No file uploads or avatar support
- No refresh token rotation — single 24h JWT is sufficient for this use case
- Client-side filtering rather than server-side for assignee filter within a project

## Running Locally

Prerequisites: Docker and Docker Compose installed.

```bash
git clone https://github.com/your-name/taskflow
cd taskflow
cp .env.example .env
docker compose up --build
```

Once all three services are healthy:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080

## Running Migrations

Migrations run **automatically** when the backend container starts. No manual steps needed.

The migration tool used is [golang-migrate](https://github.com/golang-migrate/migrate). Migration files are in `backend/migrations/`.

## Test Credentials

A seed user is created automatically via migration:

```
Email:    test@example.com
Password: password123
```

The seed also creates 1 project ("Demo Project") with 3 tasks in different statuses.

## API Reference

All endpoints return `Content-Type: application/json`. Protected endpoints require `Authorization: Bearer <token>`.

### Authentication

| Method | Endpoint          | Description                           |
|--------|-------------------|---------------------------------------|
| POST   | `/auth/register`  | Register with name, email, password   |
| POST   | `/auth/login`     | Login, returns JWT token + user       |

**POST /auth/register**
```json
// Request
{ "name": "Jane Doe", "email": "jane@example.com", "password": "secret123" }

// Response 201
{ "token": "eyJ...", "user": { "id": "uuid", "name": "Jane Doe", "email": "jane@example.com", "created_at": "..." } }
```

**POST /auth/login**
```json
// Request
{ "email": "test@example.com", "password": "password123" }

// Response 200
{ "token": "eyJ...", "user": { "id": "uuid", "name": "Test User", "email": "test@example.com", "created_at": "..." } }
```

### Projects

| Method | Endpoint              | Description                                    |
|--------|-----------------------|------------------------------------------------|
| GET    | `/projects`           | List projects (owned or assigned tasks in)     |
| POST   | `/projects`           | Create a new project                           |
| GET    | `/projects/:id`       | Get project details with tasks                 |
| PATCH  | `/projects/:id`       | Update name/description (owner only)           |
| DELETE | `/projects/:id`       | Delete project and all tasks (owner only)      |
| GET    | `/projects/:id/stats` | Task counts by status and assignee             |

**GET /projects** supports pagination: `?page=1&limit=20`

### Tasks

| Method | Endpoint                 | Description                                       |
|--------|--------------------------|---------------------------------------------------|
| GET    | `/projects/:id/tasks`    | List tasks with optional `?status=` `?assignee=`  |
| POST   | `/projects/:id/tasks`    | Create a task in a project                        |
| PATCH  | `/tasks/:id`             | Update task fields (all optional)                 |
| DELETE | `/tasks/:id`             | Delete task (project owner or creator only)       |

**POST /projects/:id/tasks**
```json
// Request
{ "title": "Design homepage", "description": "...", "priority": "high", "due_date": "2025-04-15" }

// Response 201
{ "id": "uuid", "title": "Design homepage", "status": "todo", "priority": "high", ... }
```

**PATCH /tasks/:id**
```json
// Request (all fields optional)
{ "status": "done", "priority": "low" }

// Response 200 — updated task object
```

### Error Responses

```json
// 400 Validation
{ "error": "validation failed", "fields": { "email": "is required" } }

// 401 Unauthenticated
{ "error": "missing authorization header" }

// 403 Forbidden
{ "error": "only the project owner can delete this project" }

// 404 Not Found
{ "error": "not found" }
```

## What I'd Do With More Time

- **Integration tests** — Add tests for auth, project, and task endpoints using a test database. The handlers are structured to make this straightforward (inject `*pgxpool.Pool`).
- **Refresh tokens** — Add a refresh token flow with token rotation for better security.
- **Drag-and-drop** — Kanban board view with drag-and-drop to change task status columns.
- **Real-time updates** — WebSocket or SSE for live task updates when multiple users are on the same project.
- **Assignee search** — The task modal should fetch and display project members in the assignee dropdown. Currently it accepts a UUID but doesn't show a user picker.
- **Pagination UI** — The backend supports pagination but the frontend doesn't render page controls yet.
- **Rate limiting** — Add rate limiting middleware to protect auth endpoints.
- **CI/CD pipeline** — GitHub Actions for linting, testing, and building Docker images.
- **Better error boundaries** — React error boundaries to catch rendering errors gracefully.
- **Accessibility audit** — Full WCAG 2.1 AA compliance check.
