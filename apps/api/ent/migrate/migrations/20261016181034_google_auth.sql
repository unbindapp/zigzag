-- +goose Up
-- modify "system_settings" table
ALTER TABLE "system_settings" ADD COLUMN "google_auth" jsonb NULL, ADD COLUMN "google_client_secret_encrypted" bytea NULL;
-- modify "users" table
ALTER TABLE "users" ADD COLUMN "google_subject" character varying NULL;
-- create index "users_google_subject_key" to table: "users"
CREATE UNIQUE INDEX "users_google_subject_key" ON "users" ("google_subject");

-- +goose Down
-- reverse: create index "users_google_subject_key" to table: "users"
DROP INDEX "users_google_subject_key";
-- reverse: modify "users" table
ALTER TABLE "users" DROP COLUMN "google_subject";
-- reverse: modify "system_settings" table
ALTER TABLE "system_settings" DROP COLUMN "google_client_secret_encrypted", DROP COLUMN "google_auth";
