-- +goose Up
-- modify "comments" table
ALTER TABLE "comments" ADD COLUMN "import_source" character varying NULL, ADD COLUMN "import_external_id" character varying NULL;
-- create index "comment_issue_id_import_source_import_external_id" to table: "comments"
CREATE UNIQUE INDEX "comment_issue_id_import_source_import_external_id" ON "comments" ("issue_id", "import_source", "import_external_id") WHERE ((import_source IS NOT NULL) AND (import_external_id IS NOT NULL));
-- modify "projects" table
ALTER TABLE "projects" ADD COLUMN "import_source" character varying NULL, ADD COLUMN "import_external_id" character varying NULL;
-- create index "project_workspace_id_import_source_import_external_id" to table: "projects"
CREATE UNIQUE INDEX "project_workspace_id_import_source_import_external_id" ON "projects" ("workspace_id", "import_source", "import_external_id") WHERE ((import_source IS NOT NULL) AND (import_external_id IS NOT NULL));

-- +goose Down
-- reverse: create index "project_workspace_id_import_source_import_external_id" to table: "projects"
DROP INDEX "project_workspace_id_import_source_import_external_id";
-- reverse: modify "projects" table
ALTER TABLE "projects" DROP COLUMN "import_external_id", DROP COLUMN "import_source";
-- reverse: create index "comment_issue_id_import_source_import_external_id" to table: "comments"
DROP INDEX "comment_issue_id_import_source_import_external_id";
-- reverse: modify "comments" table
ALTER TABLE "comments" DROP COLUMN "import_external_id", DROP COLUMN "import_source";
