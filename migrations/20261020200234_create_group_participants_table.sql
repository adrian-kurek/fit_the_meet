-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS group_participants 
(
    user_id      INTEGER      NOT NULL
        constraint fk_user
            references users (id)
            ON DELETE cascade,

    group_id      INTEGER      NOT NULL 
        constraint fk_group
            references groups (id)
            ON DELETE cascade,
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS group_participants;
-- +goose StatementEnd
