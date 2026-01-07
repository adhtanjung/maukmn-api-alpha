-- +goose Up
-- +goose StatementBegin
-- No schema changes required for address handling fix.
-- The addresses table and address_id foreign key already exist.
SELECT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
