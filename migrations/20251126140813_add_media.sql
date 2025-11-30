-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media (
  id UUID NOT NULL DEFAULT uuidv7() PRIMARY KEY,
  filetype TEXT NOT NULL,
  uri TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
