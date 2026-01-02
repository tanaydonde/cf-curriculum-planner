CREATE TABLE IF NOT EXISTS problems (
    problem_id TEXT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    rating INT NOT NULL,
    tags TEXT[]
);

DO $$ BEGIN 
    CREATE TYPE solve_status as ENUM ('solved', 'partially_solved', 'failed');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS user_logs (
    id SERIAL PRIMARY KEY,
    handle TEXT NOT NULL,
    problem_id TEXT REFERENCES problems(problem_id),
    status solve_status NOT NULL,

    -- time_spent_minutes is NULL when data is unavailable
    time_spent_minutes INT,

    submission_count INT DEFAULT 1,
    is_api_synced BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS topics (
    id SERIAL PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS topic_dependencies (
    parent_id INTEGER REFERENCES topics(id) ON DELETE CASCADE,
    child_id INTEGER REFERENCES topics(id) ON DELETE CASCADE,
    PRIMARY KEY (parent_id, child_id),
    CONSTRAINT no_self_referencing CHECK (parent_id <> child_id)
);

CREATE TABLE IF NOT EXISTS user_topic_stats (
    handle TEXT NOT NULL,
    topic_slug TEXT NOT NULL REFERENCES topics(slug),
    mastery_score FLOAT DEFAULT 0,
    peak_score FLOAT DEFAULT 0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (handle, topic_slug)
);

CREATE TABLE IF NOT EXISTS user_interval_stats (
    handle TEXT NOT NULL,
    topic_slug TEXT NOT NULL REFERENCES topics(slug),

    bin_idx INT NOT NULL,
    bin_score FLOAT NOT NULL,

    credits FLOAT[] NOT NULL DEFAULT '{}',
    multipliers FLOAT[] NOT NULL DEFAULT '{}',

    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (handle, topic_slug, bin_idx)
);

CREATE TABLE IF NOT EXISTS user_problems(
    handle TEXT NOT NULL,
    problem_id TEXT NOT NULL,
    status TEXT NOT NULL,
    last_attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (handle, problem_id)
);

CREATE INDEX idx_user_problems_status ON user_problems(handle, status);