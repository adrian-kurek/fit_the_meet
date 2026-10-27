-- +goose Up
-- +goose StatementBegin
ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL;

DELETE FROM group_participants duplicate
USING group_participants original
WHERE duplicate.ctid < original.ctid
    AND duplicate.group_id = original.group_id
    AND duplicate.user_id = original.user_id;

ALTER TABLE group_participants
    ADD CONSTRAINT group_participants_group_user_key
    UNIQUE (group_id, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE group_participants
    DROP CONSTRAINT IF EXISTS group_participants_group_user_key;

ALTER TABLE groups
    DROP COLUMN IF EXISTS created_at;
-- +goose StatementEnd
