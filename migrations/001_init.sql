-- 001_init.sql
-- ComicHarmony 数据库初始化

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS categories (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    sort_order  INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS comics (
    id             BIGSERIAL PRIMARY KEY,
    title          VARCHAR(255) NOT NULL,
    author         VARCHAR(128) DEFAULT '',
    description    TEXT DEFAULT '',
    cover_url      TEXT DEFAULT '',
    status         SMALLINT DEFAULT 0,  -- 0=连载, 1=完结
    total_views    BIGINT DEFAULT 0,
    total_likes    BIGINT DEFAULT 0,
    total_favorites BIGINT DEFAULT 0,
    category_id    BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    updated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chapters (
    id          BIGSERIAL PRIMARY KEY,
    comic_id    BIGINT NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    sort_order  INT DEFAULT 0,
    page_count  INT DEFAULT 0,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(64) UNIQUE NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    avatar_url  TEXT DEFAULT '',
    status      SMALLINT DEFAULT 1,  -- 1=active, 0=banned
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_comics_category ON comics(category_id);
CREATE INDEX IF NOT EXISTS idx_comics_updated ON comics(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_comics_views ON comics(total_views DESC);
CREATE INDEX IF NOT EXISTS idx_chapters_comic ON chapters(comic_id);
CREATE INDEX IF NOT EXISTS idx_chapters_order ON chapters(comic_id, sort_order);

-- Seed data
INSERT INTO categories (name, description, sort_order) VALUES
    ('热血', '热血少年漫画', 1),
    ('恋爱', '恋爱少女漫画', 2),
    ('科幻', '科幻题材漫画', 3),
    ('悬疑', '悬疑推理漫画', 4),
    ('搞笑', '搞笑日常漫画', 5)
ON CONFLICT DO NOTHING;
