CREATE TABLE tags (
    id         TEXT PRIMARY KEY DEFAULT uuid_generate_v4()::TEXT,
    user_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    color      TEXT NOT NULL DEFAULT '#6366f1',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE TABLE task_tags (
    task_id TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    tag_id  TEXT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, tag_id)
);

CREATE INDEX idx_tags_user_id      ON tags(user_id);
CREATE INDEX idx_task_tags_tag_id  ON task_tags(tag_id);
CREATE INDEX idx_task_tags_task_id ON task_tags(task_id);