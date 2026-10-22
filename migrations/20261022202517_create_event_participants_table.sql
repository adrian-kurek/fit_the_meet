-- +goose Up
-- +goose StatementBegin
CREATE TABLE event_participants (
    event_id      INTEGER      NOT NULL
        constraint fk_event
            references events (id)
            ON DELETE cascade,
    user_id      INTEGER      NOT NULL
        constraint fk_user
            references users (id)
            ON DELETE cascade,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS event_participants;
-- +goose StatementEnd
