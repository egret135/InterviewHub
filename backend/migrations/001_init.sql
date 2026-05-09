CREATE TABLE IF NOT EXISTS categories (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon        VARCHAR(50),
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS questions (
    id          SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES categories(id),
    title       VARCHAR(500) NOT NULL,
    difficulty   VARCHAR(20) NOT NULL DEFAULT 'medium',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_questions_title_cat ON questions(title, category_id);

CREATE TABLE IF NOT EXISTS answers (
    id           SERIAL PRIMARY KEY,
    question_id  INT NOT NULL UNIQUE REFERENCES questions(id),
    content_md   TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tags (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    slug VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS question_tags (
    question_id INT NOT NULL REFERENCES questions(id),
    tag_id      INT NOT NULL REFERENCES tags(id),
    PRIMARY KEY (question_id, tag_id)
);

-- Full-text search
ALTER TABLE questions ADD COLUMN IF NOT EXISTS search_vector tsvector;
CREATE INDEX IF NOT EXISTS idx_questions_search ON questions USING GIN(search_vector);
