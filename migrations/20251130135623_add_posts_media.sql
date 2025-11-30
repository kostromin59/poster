-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS posts_media (
  media_id UUID NOT NULL,
  post_id UUID NOT NULL,

  PRIMARY KEY(media_id, post_id),
  
  FOREIGN KEY(media_id) REFERENCES media(id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE ON UPDATE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
