-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    description            TEXT,
    duration_minutes      INTEGER      NOT NULL,
    code            VARCHAR(255) NOT NULL,
    is_active            BOOLEAN DEFAULT TRUE,
    created_by      INTEGER      NOT NULL
        constraint fk_user
            references users (id)
            ON DELETE cascade,
    group_id      INTEGER
        constraint fk_group
            references groups (id)
            ON DELETE cascade,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
