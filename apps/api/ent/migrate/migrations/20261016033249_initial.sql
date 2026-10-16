-- +goose Up
-- create "api_keys" table
CREATE TABLE "api_keys" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "token_prefix" character varying NOT NULL,
  "token_hash" character varying NOT NULL,
  "role" character varying NOT NULL,
  "full_access" boolean NOT NULL DEFAULT false,
  "workspace_ids" jsonb NOT NULL,
  "privileges" jsonb NOT NULL,
  "expires_at" timestamptz NULL,
  "last_used_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "api_keys_token_hash_key" to table: "api_keys"
CREATE UNIQUE INDEX "api_keys_token_hash_key" ON "api_keys" ("token_hash");
-- create index "apikey_user_id_name" to table: "api_keys"
CREATE UNIQUE INDEX "apikey_user_id_name" ON "api_keys" ("user_id", "name");
-- create "attachments" table
CREATE TABLE "attachments" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "title" character varying NOT NULL,
  "subtitle" character varying NULL,
  "url" character varying NOT NULL,
  "source_type" character varying NULL,
  "metadata" jsonb NULL,
  "sort_order" double precision NOT NULL DEFAULT 0,
  "issue_id" uuid NOT NULL,
  "upload_id" uuid NULL,
  "creator_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- create index "attachment_issue_id" to table: "attachments"
CREATE INDEX "attachment_issue_id" ON "attachments" ("issue_id");
-- create "comments" table
CREATE TABLE "comments" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "body" character varying NOT NULL,
  "edited_at" timestamptz NULL,
  "resolved_at" timestamptz NULL,
  "parent_id" uuid NULL,
  "issue_id" uuid NOT NULL,
  "author_id" uuid NULL,
  "resolved_by_id" uuid NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "comments_comments_replies" FOREIGN KEY ("parent_id") REFERENCES "comments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "comment_issue_id_created_at" to table: "comments"
CREATE INDEX "comment_issue_id_created_at" ON "comments" ("issue_id", "created_at");
-- create "cycles" table
CREATE TABLE "cycles" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "number" bigint NOT NULL,
  "name" character varying NULL,
  "description" character varying NULL,
  "starts_at" timestamptz NOT NULL,
  "ends_at" timestamptz NOT NULL,
  "completed_at" timestamptz NULL,
  "archived_at" timestamptz NULL,
  "team_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "cycle_ends_at" to table: "cycles"
CREATE INDEX "cycle_ends_at" ON "cycles" ("ends_at");
-- create index "cycle_team_id_number" to table: "cycles"
CREATE UNIQUE INDEX "cycle_team_id_number" ON "cycles" ("team_id", "number");
-- create "documents" table
CREATE TABLE "documents" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "title" character varying NOT NULL,
  "slug" character varying NOT NULL,
  "icon" character varying NULL,
  "content" character varying NOT NULL DEFAULT '',
  "archived_at" timestamptz NULL,
  "trashed_at" timestamptz NULL,
  "initiative_id" uuid NULL,
  "issue_id" uuid NULL,
  "project_id" uuid NULL,
  "creator_id" uuid NULL,
  "updated_by_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "document_workspace_id_slug" to table: "documents"
CREATE UNIQUE INDEX "document_workspace_id_slug" ON "documents" ("workspace_id", "slug");
-- create "favorites" table
CREATE TABLE "favorites" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "kind" character varying NOT NULL,
  "entity_id" uuid NOT NULL,
  "sort_order" double precision NOT NULL DEFAULT 0,
  "user_id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "favorite_user_id_kind_entity_id" to table: "favorites"
CREATE UNIQUE INDEX "favorite_user_id_kind_entity_id" ON "favorites" ("user_id", "kind", "entity_id");
-- create index "favorite_user_id_workspace_id" to table: "favorites"
CREATE INDEX "favorite_user_id_workspace_id" ON "favorites" ("user_id", "workspace_id");
-- create "imports" table
CREATE TABLE "imports" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "source" character varying NOT NULL,
  "dry_run" boolean NOT NULL DEFAULT true,
  "status" character varying NOT NULL DEFAULT 'running',
  "report" jsonb NULL,
  "finished_at" timestamptz NULL,
  "started_by_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "importjob_workspace_id_created_at" to table: "imports"
CREATE INDEX "importjob_workspace_id_created_at" ON "imports" ("workspace_id", "created_at");
-- create "initiatives" table
CREATE TABLE "initiatives" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "slug" character varying NOT NULL,
  "description" character varying NULL,
  "icon" character varying NULL,
  "color" character varying NULL,
  "status" character varying NOT NULL DEFAULT 'planned',
  "target_date" date NULL,
  "sort_order" double precision NOT NULL DEFAULT 0,
  "archived_at" timestamptz NULL,
  "trashed_at" timestamptz NULL,
  "owner_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "initiative_workspace_id_slug" to table: "initiatives"
CREATE UNIQUE INDEX "initiative_workspace_id_slug" ON "initiatives" ("workspace_id", "slug");
-- create "initiative_updates" table
CREATE TABLE "initiative_updates" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "body" character varying NOT NULL,
  "health" character varying NOT NULL,
  "edited_at" timestamptz NULL,
  "initiative_id" uuid NOT NULL,
  "author_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- create "invitations" table
CREATE TABLE "invitations" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "email" character varying NOT NULL,
  "token_hash" character varying NOT NULL,
  "role" character varying NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "accepted_at" timestamptz NULL,
  "team_ids" jsonb NOT NULL,
  "invited_by" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "invitation_workspace_id_email" to table: "invitations"
CREATE UNIQUE INDEX "invitation_workspace_id_email" ON "invitations" ("workspace_id", "email");
-- create index "invitations_token_hash_key" to table: "invitations"
CREATE UNIQUE INDEX "invitations_token_hash_key" ON "invitations" ("token_hash");
-- create "issues" table
CREATE TABLE "issues" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "number" bigint NOT NULL,
  "title" character varying NOT NULL,
  "description" character varying NULL,
  "description_text" character varying NULL,
  "priority" bigint NOT NULL DEFAULT 0,
  "estimate" double precision NULL,
  "sort_order" double precision NOT NULL DEFAULT 0,
  "sub_issue_sort_order" double precision NOT NULL DEFAULT 0,
  "due_date" date NULL,
  "started_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "canceled_at" timestamptz NULL,
  "triaged_at" timestamptz NULL,
  "snoozed_until" timestamptz NULL,
  "archived_at" timestamptz NULL,
  "trashed_at" timestamptz NULL,
  "auto_closed_at" timestamptz NULL,
  "auto_archived_at" timestamptz NULL,
  "import_source" character varying NULL,
  "import_external_id" character varying NULL,
  "import_url" character varying NULL,
  "cycle_id" uuid NULL,
  "parent_id" uuid NULL,
  "project_id" uuid NULL,
  "milestone_id" uuid NULL,
  "team_id" uuid NOT NULL,
  "assignee_id" uuid NULL,
  "creator_id" uuid NULL,
  "state_id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "issues_issues_children" FOREIGN KEY ("parent_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- create index "issue_cycle_id" to table: "issues"
