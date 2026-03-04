# Go To-Do API

A RESTful API for managing tasks, built with Go, PostgreSQL, and WebSockets.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Setup](#setup)
- [Running the Server](#running-the-server)
- [Authentication](#authentication)
- [Endpoints](#endpoints)
  - [Health](#health)
  - [Auth](#auth)
  - [Tasks](#tasks)
  - [Tags](#tags)
  - [Attachments](#attachments)
  - [WebSocket](#websocket)

---

## Prerequisites

- Go 1.21+
- PostgreSQL running locally (or a hosted instance)

---

## Setup

**1. Clone the repository**

```bash
git clone https://github.com/Tonyblaise/to-do.git
cd to-do
```

**2. Install dependencies**

```bash
go mod download
```

**3. Configure environment variables**

Copy the example env file and fill in your values:

```bash
cp .env.example .env
```

| Variable              | Default                                                              | Description                               |
|-----------------------|----------------------------------------------------------------------|-------------------------------------------|
| `PORT`                | `8080`                                                               | Port the server listens on                |
| `ENV`                 | `development`                                                        | Environment (`development`/`production`)  |
| `DATABASE_URL`        | `postgres://postgres:postgres@localhost:5432/gotodo?sslmode=disable` | PostgreSQL connection string              |
| `JWT_SECRET`          | `secret`                                                             | Secret key for signing JWT tokens         |
| `JWT_EXPIRY`          | `24h`                                                                | Token expiry duration                     |
| `ALLOWED_ORIGINS`     | `http://localhost:3000,http://localhost:5173`                        | Comma-separated CORS allowed origins      |
| `RATE_LIMIT_REQUESTS` | `100`                                                                | Max requests per window                   |
| `RATE_LIMIT_WINDOW`   | `1m`                                                                 | Rate limit window duration                |
| `STORAGE_PATH`        | `./uploads`                                                          | Directory for file uploads                |
| `MAX_UPLOAD_SIZE`     | `10485760`                                                           | Max upload size in bytes (10 MB)          |
| `SMTP_HOST`           | `localhost`                                                          | SMTP server host                          |
| `SMTP_PORT`           | `587`                                                                | SMTP server port                          |
| `SMTP_FROM`           | `noreply@gotodo.com`                                                 | From address for emails                   |

**4. Create the database**

```bash
createdb gotodo
```

Migrations run automatically on startup.

---

## Running the Server

```bash
go run ./cmd/api
```

The server starts on `http://localhost:8080` by default.

---

## Authentication

Protected endpoints require a Bearer token in the `Authorization` header:

```
Authorization: Bearer <your_jwt_token>
```

You obtain a token by signing up or logging in.

---

## Endpoints

Base URL: `http://localhost:8080/api/v1`

---

### Health

#### `GET /health`

Check server and database status. No authentication required.

**Response `200 OK`**
```json
{
  "status": "ok",
  "database": "ok",
  "version": "1.0.0"
}
```

---

### Auth

#### `POST /auth/signup`

Register a new user account.

**Request body**
```json
{
  "first_name": "Jane",
  "last_name": "Doe",
  "email": "jane@example.com",
  "password": "securepassword"
}
```

**Response `201 Created`**
```json
{
  "token": "<jwt_token>",
  "user": {
    "id": "uuid",
    "email": "jane@example.com",
    "first_name": "Jane",
    "last_name": "Doe",
    "created_at": "2026-03-04T10:00:00Z",
    "updated_at": "2026-03-04T10:00:00Z"
  }
}
```

**Errors**
- `400 Bad Request` — missing or invalid fields
- `409 Conflict` — email already registered

---

#### `POST /auth/login`

Log in with existing credentials.

**Request body**
```json
{
  "email": "jane@example.com",
  "password": "securepassword"
}
```

**Response `200 OK`**
```json
{
  "token": "<jwt_token>",
  "user": {
    "id": "uuid",
    "email": "jane@example.com",
    "first_name": "Jane",
    "last_name": "Doe",
    "created_at": "2026-03-04T10:00:00Z",
    "updated_at": "2026-03-04T10:00:00Z"
  }
}
```

**Errors**
- `400 Bad Request` — invalid JSON body
- `401 Unauthorized` — invalid credentials

---

### Tasks

All task endpoints require authentication.

#### `POST /tasks`

Create a new task.

**Request body**
```json
{
  "title": "Buy groceries",
  "description": "Milk, eggs, bread",
  "priority": "medium",
  "due_date": "2026-03-10T18:00:00Z",
  "parent_id": null,
  "tag_ids": ["tag-uuid-1"]
}
```

| Field         | Type           | Required | Description                                    |
|---------------|----------------|----------|------------------------------------------------|
| `title`       | string         | Yes      | Task title                                     |
| `description` | string         | No       | Task description                               |
| `priority`    | string         | No       | `low`, `medium`, or `high` (default: `medium`) |
| `due_date`    | RFC3339 string | No       | Due date/time                                  |
| `parent_id`   | string (UUID)  | No       | ID of the parent task (for subtasks)           |
| `tag_ids`     | array of UUIDs | No       | Tags to associate with the task                |

**Response `201 Created`** — returns the created `Task` object.

---

#### `GET /tasks`

List tasks for the authenticated user with optional filtering and cursor-based pagination.

**Query parameters**

| Parameter  | Type   | Description                                                                 |
|------------|--------|-----------------------------------------------------------------------------|
| `search`   | string | Search by title/description                                                 |
| `status`   | string | Filter by status: `pending`, `in_progress`, `completed`, `archived`        |
| `priority` | string | Filter by priority: `low`, `medium`, `high`                                 |
| `limit`    | int    | Number of results to return                                                 |
| `cursor`   | string | Pagination cursor from a previous response                                  |

**Example**
```
GET /api/v1/tasks?status=pending&priority=high&limit=20
```

**Response `200 OK`**
```json
{
  "tasks": [...],
  "next_cursor": "cursor_string",
  "total": 42
}
```

---

#### `GET /tasks/{id}`

Get a single task by its ID.

**Response `200 OK`** — returns a `Task` object including subtasks, tags, and attachments.

**Errors**
- `404 Not Found` — task not found or does not belong to user

---

#### `PATCH /tasks/{id}`

Update a task's details.

**Request body** (all fields optional)
```json
{
  "title": "Updated title",
  "description": "Updated description",
  "priority": "high",
  "due_date": "2026-04-01T09:00:00Z",
  "tag_ids": ["tag-uuid-1", "tag-uuid-2"]
}
```

**Response `200 OK`** — returns the updated `Task` object.

**Errors**
- `404 Not Found` — task not found

---

#### `PATCH /tasks/{id}/status`

Update only the status of a task.

**Request body**
```json
{
  "status": "in_progress"
}
```

Valid status values: `pending`, `in_progress`, `completed`, `archived`

**Response `200 OK`** — returns the updated `Task` object.

**Errors**
- `400 Bad Request` — invalid status value
- `404 Not Found` — task not found

---

#### `DELETE /tasks/{id}`

Soft-delete a task (sets `deleted_at`; data is preserved in the database).

**Response `204 No Content`**

**Errors**
- `404 Not Found` — task not found

---

#### `PATCH /tasks/bulk`

Update multiple tasks at once.

**Request body**
```json
{
  "task_ids": ["uuid-1", "uuid-2"],
  "update": {
    "priority": "low",
    "tag_ids": []
  }
}
```

**Response `204 No Content`**

---

#### `DELETE /tasks/bulk`

Delete multiple tasks at once.

**Request body**
```json
{
  "task_ids": ["uuid-1", "uuid-2"]
}
```

**Response `204 No Content`**

---

#### `GET /tasks/export`

Export all tasks as a CSV file download.

**Response `200 OK`**
- `Content-Type: text/csv`
- `Content-Disposition: attachment; filename=tasks.csv`

---

#### `GET /tasks/sync`

Fetch all task changes (creates, updates, deletes) since a given timestamp. Useful for offline sync.

**Query parameters**

| Parameter        | Type           | Required | Description                                |
|------------------|----------------|----------|--------------------------------------------|
| `last_synced_at` | RFC3339 string | Yes      | Fetch changes that occurred after this time |

**Example**
```
GET /api/v1/tasks/sync?last_synced_at=2026-03-01T00:00:00Z
```

**Response `200 OK`**
```json
{
  "records": [
    {
      "task": { ... },
      "action": "created"
    },
    {
      "task": { ... },
      "action": "modified"
    },
    {
      "task": { ... },
      "action": "deleted"
    }
  ],
  "last_synced_at": "2026-03-04T10:00:00Z"
}
```

**Errors**
- `400 Bad Request` — `last_synced_at` is missing or not a valid RFC3339 timestamp

---

### Tags

All tag endpoints require authentication.

#### `POST /tags`

Create a new tag.

**Request body**
```json
{
  "name": "Work",
  "color": "#6366f1"
}
```

| Field   | Type   | Required | Description                           |
|---------|--------|----------|---------------------------------------|
| `name`  | string | Yes      | Tag name (must be unique per user)    |
| `color` | string | No       | Hex color code (default: `#6366f1`)   |

**Response `201 Created`**
```json
{
  "id": "uuid",
  "user_id": "uuid",
  "name": "Work",
  "color": "#6366f1",
  "created_at": "2026-03-04T10:00:00Z"
}
```

**Errors**
- `400 Bad Request` — name is missing
- `409 Conflict` — tag name already exists

---

#### `GET /tags`

List all tags for the authenticated user.

**Response `200 OK`**
```json
[
  {
    "id": "uuid",
    "user_id": "uuid",
    "name": "Work",
    "color": "#6366f1",
    "created_at": "2026-03-04T10:00:00Z"
  }
]
```

---

#### `DELETE /tags/{id}`

Delete a tag by ID.

**Response `204 No Content`**

**Errors**
- `404 Not Found` — tag not found

---

### Attachments

All attachment endpoints require authentication.

#### `POST /tasks/{id}/attachments`

Upload a file attachment to a task. Uses `multipart/form-data`.

**Form field**

| Field  | Type | Required | Description        |
|--------|------|----------|--------------------|
| `file` | file | Yes      | The file to upload |

**Allowed MIME types:** `image/jpeg`, `image/png`, `image/gif`, `application/pdf`

**Max size:** 10 MB (configurable via `MAX_UPLOAD_SIZE`)

**Example using curl**
```bash
curl -X POST http://localhost:8080/api/v1/tasks/<task-id>/attachments \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/document.pdf"
```

**Response `201 Created`**
```json
{
  "id": "uuid",
  "task_id": "uuid",
  "user_id": "uuid",
  "filename": "document.pdf",
  "mime_type": "application/pdf",
  "size": 204800,
  "url": "/api/v1/attachments/uuid",
  "created_at": "2026-03-04T10:00:00Z"
}
```

**Errors**
- `400 Bad Request` — unsupported file type, file too large, or missing file field
- `404 Not Found` — task not found

---

#### `GET /attachments/{id}`

Download an attachment file by its ID.

**Response `200 OK`** — file is streamed with the original `Content-Type` and filename as `Content-Disposition`.

**Errors**
- `404 Not Found` — attachment not found

---

#### `DELETE /attachments/{id}`

Delete an attachment by its ID. Also removes the file from disk.

**Response `204 No Content`**

**Errors**
- `404 Not Found` — attachment not found

---

### WebSocket

#### `GET /ws`

Connect to the real-time WebSocket for live task updates. Requires a valid JWT token passed as a query parameter or `Authorization` header.

**URL**
```
ws://localhost:8080/ws
```

Once connected, the server pushes the following event types to the authenticated user:

| Event type             | Description                  | Payload            |
|------------------------|------------------------------|--------------------|
| `task.created`         | A new task was created       | `Task` object      |
| `task.updated`         | A task was updated           | `Task` object      |
| `task.status_changed`  | A task's status was changed  | `Task` object      |
| `task.deleted`         | A task was soft-deleted      | `{ "id": "uuid" }` |

**Message format**
```json
{
  "type": "task.created",
  "payload": { ... }
}
```

Events are only broadcast to the authenticated user's own connections.
