-- +goose Up
-- +goose StatementBegin
create table puzzle_templates (
	id integer primary key autoincrement,
	seed integer not null unique,
	board text not null unique
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table puzzle_templates;
-- +goose StatementEnd