CREATE INDEX "issue_cycle_id" ON "issues" ("cycle_id");
-- create index "issue_parent_id" to table: "issues"
CREATE INDEX "issue_parent_id" ON "issues" ("parent_id");
-- create index "issue_project_id" to table: "issues"
CREATE INDEX "issue_project_id" ON "issues" ("project_id");
-- create index "issue_team_id_number" to table: "issues"
CREATE UNIQUE INDEX "issue_team_id_number" ON "issues" ("team_id", "number");
-- create index "issue_team_id_state_id" to table: "issues"
CREATE INDEX "issue_team_id_state_id" ON "issues" ("team_id", "state_id");
-- create index "issue_workspace_id_assignee_id" to table: "issues"
CREATE INDEX "issue_workspace_id_assignee_id" ON "issues" ("workspace_id", "assignee_id");
-- create index "issue_workspace_id_import_source_import_external_id" to table: "issues"
CREATE UNIQUE INDEX "issue_workspace_id_import_source_import_external_id" ON "issues" ("workspace_id", "import_source", "import_external_id") WHERE ((import_source IS NOT NULL) AND (import_external_id IS NOT NULL));
-- create index "issue_workspace_id_updated_at" to table: "issues"
CREATE INDEX "issue_workspace_id_updated_at" ON "issues" ("workspace_id", "updated_at");
-- create index "issues_description_text_trgm" to table: "issues"
CREATE INDEX "issues_description_text_trgm" ON "issues" USING gin ("description_text" gin_trgm_ops);
-- create index "issues_title_trgm" to table: "issues"
CREATE INDEX "issues_title_trgm" ON "issues" USING gin ("title" gin_trgm_ops);
-- create "issue_history" table
CREATE TABLE "issue_history" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "changes" jsonb NOT NULL,
  "source" character varying NOT NULL,
  "issue_id" uuid NOT NULL,
  "actor_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- create index "issuehistory_issue_id_created_at" to table: "issue_history"
CREATE INDEX "issuehistory_issue_id_created_at" ON "issue_history" ("issue_id", "created_at");
-- create "issue_moves" table
CREATE TABLE "issue_moves" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "old_number" bigint NOT NULL,
  "issue_id" uuid NOT NULL,
  "old_team_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "issuemove_old_team_id_old_number" to table: "issue_moves"
CREATE UNIQUE INDEX "issuemove_old_team_id_old_number" ON "issue_moves" ("old_team_id", "old_number");
-- create "issue_relations" table
CREATE TABLE "issue_relations" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "type" character varying NOT NULL,
  "issue_id" uuid NOT NULL,
  "related_issue_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "issuerelation_issue_id_related_issue_id_type" to table: "issue_relations"
CREATE UNIQUE INDEX "issuerelation_issue_id_related_issue_id_type" ON "issue_relations" ("issue_id", "related_issue_id", "type");
-- create index "issuerelation_related_issue_id" to table: "issue_relations"
CREATE INDEX "issuerelation_related_issue_id" ON "issue_relations" ("related_issue_id");
-- create "jwt_keys" table
CREATE TABLE "jwt_keys" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "label" character varying NOT NULL,
  "private_key" bytea NOT NULL,
  PRIMARY KEY ("id")
);
-- create "labels" table
CREATE TABLE "labels" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "color" character varying NOT NULL,
  "description" character varying NULL,
  "is_group" boolean NOT NULL DEFAULT false,
  "group_single_select" boolean NOT NULL DEFAULT false,
  "archived_at" timestamptz NULL,
  "parent_id" uuid NULL,
  "team_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "labels_labels_children" FOREIGN KEY ("parent_id") REFERENCES "labels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "label_parent_id" to table: "labels"
CREATE INDEX "label_parent_id" ON "labels" ("parent_id");
-- create index "label_workspace_id_team_id" to table: "labels"
CREATE INDEX "label_workspace_id_team_id" ON "labels" ("workspace_id", "team_id");
-- create "notifications" table
CREATE TABLE "notifications" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "type" character varying NOT NULL,
  "title" character varying NOT NULL,
  "body" character varying NULL,
  "read_at" timestamptz NULL,
  "archived_at" timestamptz NULL,
  "snoozed_until" timestamptz NULL,
  "emailed_at" timestamptz NULL,
  "comment_id" uuid NULL,
  "issue_id" uuid NULL,
  "project_id" uuid NULL,
  "project_update_id" uuid NULL,
  "user_id" uuid NOT NULL,
  "actor_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "notification_emailed_at" to table: "notifications"
CREATE INDEX "notification_emailed_at" ON "notifications" ("emailed_at");
-- create index "notification_user_id_archived_at_read_at" to table: "notifications"
CREATE INDEX "notification_user_id_archived_at_read_at" ON "notifications" ("user_id", "archived_at", "read_at");
-- create "notification_settings" table
CREATE TABLE "notification_settings" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "email_enabled" boolean NOT NULL DEFAULT true,
  "assigned" boolean NOT NULL DEFAULT true,
  "mentioned" boolean NOT NULL DEFAULT true,
  "commented" boolean NOT NULL DEFAULT true,
  "state_changed" boolean NOT NULL DEFAULT true,
  "due_soon" boolean NOT NULL DEFAULT true,
  "project_updates" boolean NOT NULL DEFAULT true,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "notification_settings_user_id_key" to table: "notification_settings"
CREATE UNIQUE INDEX "notification_settings_user_id_key" ON "notification_settings" ("user_id");
-- create "oauth_authorization_codes" table
CREATE TABLE "oauth_authorization_codes" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "code_hash" character varying NOT NULL,
  "client_id" character varying NOT NULL,
  "client_name" character varying NOT NULL,
  "client_kind" character varying NOT NULL,
  "client_uri" character varying NULL,
  "redirect_uri" character varying NOT NULL,
  "code_challenge" character varying NOT NULL,
  "resource" character varying NOT NULL,
  "scope" character varying NULL,
  "role" character varying NOT NULL,
  "full_access" boolean NOT NULL DEFAULT false,
  "workspace_ids" jsonb NOT NULL,
  "privileges" jsonb NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz NULL,
  "grant_id" uuid NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "oauth_authorization_codes_code_hash_key" to table: "oauth_authorization_codes"
CREATE UNIQUE INDEX "oauth_authorization_codes_code_hash_key" ON "oauth_authorization_codes" ("code_hash");
-- create index "oauthauthorizationcode_expires_at" to table: "oauth_authorization_codes"
CREATE INDEX "oauthauthorizationcode_expires_at" ON "oauth_authorization_codes" ("expires_at");
-- create "oauth_clients" table
CREATE TABLE "oauth_clients" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "client_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "redirect_uris" jsonb NOT NULL,
  "client_uri" character varying NULL,
  "last_used_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "oauth_clients_client_id_key" to table: "oauth_clients"
CREATE UNIQUE INDEX "oauth_clients_client_id_key" ON "oauth_clients" ("client_id");
-- create index "oauthclient_created_at" to table: "oauth_clients"
CREATE INDEX "oauthclient_created_at" ON "oauth_clients" ("created_at");
-- create "oauth_grants" table
CREATE TABLE "oauth_grants" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "client_id" character varying NOT NULL,
  "client_name" character varying NOT NULL,
  "client_kind" character varying NOT NULL,
  "client_uri" character varying NULL,
  "redirect_uri" character varying NOT NULL,
  "role" character varying NOT NULL,
  "full_access" boolean NOT NULL DEFAULT false,
  "workspace_ids" jsonb NOT NULL,
  "privileges" jsonb NOT NULL,
  "resource" character varying NOT NULL,
  "scope" character varying NULL,
  "last_used_at" timestamptz NULL,
  "revoked_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "oauthgrant_client_id" to table: "oauth_grants"
CREATE INDEX "oauthgrant_client_id" ON "oauth_grants" ("client_id");
-- create index "oauthgrant_user_id" to table: "oauth_grants"
CREATE INDEX "oauthgrant_user_id" ON "oauth_grants" ("user_id");
-- create "oauth_grant_tokens" table
CREATE TABLE "oauth_grant_tokens" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "kind" character varying NOT NULL,
  "token_hash" character varying NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz NULL,
  "grant_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "oauth_grant_tokens_token_hash_key" to table: "oauth_grant_tokens"
CREATE UNIQUE INDEX "oauth_grant_tokens_token_hash_key" ON "oauth_grant_tokens" ("token_hash");
-- create index "oauthgranttoken_expires_at" to table: "oauth_grant_tokens"
CREATE INDEX "oauthgranttoken_expires_at" ON "oauth_grant_tokens" ("expires_at");
-- create index "oauthgranttoken_grant_id" to table: "oauth_grant_tokens"
CREATE INDEX "oauthgranttoken_grant_id" ON "oauth_grant_tokens" ("grant_id");
-- create "oauth2_tokens" table
CREATE TABLE "oauth2_tokens" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "access_token" character varying NOT NULL,
  "refresh_token" character varying NOT NULL,
  "client_id" character varying NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "revoked" boolean NOT NULL DEFAULT false,
  "scope" character varying NOT NULL,
  "device_info" character varying NULL,
  "user_oauth2_tokens" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "oauth2_tokens_access_token_key" to table: "oauth2_tokens"
CREATE UNIQUE INDEX "oauth2_tokens_access_token_key" ON "oauth2_tokens" ("access_token");
-- create index "oauth2_tokens_refresh_token_key" to table: "oauth2_tokens"
CREATE UNIQUE INDEX "oauth2_tokens_refresh_token_key" ON "oauth2_tokens" ("refresh_token");
-- create "password_reset_tokens" table
CREATE TABLE "password_reset_tokens" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "token_hash" character varying NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "used_at" timestamptz NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "password_reset_tokens_token_hash_key" to table: "password_reset_tokens"
CREATE UNIQUE INDEX "password_reset_tokens_token_hash_key" ON "password_reset_tokens" ("token_hash");
-- create index "passwordresettoken_expires_at" to table: "password_reset_tokens"
CREATE INDEX "passwordresettoken_expires_at" ON "password_reset_tokens" ("expires_at");
-- create "projects" table
CREATE TABLE "projects" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "slug" character varying NOT NULL,
  "summary" character varying NULL,
  "description" character varying NULL,
  "icon" character varying NULL,
  "color" character varying NULL,
  "priority" bigint NOT NULL DEFAULT 0,
  "start_date" date NULL,
  "target_date" date NULL,
  "start_date_resolution" character varying NOT NULL DEFAULT 'day',
  "target_date_resolution" character varying NOT NULL DEFAULT 'day',
  "sort_order" double precision NOT NULL DEFAULT 0,
  "started_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "canceled_at" timestamptz NULL,
  "archived_at" timestamptz NULL,
  "trashed_at" timestamptz NULL,
  "health" character varying NULL,
  "status_id" uuid NOT NULL,
  "lead_id" uuid NULL,
  "creator_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "project_status_id" to table: "projects"
CREATE INDEX "project_status_id" ON "projects" ("status_id");
-- create index "project_workspace_id_slug" to table: "projects"
CREATE UNIQUE INDEX "project_workspace_id_slug" ON "projects" ("workspace_id", "slug");
-- create "project_links" table
CREATE TABLE "project_links" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "url" character varying NOT NULL,
  "label" character varying NOT NULL,
  "sort_order" double precision NOT NULL DEFAULT 0,
  "project_id" uuid NOT NULL,
  "creator_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- create "project_milestones" table
CREATE TABLE "project_milestones" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "target_date" date NULL,
  "sort_order" double precision NOT NULL DEFAULT 0,
  "archived_at" timestamptz NULL,
  "project_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create "project_statuses" table
CREATE TABLE "project_statuses" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "type" character varying NOT NULL,
  "color" character varying NOT NULL,
  "position" double precision NOT NULL DEFAULT 0,
  "description" character varying NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "projectstatus_workspace_id" to table: "project_statuses"
CREATE INDEX "projectstatus_workspace_id" ON "project_statuses" ("workspace_id");
-- create "project_updates" table
CREATE TABLE "project_updates" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "body" character varying NOT NULL,
  "health" character varying NOT NULL,
  "edited_at" timestamptz NULL,
  "project_id" uuid NOT NULL,
  "author_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- create "reactions" table
CREATE TABLE "reactions" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "emoji" character varying NOT NULL,
  "comment_id" uuid NULL,
  "issue_id" uuid NULL,
  "project_update_id" uuid NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "reaction_user_id_emoji_comment_id" to table: "reactions"
CREATE UNIQUE INDEX "reaction_user_id_emoji_comment_id" ON "reactions" ("user_id", "emoji", "comment_id");
-- create index "reaction_user_id_emoji_issue_id" to table: "reactions"
CREATE UNIQUE INDEX "reaction_user_id_emoji_issue_id" ON "reactions" ("user_id", "emoji", "issue_id");
-- create index "reaction_user_id_emoji_project_update_id" to table: "reactions"
CREATE UNIQUE INDEX "reaction_user_id_emoji_project_update_id" ON "reactions" ("user_id", "emoji", "project_update_id");
-- create "system_settings" table
CREATE TABLE "system_settings" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "instance_name" character varying NOT NULL DEFAULT 'Zigzag',
  "email_settings" jsonb NULL,
  "email_password_encrypted" bytea NULL,
  "allow_public_signup" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- create "teams" table
CREATE TABLE "teams" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "key" character varying NOT NULL,
  "description" character varying NULL,
  "icon" character varying NULL,
  "color" character varying NULL,
  "issue_counter" bigint NOT NULL DEFAULT 0,
  "timezone" character varying NOT NULL DEFAULT 'UTC',
  "private" boolean NOT NULL DEFAULT false,
  "cycles_enabled" boolean NOT NULL DEFAULT false,
  "cycle_duration_weeks" bigint NOT NULL DEFAULT 2,
  "cycle_cooldown_weeks" bigint NOT NULL DEFAULT 0,
  "cycle_start_day" bigint NOT NULL DEFAULT 1,
  "cycle_auto_move" boolean NOT NULL DEFAULT true,
  "cycle_auto_add_started" boolean NOT NULL DEFAULT true,
  "estimate_type" character varying NOT NULL DEFAULT 'none',
  "estimate_allow_zero" boolean NOT NULL DEFAULT false,
  "estimate_extended" boolean NOT NULL DEFAULT false,
  "triage_enabled" boolean NOT NULL DEFAULT false,
  "auto_close_months" bigint NULL,
  "auto_archive_months" bigint NOT NULL DEFAULT 6,
  "archived_at" timestamptz NULL,
  "default_state_id" uuid NULL,
  "triage_state_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "team_workspace_id_key" to table: "teams"
CREATE UNIQUE INDEX "team_workspace_id_key" ON "teams" ("workspace_id", "key");
-- create index "team_workspace_id_name" to table: "teams"
CREATE UNIQUE INDEX "team_workspace_id_name" ON "teams" ("workspace_id", "name");
-- create "team_members" table
CREATE TABLE "team_members" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "owner" boolean NOT NULL DEFAULT false,
  "team_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "teammember_team_id_user_id" to table: "team_members"
CREATE UNIQUE INDEX "teammember_team_id_user_id" ON "team_members" ("team_id", "user_id");
-- create index "teammember_user_id" to table: "team_members"
CREATE INDEX "teammember_user_id" ON "team_members" ("user_id");
-- create "templates" table
CREATE TABLE "templates" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "kind" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "data" jsonb NOT NULL,
  "sort_order" double precision NOT NULL DEFAULT 0,
  "team_id" uuid NULL,
  "creator_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "template_workspace_id_kind" to table: "templates"
CREATE INDEX "template_workspace_id_kind" ON "templates" ("workspace_id", "kind");
-- create "uploads" table
CREATE TABLE "uploads" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "file_name" character varying NOT NULL,
  "content_type" character varying NOT NULL,
  "size_bytes" bigint NOT NULL,
  "storage_key" character varying NOT NULL,
  "width" bigint NULL,
  "height" bigint NULL,
  "uploader_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "upload_workspace_id" to table: "uploads"
CREATE INDEX "upload_workspace_id" ON "uploads" ("workspace_id");
-- create index "uploads_storage_key_key" to table: "uploads"
CREATE UNIQUE INDEX "uploads_storage_key_key" ON "uploads" ("storage_key");
-- create "users" table
CREATE TABLE "users" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "email" character varying NOT NULL,
  "password_hash" character varying NOT NULL DEFAULT '',
  "name" character varying NOT NULL DEFAULT '',
  "display_name" character varying NOT NULL,
  "system_admin" boolean NOT NULL DEFAULT false,
  "active" boolean NOT NULL DEFAULT true,
  "timezone" character varying NOT NULL DEFAULT 'UTC',
  "last_seen_at" timestamptz NULL,
  "avatar_upload_id" uuid NULL,
  PRIMARY KEY ("id")
);
-- create index "users_display_name_key" to table: "users"
CREATE UNIQUE INDEX "users_display_name_key" ON "users" ("display_name");
-- create index "users_email_key" to table: "users"
CREATE UNIQUE INDEX "users_email_key" ON "users" ("email");
-- create "views" table
CREATE TABLE "views" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "icon" character varying NULL,
  "color" character varying NULL,
  "kind" character varying NOT NULL,
  "shared" boolean NOT NULL DEFAULT false,
  "filters" jsonb NOT NULL,
  "display" jsonb NOT NULL,
  "team_id" uuid NULL,
  "owner_id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "view_workspace_id_owner_id" to table: "views"
CREATE INDEX "view_workspace_id_owner_id" ON "views" ("workspace_id", "owner_id");
-- create "webhooks" table
CREATE TABLE "webhooks" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "url" character varying NOT NULL,
  "secret" character varying NOT NULL,
  "enabled" boolean NOT NULL DEFAULT true,
  "events" jsonb NOT NULL,
  "label" character varying NOT NULL DEFAULT '',
  "team_id" uuid NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "webhook_workspace_id" to table: "webhooks"
CREATE INDEX "webhook_workspace_id" ON "webhooks" ("workspace_id");
-- create "workflow_states" table
CREATE TABLE "workflow_states" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "type" character varying NOT NULL,
  "color" character varying NOT NULL,
  "description" character varying NULL,
  "position" double precision NOT NULL DEFAULT 0,
  "archived_at" timestamptz NULL,
  "team_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "workflowstate_team_id_name" to table: "workflow_states"
CREATE UNIQUE INDEX "workflowstate_team_id_name" ON "workflow_states" ("team_id", "name");
-- create "workspaces" table
CREATE TABLE "workspaces" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "slug" character varying NOT NULL,
  "icon" character varying NULL,
  "color" character varying NULL,
  "archived_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- create index "workspaces_slug_key" to table: "workspaces"
CREATE UNIQUE INDEX "workspaces_slug_key" ON "workspaces" ("slug");
-- create "workspace_members" table
CREATE TABLE "workspace_members" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "role" character varying NOT NULL,
  "joined_at" timestamptz NOT NULL,
  "user_id" uuid NOT NULL,
  "workspace_id" uuid NOT NULL,
  PRIMARY KEY ("id")
);
-- create index "workspacemember_user_id" to table: "workspace_members"
CREATE INDEX "workspacemember_user_id" ON "workspace_members" ("user_id");
-- create index "workspacemember_workspace_id_user_id" to table: "workspace_members"
CREATE UNIQUE INDEX "workspacemember_workspace_id_user_id" ON "workspace_members" ("workspace_id", "user_id");
-- create "initiative_projects" table
CREATE TABLE "initiative_projects" (
  "initiative_id" uuid NOT NULL,
  "project_id" uuid NOT NULL,
  PRIMARY KEY ("initiative_id", "project_id")
);
-- create "issue_labels" table
CREATE TABLE "issue_labels" (
  "issue_id" uuid NOT NULL,
  "label_id" uuid NOT NULL,
  PRIMARY KEY ("issue_id", "label_id")
);
-- create "issue_subscriptions" table
CREATE TABLE "issue_subscriptions" (
  "issue_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("issue_id", "user_id")
);
-- create "project_teams" table
CREATE TABLE "project_teams" (
  "project_id" uuid NOT NULL,
  "team_id" uuid NOT NULL,
  PRIMARY KEY ("project_id", "team_id")
);
-- create "project_members" table
CREATE TABLE "project_members" (
  "project_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  PRIMARY KEY ("project_id", "user_id")
);
-- create "project_labels" table
CREATE TABLE "project_labels" (
  "project_id" uuid NOT NULL,
  "label_id" uuid NOT NULL,
  PRIMARY KEY ("project_id", "label_id")
);
-- modify "api_keys" table
ALTER TABLE "api_keys" ADD
CONSTRAINT "api_keys_users_api_keys" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "attachments" table
ALTER TABLE "attachments" ADD
CONSTRAINT "attachments_issues_attachments" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "attachments_uploads_attachments" FOREIGN KEY ("upload_id") REFERENCES "uploads" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "attachments_users_attachments" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "comments" table
ALTER TABLE "comments" ADD
CONSTRAINT "comments_issues_comments" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "comments_users_comments" FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "comments_users_resolved_comments" FOREIGN KEY ("resolved_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "cycles" table
ALTER TABLE "cycles" ADD
CONSTRAINT "cycles_teams_cycles" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "documents" table
ALTER TABLE "documents" ADD
CONSTRAINT "documents_initiatives_documents" FOREIGN KEY ("initiative_id") REFERENCES "initiatives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "documents_issues_documents" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "documents_projects_documents" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "documents_users_created_documents" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "documents_users_updated_documents" FOREIGN KEY ("updated_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "documents_workspaces_documents" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "favorites" table
ALTER TABLE "favorites" ADD
CONSTRAINT "favorites_users_favorites" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "favorites_workspaces_favorites" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "imports" table
ALTER TABLE "imports" ADD
CONSTRAINT "imports_users_imports" FOREIGN KEY ("started_by_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "imports_workspaces_imports" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "initiatives" table
ALTER TABLE "initiatives" ADD
CONSTRAINT "initiatives_users_owned_initiatives" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "initiatives_workspaces_initiatives" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "initiative_updates" table
ALTER TABLE "initiative_updates" ADD
CONSTRAINT "initiative_updates_initiatives_updates" FOREIGN KEY ("initiative_id") REFERENCES "initiatives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "initiative_updates_users_initiative_updates" FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "invitations" table
ALTER TABLE "invitations" ADD
CONSTRAINT "invitations_users_sent_invitations" FOREIGN KEY ("invited_by") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "invitations_workspaces_invitations" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "issues" table
ALTER TABLE "issues" ADD
CONSTRAINT "issues_cycles_issues" FOREIGN KEY ("cycle_id") REFERENCES "cycles" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "issues_project_milestones_issues" FOREIGN KEY ("milestone_id") REFERENCES "project_milestones" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "issues_projects_issues" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "issues_teams_issues" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "issues_users_assigned_issues" FOREIGN KEY ("assignee_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "issues_users_created_issues" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "issues_workflow_states_issues" FOREIGN KEY ("state_id") REFERENCES "workflow_states" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
CONSTRAINT "issues_workspaces_issues" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "issue_history" table
ALTER TABLE "issue_history" ADD
CONSTRAINT "issue_history_issues_history" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "issue_history_users_issue_history" FOREIGN KEY ("actor_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "issue_moves" table
ALTER TABLE "issue_moves" ADD
CONSTRAINT "issue_moves_issues_moves" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "issue_moves_teams_issue_moves" FOREIGN KEY ("old_team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "issue_relations" table
ALTER TABLE "issue_relations" ADD
CONSTRAINT "issue_relations_issues_inverse_relations" FOREIGN KEY ("related_issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "issue_relations_issues_relations" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "labels" table
ALTER TABLE "labels" ADD
CONSTRAINT "labels_teams_labels" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "labels_workspaces_labels" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "notifications" table
ALTER TABLE "notifications" ADD
CONSTRAINT "notifications_comments_notifications" FOREIGN KEY ("comment_id") REFERENCES "comments" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "notifications_issues_notifications" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "notifications_project_updates_notifications" FOREIGN KEY ("project_update_id") REFERENCES "project_updates" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "notifications_projects_notifications" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "notifications_users_acted_notifications" FOREIGN KEY ("actor_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "notifications_users_notifications" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "notifications_workspaces_notifications" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "notification_settings" table
ALTER TABLE "notification_settings" ADD
CONSTRAINT "notification_settings_users_notification_setting" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "oauth_authorization_codes" table
ALTER TABLE "oauth_authorization_codes" ADD
CONSTRAINT "oauth_authorization_codes_users_oauth_authorization_codes" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "oauth_grants" table
ALTER TABLE "oauth_grants" ADD
CONSTRAINT "oauth_grants_users_oauth_grants" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "oauth_grant_tokens" table
ALTER TABLE "oauth_grant_tokens" ADD
CONSTRAINT "oauth_grant_tokens_oauth_grants_tokens" FOREIGN KEY ("grant_id") REFERENCES "oauth_grants" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "oauth2_tokens" table
ALTER TABLE "oauth2_tokens" ADD
CONSTRAINT "oauth2_tokens_users_oauth2_tokens" FOREIGN KEY ("user_oauth2_tokens") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" ADD
CONSTRAINT "password_reset_tokens_users_password_reset_tokens" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "projects" table
ALTER TABLE "projects" ADD
CONSTRAINT "projects_project_statuses_projects" FOREIGN KEY ("status_id") REFERENCES "project_statuses" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION, ADD
CONSTRAINT "projects_users_created_projects" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "projects_users_led_projects" FOREIGN KEY ("lead_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "projects_workspaces_projects" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "project_links" table
ALTER TABLE "project_links" ADD
CONSTRAINT "project_links_projects_links" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "project_links_users_project_links" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "project_milestones" table
ALTER TABLE "project_milestones" ADD
CONSTRAINT "project_milestones_projects_milestones" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "project_statuses" table
ALTER TABLE "project_statuses" ADD
CONSTRAINT "project_statuses_workspaces_project_statuses" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "project_updates" table
ALTER TABLE "project_updates" ADD
CONSTRAINT "project_updates_projects_updates" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "project_updates_users_project_updates" FOREIGN KEY ("author_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "reactions" table
ALTER TABLE "reactions" ADD
CONSTRAINT "reactions_comments_reactions" FOREIGN KEY ("comment_id") REFERENCES "comments" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "reactions_issues_reactions" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "reactions_project_updates_reactions" FOREIGN KEY ("project_update_id") REFERENCES "project_updates" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "reactions_users_reactions" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "teams" table
ALTER TABLE "teams" ADD
CONSTRAINT "teams_workflow_states_default_for_teams" FOREIGN KEY ("default_state_id") REFERENCES "workflow_states" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "teams_workflow_states_triage_for_teams" FOREIGN KEY ("triage_state_id") REFERENCES "workflow_states" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "teams_workspaces_teams" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "team_members" table
ALTER TABLE "team_members" ADD
CONSTRAINT "team_members_teams_members" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "team_members_users_team_memberships" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "templates" table
ALTER TABLE "templates" ADD
CONSTRAINT "templates_teams_templates" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "templates_users_templates" FOREIGN KEY ("creator_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "templates_workspaces_templates" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "uploads" table
ALTER TABLE "uploads" ADD
CONSTRAINT "uploads_users_uploads" FOREIGN KEY ("uploader_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE SET NULL, ADD
CONSTRAINT "uploads_workspaces_uploads" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "users" table
ALTER TABLE "users" ADD
CONSTRAINT "users_uploads_avatar_users" FOREIGN KEY ("avatar_upload_id") REFERENCES "uploads" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;
-- modify "views" table
ALTER TABLE "views" ADD
CONSTRAINT "views_teams_views" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "views_users_views" FOREIGN KEY ("owner_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "views_workspaces_views" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "webhooks" table
ALTER TABLE "webhooks" ADD
CONSTRAINT "webhooks_teams_webhooks" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "webhooks_workspaces_webhooks" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "workflow_states" table
ALTER TABLE "workflow_states" ADD
CONSTRAINT "workflow_states_teams_states" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "workspace_members" table
ALTER TABLE "workspace_members" ADD
CONSTRAINT "workspace_members_users_workspace_memberships" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "workspace_members_workspaces_members" FOREIGN KEY ("workspace_id") REFERENCES "workspaces" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "initiative_projects" table
ALTER TABLE "initiative_projects" ADD
CONSTRAINT "initiative_projects_initiative_id" FOREIGN KEY ("initiative_id") REFERENCES "initiatives" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "initiative_projects_project_id" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "issue_labels" table
ALTER TABLE "issue_labels" ADD
CONSTRAINT "issue_labels_issue_id" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "issue_labels_label_id" FOREIGN KEY ("label_id") REFERENCES "labels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "issue_subscriptions" table
ALTER TABLE "issue_subscriptions" ADD
CONSTRAINT "issue_subscriptions_issue_id" FOREIGN KEY ("issue_id") REFERENCES "issues" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "issue_subscriptions_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "project_teams" table
ALTER TABLE "project_teams" ADD
CONSTRAINT "project_teams_project_id" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "project_teams_team_id" FOREIGN KEY ("team_id") REFERENCES "teams" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "project_members" table
ALTER TABLE "project_members" ADD
CONSTRAINT "project_members_project_id" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "project_members_user_id" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;
-- modify "project_labels" table
ALTER TABLE "project_labels" ADD
CONSTRAINT "project_labels_label_id" FOREIGN KEY ("label_id") REFERENCES "labels" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, ADD
CONSTRAINT "project_labels_project_id" FOREIGN KEY ("project_id") REFERENCES "projects" ("id") ON UPDATE NO ACTION ON DELETE CASCADE;

-- +goose Down
-- reverse: modify "project_labels" table
ALTER TABLE "project_labels" DROP CONSTRAINT "project_labels_project_id", DROP CONSTRAINT "project_labels_label_id";
-- reverse: modify "project_members" table
ALTER TABLE "project_members" DROP CONSTRAINT "project_members_user_id", DROP CONSTRAINT "project_members_project_id";
-- reverse: modify "project_teams" table
ALTER TABLE "project_teams" DROP CONSTRAINT "project_teams_team_id", DROP CONSTRAINT "project_teams_project_id";
-- reverse: modify "issue_subscriptions" table
ALTER TABLE "issue_subscriptions" DROP CONSTRAINT "issue_subscriptions_user_id", DROP CONSTRAINT "issue_subscriptions_issue_id";
-- reverse: modify "issue_labels" table
ALTER TABLE "issue_labels" DROP CONSTRAINT "issue_labels_label_id", DROP CONSTRAINT "issue_labels_issue_id";
-- reverse: modify "initiative_projects" table
ALTER TABLE "initiative_projects" DROP CONSTRAINT "initiative_projects_project_id", DROP CONSTRAINT "initiative_projects_initiative_id";
-- reverse: modify "workspace_members" table
ALTER TABLE "workspace_members" DROP CONSTRAINT "workspace_members_workspaces_members", DROP CONSTRAINT "workspace_members_users_workspace_memberships";
-- reverse: modify "workflow_states" table
ALTER TABLE "workflow_states" DROP CONSTRAINT "workflow_states_teams_states";
-- reverse: modify "webhooks" table
ALTER TABLE "webhooks" DROP CONSTRAINT "webhooks_workspaces_webhooks", DROP CONSTRAINT "webhooks_teams_webhooks";
-- reverse: modify "views" table
ALTER TABLE "views" DROP CONSTRAINT "views_workspaces_views", DROP CONSTRAINT "views_users_views", DROP CONSTRAINT "views_teams_views";
-- reverse: modify "users" table
ALTER TABLE "users" DROP CONSTRAINT "users_uploads_avatar_users";
-- reverse: modify "uploads" table
ALTER TABLE "uploads" DROP CONSTRAINT "uploads_workspaces_uploads", DROP CONSTRAINT "uploads_users_uploads";
-- reverse: modify "templates" table
ALTER TABLE "templates" DROP CONSTRAINT "templates_workspaces_templates", DROP CONSTRAINT "templates_users_templates", DROP CONSTRAINT "templates_teams_templates";
-- reverse: modify "team_members" table
ALTER TABLE "team_members" DROP CONSTRAINT "team_members_users_team_memberships", DROP CONSTRAINT "team_members_teams_members";
-- reverse: modify "teams" table
ALTER TABLE "teams" DROP CONSTRAINT "teams_workspaces_teams", DROP CONSTRAINT "teams_workflow_states_triage_for_teams", DROP CONSTRAINT "teams_workflow_states_default_for_teams";
-- reverse: modify "reactions" table
ALTER TABLE "reactions" DROP CONSTRAINT "reactions_users_reactions", DROP CONSTRAINT "reactions_project_updates_reactions", DROP CONSTRAINT "reactions_issues_reactions", DROP CONSTRAINT "reactions_comments_reactions";
-- reverse: modify "project_updates" table
ALTER TABLE "project_updates" DROP CONSTRAINT "project_updates_users_project_updates", DROP CONSTRAINT "project_updates_projects_updates";
-- reverse: modify "project_statuses" table
ALTER TABLE "project_statuses" DROP CONSTRAINT "project_statuses_workspaces_project_statuses";
-- reverse: modify "project_milestones" table
ALTER TABLE "project_milestones" DROP CONSTRAINT "project_milestones_projects_milestones";
-- reverse: modify "project_links" table
ALTER TABLE "project_links" DROP CONSTRAINT "project_links_users_project_links", DROP CONSTRAINT "project_links_projects_links";
-- reverse: modify "projects" table
ALTER TABLE "projects" DROP CONSTRAINT "projects_workspaces_projects", DROP CONSTRAINT "projects_users_led_projects", DROP CONSTRAINT "projects_users_created_projects", DROP CONSTRAINT "projects_project_statuses_projects";
-- reverse: modify "password_reset_tokens" table
ALTER TABLE "password_reset_tokens" DROP CONSTRAINT "password_reset_tokens_users_password_reset_tokens";
-- reverse: modify "oauth2_tokens" table
ALTER TABLE "oauth2_tokens" DROP CONSTRAINT "oauth2_tokens_users_oauth2_tokens";
-- reverse: modify "oauth_grant_tokens" table
ALTER TABLE "oauth_grant_tokens" DROP CONSTRAINT "oauth_grant_tokens_oauth_grants_tokens";
-- reverse: modify "oauth_grants" table
ALTER TABLE "oauth_grants" DROP CONSTRAINT "oauth_grants_users_oauth_grants";
-- reverse: modify "oauth_authorization_codes" table
ALTER TABLE "oauth_authorization_codes" DROP CONSTRAINT "oauth_authorization_codes_users_oauth_authorization_codes";
-- reverse: modify "notification_settings" table
ALTER TABLE "notification_settings" DROP CONSTRAINT "notification_settings_users_notification_setting";
-- reverse: modify "notifications" table
ALTER TABLE "notifications" DROP CONSTRAINT "notifications_workspaces_notifications", DROP CONSTRAINT "notifications_users_notifications", DROP CONSTRAINT "notifications_users_acted_notifications", DROP CONSTRAINT "notifications_projects_notifications", DROP CONSTRAINT "notifications_project_updates_notifications", DROP CONSTRAINT "notifications_issues_notifications", DROP CONSTRAINT "notifications_comments_notifications";
-- reverse: modify "labels" table
ALTER TABLE "labels" DROP CONSTRAINT "labels_workspaces_labels", DROP CONSTRAINT "labels_teams_labels";
-- reverse: modify "issue_relations" table
ALTER TABLE "issue_relations" DROP CONSTRAINT "issue_relations_issues_relations", DROP CONSTRAINT "issue_relations_issues_inverse_relations";
-- reverse: modify "issue_moves" table
ALTER TABLE "issue_moves" DROP CONSTRAINT "issue_moves_teams_issue_moves", DROP CONSTRAINT "issue_moves_issues_moves";
-- reverse: modify "issue_history" table
ALTER TABLE "issue_history" DROP CONSTRAINT "issue_history_users_issue_history", DROP CONSTRAINT "issue_history_issues_history";
-- reverse: modify "issues" table
ALTER TABLE "issues" DROP CONSTRAINT "issues_workspaces_issues", DROP CONSTRAINT "issues_workflow_states_issues", DROP CONSTRAINT "issues_users_created_issues", DROP CONSTRAINT "issues_users_assigned_issues", DROP CONSTRAINT "issues_teams_issues", DROP CONSTRAINT "issues_projects_issues", DROP CONSTRAINT "issues_project_milestones_issues", DROP CONSTRAINT "issues_cycles_issues";
-- reverse: modify "invitations" table
ALTER TABLE "invitations" DROP CONSTRAINT "invitations_workspaces_invitations", DROP CONSTRAINT "invitations_users_sent_invitations";
-- reverse: modify "initiative_updates" table
ALTER TABLE "initiative_updates" DROP CONSTRAINT "initiative_updates_users_initiative_updates", DROP CONSTRAINT "initiative_updates_initiatives_updates";
-- reverse: modify "initiatives" table
ALTER TABLE "initiatives" DROP CONSTRAINT "initiatives_workspaces_initiatives", DROP CONSTRAINT "initiatives_users_owned_initiatives";
-- reverse: modify "imports" table
ALTER TABLE "imports" DROP CONSTRAINT "imports_workspaces_imports", DROP CONSTRAINT "imports_users_imports";
-- reverse: modify "favorites" table
ALTER TABLE "favorites" DROP CONSTRAINT "favorites_workspaces_favorites", DROP CONSTRAINT "favorites_users_favorites";
-- reverse: modify "documents" table
ALTER TABLE "documents" DROP CONSTRAINT "documents_workspaces_documents", DROP CONSTRAINT "documents_users_updated_documents", DROP CONSTRAINT "documents_users_created_documents", DROP CONSTRAINT "documents_projects_documents", DROP CONSTRAINT "documents_issues_documents", DROP CONSTRAINT "documents_initiatives_documents";
-- reverse: modify "cycles" table
ALTER TABLE "cycles" DROP CONSTRAINT "cycles_teams_cycles";
-- reverse: modify "comments" table
ALTER TABLE "comments" DROP CONSTRAINT "comments_users_resolved_comments", DROP CONSTRAINT "comments_users_comments", DROP CONSTRAINT "comments_issues_comments";
-- reverse: modify "attachments" table
ALTER TABLE "attachments" DROP CONSTRAINT "attachments_users_attachments", DROP CONSTRAINT "attachments_uploads_attachments", DROP CONSTRAINT "attachments_issues_attachments";
-- reverse: modify "api_keys" table
ALTER TABLE "api_keys" DROP CONSTRAINT "api_keys_users_api_keys";
-- reverse: create "project_labels" table
DROP TABLE "project_labels";
-- reverse: create "project_members" table
DROP TABLE "project_members";
-- reverse: create "project_teams" table
DROP TABLE "project_teams";
-- reverse: create "issue_subscriptions" table
DROP TABLE "issue_subscriptions";
-- reverse: create "issue_labels" table
DROP TABLE "issue_labels";
-- reverse: create "initiative_projects" table
DROP TABLE "initiative_projects";
-- reverse: create index "workspacemember_workspace_id_user_id" to table: "workspace_members"
DROP INDEX "workspacemember_workspace_id_user_id";
-- reverse: create index "workspacemember_user_id" to table: "workspace_members"
DROP INDEX "workspacemember_user_id";
-- reverse: create "workspace_members" table
DROP TABLE "workspace_members";
-- reverse: create index "workspaces_slug_key" to table: "workspaces"
DROP INDEX "workspaces_slug_key";
-- reverse: create "workspaces" table
DROP TABLE "workspaces";
-- reverse: create index "workflowstate_team_id_name" to table: "workflow_states"
DROP INDEX "workflowstate_team_id_name";
-- reverse: create "workflow_states" table
DROP TABLE "workflow_states";
-- reverse: create index "webhook_workspace_id" to table: "webhooks"
DROP INDEX "webhook_workspace_id";
-- reverse: create "webhooks" table
DROP TABLE "webhooks";
-- reverse: create index "view_workspace_id_owner_id" to table: "views"
DROP INDEX "view_workspace_id_owner_id";
-- reverse: create "views" table
DROP TABLE "views";
-- reverse: create index "users_email_key" to table: "users"
DROP INDEX "users_email_key";
-- reverse: create index "users_display_name_key" to table: "users"
DROP INDEX "users_display_name_key";
-- reverse: create "users" table
DROP TABLE "users";
-- reverse: create index "uploads_storage_key_key" to table: "uploads"
DROP INDEX "uploads_storage_key_key";
-- reverse: create index "upload_workspace_id" to table: "uploads"
DROP INDEX "upload_workspace_id";
-- reverse: create "uploads" table
DROP TABLE "uploads";
-- reverse: create index "template_workspace_id_kind" to table: "templates"
DROP INDEX "template_workspace_id_kind";
-- reverse: create "templates" table
DROP TABLE "templates";
-- reverse: create index "teammember_user_id" to table: "team_members"
DROP INDEX "teammember_user_id";
-- reverse: create index "teammember_team_id_user_id" to table: "team_members"
DROP INDEX "teammember_team_id_user_id";
-- reverse: create "team_members" table
DROP TABLE "team_members";
-- reverse: create index "team_workspace_id_name" to table: "teams"
DROP INDEX "team_workspace_id_name";
-- reverse: create index "team_workspace_id_key" to table: "teams"
DROP INDEX "team_workspace_id_key";
-- reverse: create "teams" table
DROP TABLE "teams";
-- reverse: create "system_settings" table
DROP TABLE "system_settings";
-- reverse: create index "reaction_user_id_emoji_project_update_id" to table: "reactions"
DROP INDEX "reaction_user_id_emoji_project_update_id";
-- reverse: create index "reaction_user_id_emoji_issue_id" to table: "reactions"
DROP INDEX "reaction_user_id_emoji_issue_id";
-- reverse: create index "reaction_user_id_emoji_comment_id" to table: "reactions"
DROP INDEX "reaction_user_id_emoji_comment_id";
-- reverse: create "reactions" table
DROP TABLE "reactions";
-- reverse: create "project_updates" table
DROP TABLE "project_updates";
-- reverse: create index "projectstatus_workspace_id" to table: "project_statuses"
DROP INDEX "projectstatus_workspace_id";
-- reverse: create "project_statuses" table
DROP TABLE "project_statuses";
-- reverse: create "project_milestones" table
DROP TABLE "project_milestones";
-- reverse: create "project_links" table
DROP TABLE "project_links";
-- reverse: create index "project_workspace_id_slug" to table: "projects"
DROP INDEX "project_workspace_id_slug";
-- reverse: create index "project_status_id" to table: "projects"
DROP INDEX "project_status_id";
-- reverse: create "projects" table
DROP TABLE "projects";
-- reverse: create index "passwordresettoken_expires_at" to table: "password_reset_tokens"
DROP INDEX "passwordresettoken_expires_at";
-- reverse: create index "password_reset_tokens_token_hash_key" to table: "password_reset_tokens"
DROP INDEX "password_reset_tokens_token_hash_key";
-- reverse: create "password_reset_tokens" table
DROP TABLE "password_reset_tokens";
-- reverse: create index "oauth2_tokens_refresh_token_key" to table: "oauth2_tokens"
DROP INDEX "oauth2_tokens_refresh_token_key";
-- reverse: create index "oauth2_tokens_access_token_key" to table: "oauth2_tokens"
DROP INDEX "oauth2_tokens_access_token_key";
-- reverse: create "oauth2_tokens" table
DROP TABLE "oauth2_tokens";
-- reverse: create index "oauthgranttoken_grant_id" to table: "oauth_grant_tokens"
DROP INDEX "oauthgranttoken_grant_id";
-- reverse: create index "oauthgranttoken_expires_at" to table: "oauth_grant_tokens"
DROP INDEX "oauthgranttoken_expires_at";
-- reverse: create index "oauth_grant_tokens_token_hash_key" to table: "oauth_grant_tokens"
DROP INDEX "oauth_grant_tokens_token_hash_key";
-- reverse: create "oauth_grant_tokens" table
DROP TABLE "oauth_grant_tokens";
-- reverse: create index "oauthgrant_user_id" to table: "oauth_grants"
DROP INDEX "oauthgrant_user_id";
-- reverse: create index "oauthgrant_client_id" to table: "oauth_grants"
DROP INDEX "oauthgrant_client_id";
-- reverse: create "oauth_grants" table
DROP TABLE "oauth_grants";
-- reverse: create index "oauthclient_created_at" to table: "oauth_clients"
DROP INDEX "oauthclient_created_at";
-- reverse: create index "oauth_clients_client_id_key" to table: "oauth_clients"
DROP INDEX "oauth_clients_client_id_key";
-- reverse: create "oauth_clients" table
DROP TABLE "oauth_clients";
-- reverse: create index "oauthauthorizationcode_expires_at" to table: "oauth_authorization_codes"
DROP INDEX "oauthauthorizationcode_expires_at";
-- reverse: create index "oauth_authorization_codes_code_hash_key" to table: "oauth_authorization_codes"
DROP INDEX "oauth_authorization_codes_code_hash_key";
-- reverse: create "oauth_authorization_codes" table
DROP TABLE "oauth_authorization_codes";
-- reverse: create index "notification_settings_user_id_key" to table: "notification_settings"
DROP INDEX "notification_settings_user_id_key";
-- reverse: create "notification_settings" table
DROP TABLE "notification_settings";
-- reverse: create index "notification_user_id_archived_at_read_at" to table: "notifications"
DROP INDEX "notification_user_id_archived_at_read_at";
-- reverse: create index "notification_emailed_at" to table: "notifications"
DROP INDEX "notification_emailed_at";
-- reverse: create "notifications" table
DROP TABLE "notifications";
-- reverse: create index "label_workspace_id_team_id" to table: "labels"
DROP INDEX "label_workspace_id_team_id";
-- reverse: create index "label_parent_id" to table: "labels"
DROP INDEX "label_parent_id";
-- reverse: create "labels" table
DROP TABLE "labels";
-- reverse: create "jwt_keys" table
DROP TABLE "jwt_keys";
-- reverse: create index "issuerelation_related_issue_id" to table: "issue_relations"
DROP INDEX "issuerelation_related_issue_id";
-- reverse: create index "issuerelation_issue_id_related_issue_id_type" to table: "issue_relations"
DROP INDEX "issuerelation_issue_id_related_issue_id_type";
-- reverse: create "issue_relations" table
DROP TABLE "issue_relations";
-- reverse: create index "issuemove_old_team_id_old_number" to table: "issue_moves"
DROP INDEX "issuemove_old_team_id_old_number";
-- reverse: create "issue_moves" table
DROP TABLE "issue_moves";
-- reverse: create index "issuehistory_issue_id_created_at" to table: "issue_history"
DROP INDEX "issuehistory_issue_id_created_at";
-- reverse: create "issue_history" table
DROP TABLE "issue_history";
-- reverse: create index "issues_title_trgm" to table: "issues"
DROP INDEX "issues_title_trgm";
-- reverse: create index "issues_description_text_trgm" to table: "issues"
DROP INDEX "issues_description_text_trgm";
-- reverse: create index "issue_workspace_id_updated_at" to table: "issues"
DROP INDEX "issue_workspace_id_updated_at";
-- reverse: create index "issue_workspace_id_import_source_import_external_id" to table: "issues"
DROP INDEX "issue_workspace_id_import_source_import_external_id";
-- reverse: create index "issue_workspace_id_assignee_id" to table: "issues"
DROP INDEX "issue_workspace_id_assignee_id";
-- reverse: create index "issue_team_id_state_id" to table: "issues"
DROP INDEX "issue_team_id_state_id";
-- reverse: create index "issue_team_id_number" to table: "issues"
DROP INDEX "issue_team_id_number";
-- reverse: create index "issue_project_id" to table: "issues"
DROP INDEX "issue_project_id";
-- reverse: create index "issue_parent_id" to table: "issues"
DROP INDEX "issue_parent_id";
-- reverse: create index "issue_cycle_id" to table: "issues"
DROP INDEX "issue_cycle_id";
-- reverse: create "issues" table
DROP TABLE "issues";
-- reverse: create index "invitations_token_hash_key" to table: "invitations"
DROP INDEX "invitations_token_hash_key";
-- reverse: create index "invitation_workspace_id_email" to table: "invitations"
DROP INDEX "invitation_workspace_id_email";
-- reverse: create "invitations" table
DROP TABLE "invitations";
-- reverse: create "initiative_updates" table
DROP TABLE "initiative_updates";
-- reverse: create index "initiative_workspace_id_slug" to table: "initiatives"
DROP INDEX "initiative_workspace_id_slug";
-- reverse: create "initiatives" table
DROP TABLE "initiatives";
-- reverse: create index "importjob_workspace_id_created_at" to table: "imports"
DROP INDEX "importjob_workspace_id_created_at";
-- reverse: create "imports" table
DROP TABLE "imports";
-- reverse: create index "favorite_user_id_workspace_id" to table: "favorites"
DROP INDEX "favorite_user_id_workspace_id";
-- reverse: create index "favorite_user_id_kind_entity_id" to table: "favorites"
DROP INDEX "favorite_user_id_kind_entity_id";
-- reverse: create "favorites" table
DROP TABLE "favorites";
-- reverse: create index "document_workspace_id_slug" to table: "documents"
DROP INDEX "document_workspace_id_slug";
-- reverse: create "documents" table
DROP TABLE "documents";
-- reverse: create index "cycle_team_id_number" to table: "cycles"
DROP INDEX "cycle_team_id_number";
-- reverse: create index "cycle_ends_at" to table: "cycles"
DROP INDEX "cycle_ends_at";
-- reverse: create "cycles" table
DROP TABLE "cycles";
-- reverse: create index "comment_issue_id_created_at" to table: "comments"
DROP INDEX "comment_issue_id_created_at";
-- reverse: create "comments" table
DROP TABLE "comments";
-- reverse: create index "attachment_issue_id" to table: "attachments"
DROP INDEX "attachment_issue_id";
-- reverse: create "attachments" table
DROP TABLE "attachments";
-- reverse: create index "apikey_user_id_name" to table: "api_keys"
DROP INDEX "apikey_user_id_name";
-- reverse: create index "api_keys_token_hash_key" to table: "api_keys"
DROP INDEX "api_keys_token_hash_key";
-- reverse: create "api_keys" table
DROP TABLE "api_keys";
