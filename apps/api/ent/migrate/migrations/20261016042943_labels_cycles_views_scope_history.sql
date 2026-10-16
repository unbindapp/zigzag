-- +goose Up
-- modify "labels" table
ALTER TABLE "labels" ADD COLUMN "sort_order" double precision NOT NULL DEFAULT 0;
-- create "cycle_scope_history" table
CREATE TABLE "cycle_scope_history" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL,
  "updated_at" timestamptz NOT NULL,
  "date" date NOT NULL,
  "scope" double precision NOT NULL,
  "completed" double precision NOT NULL,
  "started" double precision NOT NULL,
  "cycle_id" uuid NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "cycle_scope_history_cycles_scope_history" FOREIGN KEY ("cycle_id") REFERENCES "cycles" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- create index "cyclescopehistory_cycle_id_date" to table: "cycle_scope_history"
CREATE UNIQUE INDEX "cyclescopehistory_cycle_id_date" ON "cycle_scope_history" ("cycle_id", "date");

-- +goose Down
-- reverse: create index "cyclescopehistory_cycle_id_date" to table: "cycle_scope_history"
DROP INDEX "cyclescopehistory_cycle_id_date";
-- reverse: create "cycle_scope_history" table
DROP TABLE "cycle_scope_history";
-- reverse: modify "labels" table
ALTER TABLE "labels" DROP COLUMN "sort_order";
