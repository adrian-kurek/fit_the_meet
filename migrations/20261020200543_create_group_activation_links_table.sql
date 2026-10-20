-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS group_participants 
(
    id           SERIAL PRIMARY KEY,
    group_id      INTEGER      NOT NULL
        constraint fk_group
            references groups (id)
            ON DELETE cascade,
    code            VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS group_participants;
-- +goose StatementEnd
