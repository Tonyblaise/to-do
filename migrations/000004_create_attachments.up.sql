CREATE TABLE task_attachments (
    id         TEXT PRIMARY KEY DEFAULT uuid_generate_v4()::TEXT,
    task_id    TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename   TEXT NOT NULL,
    mime_type  TEXT NOT NULL,
    size       BIGINT NOT NULL DEFAULT 0,
    path       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_attachments_task_id ON task_attachments(task_id);
CREATE INDEX idx_attachments_user_id ON task_attachments(user_id);