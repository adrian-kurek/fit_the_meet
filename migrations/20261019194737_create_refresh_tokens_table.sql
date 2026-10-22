-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS refresh_tokens
(
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER      not null
        constraint fk_user
            references users (id)
            on delete cascade,
    token_hash   VARCHAR(255) not null unique,
    device_info  VARCHAR(255),
    ip_address   VARCHAR(45),
    created_at   TIMESTAMP with time zone default CURRENT_TIMESTAMP,
    expires_at   TIMESTAMP with time zone,
    last_used_at TIMESTAMP with time zone
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_tokens;
-- +goose StatementEnd
