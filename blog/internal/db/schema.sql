-- schema applied by db.Apply(). Safe to re-run.

CREATE TABLE IF NOT EXISTS posts (
    id          INTEGER  PRIMARY KEY AUTOINCREMENT,
    title       TEXT     NOT NULL,
    slug        TEXT     NOT NULL UNIQUE,
    body        TEXT     NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published   INTEGER  NOT NULL DEFAULT 0  -- 0 = draft, 1 = published
);

CREATE INDEX IF NOT EXISTS idx_posts_slug
    ON posts(slug);

CREATE INDEX IF NOT EXISTS idx_posts_published_created
    ON posts(published, created_at DESC);
