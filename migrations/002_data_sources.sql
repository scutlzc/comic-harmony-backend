-- 002_data_sources.sql
-- 数据源配置持久化

CREATE TABLE IF NOT EXISTS data_sources (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        VARCHAR(128) NOT NULL,
    source_type VARCHAR(32) NOT NULL,  -- 'komga', 'webdav', 'clouddrive'
    url         TEXT NOT NULL,
    username    VARCHAR(128) DEFAULT '',
    password    TEXT DEFAULT '',       -- base64 encoded
    root_path   TEXT DEFAULT '/',
    enabled     BOOLEAN DEFAULT true,
    last_health TIMESTAMPTZ,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_datasources_user ON data_sources(user_id);
CREATE INDEX IF NOT EXISTS idx_datasources_type ON data_sources(source_type);
