CREATE TABLE IF NOT EXISTS users (
    user_id   TEXT PRIMARY KEY,
    username  TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );


CREATE TABLE IF NOT EXISTS teams (
    id         BIGSERIAL PRIMARY KEY,
    team_name  TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS team_members (
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id TEXT   NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, user_id)
);


CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id         TEXT NOT NULL REFERENCES users(user_id),
    status            TEXT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    merged_at         TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_team_members_user_id
    ON team_members (user_id);

CREATE INDEX IF NOT EXISTS idx_pull_request_reviewers_reviewer_id
    ON pull_request_reviewers (reviewer_id);

CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id
    ON pull_requests (author_id);
