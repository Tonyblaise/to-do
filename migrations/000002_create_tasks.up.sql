CREATE TYPE task_status AS ENUM ('pending', 'in_progress', 'completed', 'archived');
CREATE TYPE task_priority AS ENUM ('low', 'medium', 'high');

CREATE TABLE tasks (
    id          TEXT PRIMARY KEY DEFAULT uuid_generate_v4()::TEXT,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id   TEXT REFERENCES tasks(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      task_status NOT NULL DEFAULT 'pending',
    priority    task_priority NOT NULL DEFAULT 'medium',
    due_date    TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tasks_user_id       ON tasks(user_id);
CREATE INDEX idx_tasks_user_status   ON tasks(user_id, status)    WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_user_priority ON tasks(user_id, priority)  WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_parent_id     ON tasks(parent_id)          WHERE deleted_at IS NULL AND parent_id IS NOT NULL;
CREATE INDEX idx_tasks_updated_at    ON tasks(user_id, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_due_date      ON tasks(user_id, due_date)  WHERE deleted_at IS NULL AND status <> 'completed' AND due_date IS NOT NULL;

CREATE INDEX idx_tasks_search ON tasks USING gin(
  to_tsvector('english', title || ' ' || coalesce(description, ''))
) WHERE deleted_at IS NULL;