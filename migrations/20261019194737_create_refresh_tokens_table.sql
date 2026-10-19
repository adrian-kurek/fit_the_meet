-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS refresh_tokens
(
    id           SERIAL PRIMARY KEY,
    user_id      integer      not null
        constraint fk_user
            references users (id)
            on delete cascade,
    token_hash   varchar(255) not null unique,
    device_info  varchar(255),
    ip_address   varchar(45),
    created_at   timestamp with time zone default CURRENT_TIMESTAMP,
    expires_at   timestamp with time zone,
    last_used_at timestamp with time zone
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_tokens;
-- +goose StatementEnd
