-- 003_favorites_history.sql
-- 收藏 + 阅读历史

CREATE TABLE IF NOT EXISTS favorites (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comic_id   BIGINT NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, comic_id)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user ON favorites(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS reading_history (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comic_id     BIGINT NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
    chapter_id   BIGINT REFERENCES chapters(id) ON DELETE SET NULL,
    chapter_title VARCHAR(255) DEFAULT '',
    page         INT DEFAULT 1,
    total_pages  INT DEFAULT 0,
    progress     REAL DEFAULT 0.0,  -- 0.0 ~ 1.0
    read_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, comic_id)
);

CREATE INDEX IF NOT EXISTS idx_history_user ON reading_history(user_id, read_at DESC);
