-- +goose Up
-- modify "workspaces" table
ALTER TABLE "workspaces" ADD COLUMN "icon_upload_id" uuid NULL, ADD
CONSTRAINT "workspaces_uploads_icon_workspaces" FOREIGN KEY ("icon_upload_id") REFERENCES "uploads" ("id") ON UPDATE NO ACTION ON DELETE SET NULL;

-- +goose Down
-- reverse: modify "workspaces" table
ALTER TABLE "workspaces" DROP CONSTRAINT "workspaces_uploads_icon_workspaces", DROP COLUMN "icon_upload_id";
