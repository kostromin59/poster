-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tags (
  value TEXT NOT NULL PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
