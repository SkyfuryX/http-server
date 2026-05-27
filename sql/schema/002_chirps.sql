-- +goose Up
CREATE TABLE chirps (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    body TEXT NOT NULL,
    user_id UUID NOT NULL references users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE chirps;

