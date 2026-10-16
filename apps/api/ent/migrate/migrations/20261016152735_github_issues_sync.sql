-- +goose Up
-- modify "github_integrations" table
ALTER TABLE "github_integrations" ADD COLUMN "token_encrypted" bytea NULL, ADD COLUMN "token_prefix" character varying NOT NULL DEFAULT '';
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "github_login" character varying NULL;
-- create index "users_github_login_key" to table: "users"
CREATE UNIQUE INDEX "users_github_login_key" ON "users" ("github_login");
-- create "github_repo_syncs" table
CREATE TABLE "github_repo_syncs" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "full_name" character varying NOT NULL,
  "sync_closed" boolean NOT NULL DEFAULT false,
  "last_error" character varying NOT NULL DEFAULT '',
  "last_synced_at" timestamptz NULL,
  "backfill_status" character varying NOT NULL DEFAULT 'none',
  "backfill_report" jsonb NULL,
  "integration_id" uuid NOT NULL,
  "project_id" uuid NULL,
  "team_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "github_repo_syncs_github_integrations_repo_syncs" FOREIGN KEY ("integration_id") REFERENCES "github_integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "github_repo_syncs_projects_github_repo_syncs" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "github_repo_syncs_teams_github_repo_syncs" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "github_repo_syncs_full_name_key" to table: "github_repo_syncs"
CREATE UNIQUE INDEX "github_repo_syncs_full_name_key" ON "github_repo_syncs" ("full_name");
-- create "github_issue_links" table
CREATE TABLE "github_issue_links" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "number" bigint NOT NULL,
  "node_id" character varying NOT NULL,
  "html_url" character varying NOT NULL,
  "github_updated_at" timestamptz NOT NULL,
  "last_outbound_at" timestamptz NULL,
  "repo_sync_id" uuid NOT NULL,
  "issue_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "github_issue_links_github_repo_syncs_issue_links" FOREIGN KEY ("repo_sync_id") REFERENCES "github_repo_syncs" ("id") ON UPDATE NO ACTION ON DELETE CASCADE,
  CONSTRAINT "github_issue_links_issues_github_link" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "github_issue_links_issue_id_key" to table: "github_issue_links"
CREATE UNIQUE INDEX "github_issue_links_issue_id_key" ON "github_issue_links" ("issue_id");
-- create index "github_issue_links_node_id_key" to table: "github_issue_links"
CREATE UNIQUE INDEX "github_issue_links_node_id_key" ON "github_issue_links" ("node_id");
-- create index "githubissuelink_repo_sync_id_number" to table: "github_issue_links"
CREATE INDEX "githubissuelink_repo_sync_id_number" ON "github_issue_links" ("repo_sync_id", "number");
-- create "github_comment_links" table
CREATE TABLE "github_comment_links" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "comment_id" uuid NOT NULL,
  "github_comment_id" bigint NOT NULL,
  "issue_link_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "github_comment_links_github_issue_links_comment_links" FOREIGN KEY ("issue_link_id") REFERENCES "github_issue_links" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "github_comment_links_comment_id_key" to table: "github_comment_links"
CREATE UNIQUE INDEX "github_comment_links_comment_id_key" ON "github_comment_links" ("comment_id");
-- create index "github_comment_links_github_comment_id_key" to table: "github_comment_links"
CREATE UNIQUE INDEX "github_comment_links_github_comment_id_key" ON "github_comment_links" ("github_comment_id");

-- +goose Down
-- reverse: create index "github_comment_links_github_comment_id_key" to table: "github_comment_links"
DROP INDEX "github_comment_links_github_comment_id_key";
-- reverse: create index "github_comment_links_comment_id_key" to table: "github_comment_links"
DROP INDEX "github_comment_links_comment_id_key";
-- reverse: create "github_comment_links" table
DROP TABLE "github_comment_links";
-- reverse: create index "githubissuelink_repo_sync_id_number" to table: "github_issue_links"
DROP INDEX "githubissuelink_repo_sync_id_number";
-- reverse: create index "github_issue_links_node_id_key" to table: "github_issue_links"
DROP INDEX "github_issue_links_node_id_key";
-- reverse: create index "github_issue_links_issue_id_key" to table: "github_issue_links"
DROP INDEX "github_issue_links_issue_id_key";
-- reverse: create "github_issue_links" table
DROP TABLE "github_issue_links";
-- reverse: create index "github_repo_syncs_full_name_key" to table: "github_repo_syncs"
DROP INDEX "github_repo_syncs_full_name_key";
-- reverse: create "github_repo_syncs" table
DROP TABLE "github_repo_syncs";
-- reverse: create index "users_github_login_key" to table: "users"
DROP INDEX "users_github_login_key";
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "github_login";
-- reverse: modify "github_integrations" table
ALTER TABLE "github_integrations" DROP COLUMN "token_prefix", DROP COLUMN "token_encrypted";
