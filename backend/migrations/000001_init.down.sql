-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

-- Drop custom enum types
DROP TYPE IF EXISTS task_priority;
DROP TYPE IF EXISTS task_status;
