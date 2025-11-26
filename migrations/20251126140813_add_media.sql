-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS media (
  id UUID NOT NULL DEFAULT uuidv7() PRIMARY KEY,
  filetype TEXT NOT NULL,
  uri TEXT NOT NULL,
  post_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE SET NULL ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
