-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS polls (
    id SERIAL PRIMARY KEY,
    event_id      INTEGER      NOT NULL
        constraint fk_event
            references events (id)
            on delete cascade,
    is_active            BOOLEAN DEFAULT TRUE,
    created_by      INTEGER      NOT NULL
        constraint fk_user
            references users (id)
            on delete cascade,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS polls;
-- +goose StatementEnd
