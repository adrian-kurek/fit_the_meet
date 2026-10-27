-- +goose Up
-- +goose StatementBegin
ALTER TABLE poll_answers
    ADD CONSTRAINT poll_answers_option_user_key
    UNIQUE (poll_option_id, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE poll_answers
    DROP CONSTRAINT IF EXISTS poll_answers_option_user_key;
-- +goose StatementEnd
