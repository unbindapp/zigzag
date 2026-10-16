-- +goose Up
-- create "upload_blobs" table
CREATE TABLE "upload_blobs" (
  "id" uuid NOT NULL,
  "data" bytea NOT NULL,
  "created_at" timestamptz NOT NULL,
  "upload_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "upload_blobs_uploads_blob" FOREIGN KEY ("upload_id") REFERENCES "uploads" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "upload_blobs_upload_id_key" to table: "upload_blobs"
CREATE UNIQUE INDEX "upload_blobs_upload_id_key" ON "upload_blobs" ("upload_id");

-- +goose Down
-- reverse: create index "upload_blobs_upload_id_key" to table: "upload_blobs"
DROP INDEX "upload_blobs_upload_id_key";
-- reverse: create "upload_blobs" table
DROP TABLE "upload_blobs";
