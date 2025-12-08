-- +goose Up
-- +goose StatementBegin
INSERT INTO sources (source) VALUES ('Телеграмм') ON CONFLICT (source) DO NOTHING;
INSERT INTO sources (source) VALUES ('Вебсайт') ON CONFLICT (source) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd