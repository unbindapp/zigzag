-- +goose Up
-- create "github_integrations" table
CREATE TABLE "github_integrations" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "secret_encrypted" bytea NOT NULL,
  "secret_prefix" character varying NOT NULL,
  "last_delivery_at" timestamptz NULL,
  "created_by_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "github_integrations_users_github_integrations" FOREIGN KEY ("created_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL,
  CONSTRAINT "github_integrations_workspaces_github_integration" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "github_integrations_workspace_id_key" to table: "github_integrations"
CREATE UNIQUE INDEX "github_integrations_workspace_id_key" ON "github_integrations" ("workspace_id");
-- create "github_repositories" table
CREATE TABLE "github_repositories" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "full_name" character varying NOT NULL,
  "last_event_at" timestamptz NOT NULL,
  "last_event" character varying NOT NULL,
  "integration_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "github_repositories_github_integrations_repositories" FOREIGN KEY ("integration_id") REFERENCES "github_integrations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "githubrepository_integration_id_full_name" to table: "github_repositories"
CREATE UNIQUE INDEX "githubrepository_integration_id_full_name" ON "github_repositories" ("integration_id", "full_name");
-- modify "teams" table
ALTER TABLE "teams" ADD COLUMN "github_pr_opened_state_id" uuid NULL, ADD COLUMN "github_pr_merged_state_id" uuid NULL, ADD
CONSTRAINT "teams_workflow_states_github_pr_merged_for_teams" FOREIGN KEY ("github_pr_merged_state_id") REFERENCES "workflow_states" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "teams_workflow_states_github_pr_opened_for_teams" FOREIGN KEY ("github_pr_opened_state_id") REFERENCES "workflow_states" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "teams" table
ALTER TABLE "teams" DROP CONSTRAINT "teams_workflow_states_github_pr_opened_for_teams", DROP CONSTRAINT "teams_workflow_states_github_pr_merged_for_teams", DROP COLUMN "github_pr_merged_state_id", DROP COLUMN "github_pr_opened_state_id";
-- reverse: create index "githubrepository_integration_id_full_name" to table: "github_repositories"
DROP INDEX "githubrepository_integration_id_full_name";
-- reverse: create "github_repositories" table
DROP TABLE "github_repositories";
-- reverse: create index "github_integrations_workspace_id_key" to table: "github_integrations"
DROP INDEX "github_integrations_workspace_id_key";
-- reverse: create "github_integrations" table
DROP TABLE "github_integrations";
