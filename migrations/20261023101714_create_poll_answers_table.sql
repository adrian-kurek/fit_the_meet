-- +goose Up
-- +goose StatementBegin
CREATE TABLE poll_answers (
    id SERIAL PRIMARY KEY,
    poll_option_id      INTEGER      NOT NULL
        constraint fk_poll_option
            references poll_options (id)
            on delete cascade,
    user_id      INTEGER      NOT NULL
        constraint fk_user
            references users (id)
            on delete cascade,
    is_active            BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS poll_answers;
-- +goose StatementEnd
