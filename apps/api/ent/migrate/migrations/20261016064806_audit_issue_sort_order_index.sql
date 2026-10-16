-- +goose Up
-- create index "issue_workspace_id_sort_order" to table: "issues"
CREATE INDEX "issue_workspace_id_sort_order" ON "issues" ("workspace_id", "sort_order");

-- +goose Down
-- reverse: create index "issue_workspace_id_sort_order" to table: "issues"
DROP INDEX "issue_workspace_id_sort_order";
