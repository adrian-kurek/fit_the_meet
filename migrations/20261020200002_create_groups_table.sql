-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS groups 
(
    id           SERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    created_by      INTEGER      NOT NULL
        constraint fk_user
            references users (id)
            ON DELETE cascade,
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS groups;
-- +goose StatementEnd
