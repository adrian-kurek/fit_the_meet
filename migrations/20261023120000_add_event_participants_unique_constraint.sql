-- +goose Up
-- +goose StatementBegin
ALTER TABLE event_participants
    ADD CONSTRAINT event_participants_event_user_date_key
    UNIQUE (event_id, user_id, date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE event_participants
    DROP CONSTRAINT IF EXISTS event_participants_event_user_date_key;
-- +goose StatementEnd
