-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS posts_tags (
  tag TEXT NOT NULL,
  post_id UUID NOT NULL,

  PRIMARY KEY(tag, post_id),
  
  FOREIGN KEY(tag) REFERENCES tags(tag) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
