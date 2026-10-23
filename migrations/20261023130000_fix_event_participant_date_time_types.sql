-- +goose Up
-- +goose StatementBegin
ALTER TABLE event_participants
    ALTER COLUMN date TYPE DATE USING date::date,
    ALTER COLUMN start_time TYPE TIME WITHOUT TIME ZONE USING start_time::time,
    ALTER COLUMN end_time TYPE TIME WITHOUT TIME ZONE USING end_time::time;

ALTER TABLE event_participants
    ADD CONSTRAINT event_participants_valid_time_range
    CHECK (start_time < end_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE event_participants
    DROP CONSTRAINT IF EXISTS event_participants_valid_time_range;

ALTER TABLE event_participants
    ALTER COLUMN start_time TYPE TIMESTAMP WITHOUT TIME ZONE USING date + start_time,
    ALTER COLUMN end_time TYPE TIMESTAMP WITHOUT TIME ZONE USING date + end_time;

ALTER TABLE event_participants
    ALTER COLUMN date TYPE TIMESTAMP WITHOUT TIME ZONE USING date::timestamp;
-- +goose StatementEnd
