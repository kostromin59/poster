-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sources (
  source TEXT NOT NULL PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
