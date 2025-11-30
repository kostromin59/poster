-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS posts_sources (
  source TEXT NOT NULL,
  post_id UUID NOT NULL,

  PRIMARY KEY(source, post_id),
  
  FOREIGN KEY(source) REFERENCES sources(source) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
