import assert from "node:assert/strict";
import { test } from "node:test";

import { DEFAULT_DISPLAY } from "./display.ts";
import {
  activeFilterFields,
  buildListIssuesInput,
  clearFilterField,
  filtersFromSearch,
  isHiddenCompleted,
  issueSearchSchema,
  mergeFilters,
} from "./filters.ts";

test("search params parse into list-issues filters and drop bad values", () => {
  const search = issueSearchSchema.parse({
    assignee_ids: ["u1", "null"],
    priorities: [1, 9],
    due_before: "not a date",
    peek: "ENG-1",
  });
  assert.deepEqual(filtersFromSearch(search), { assignee_ids: ["u1", "null"] });
});

test("fixed filters narrow user filters instead of replacing them", () => {
  const merged = mergeFilters(
    { state_types: ["backlog", "started"], assignee_ids: ["u1"] },
    { state_types: ["backlog"], team_ids: ["t1"] },
  );
  assert.deepEqual(merged.state_types, ["backlog"]);
  assert.deepEqual(merged.team_ids, ["t1"]);
  assert.deepEqual(merged.assignee_ids, ["u1"]);
});

test("hiding completed issues removes completed and canceled state types", () => {
  const input = buildListIssuesInput({
    workspaceId: "w",
    filters: {},
    display: { ...DEFAULT_DISPLAY, show_completed: "none" },
  });
  assert.deepEqual(input.state_types, ["triage", "backlog", "unstarted", "started"]);
});

test("hiding sub-issues asks for top level issues only", () => {
  const input = buildListIssuesInput({
    workspaceId: "w",
    filters: { assignee_ids: [] },
    display: { ...DEFAULT_DISPLAY, show_sub_issues: false },
  });
  assert.equal(input.has_parent, "false");
  assert.equal(input.assignee_ids, undefined);
  assert.equal(input.workspace_id, "w");
  assert.equal(input.order_by, "manual");
});

test("active fields and clearing a field round trip", () => {
  const filters = { priorities: [1], due_after: "2026-01-01", label_ids: ["l"] };
  assert.deepEqual(activeFilterFields(filters), ["priority", "labels", "due_date"]);
  assert.deepEqual(clearFilterField(filters, "due_date"), { priorities: [1], label_ids: ["l"] });
});

test("past week keeps recently completed issues only", () => {
  const now = new Date("2026-03-10T00:00:00Z").getTime();
  const done = (completedAt: string) => ({
    state: { type: "completed" as const },
    completed_at: completedAt,
    canceled_at: null,
  });
  assert.equal(isHiddenCompleted(done("2026-03-08T00:00:00Z"), "past_week", now), false);
  assert.equal(isHiddenCompleted(done("2026-02-01T00:00:00Z"), "past_week", now), true);
  assert.equal(isHiddenCompleted(done("2026-02-01T00:00:00Z"), "all", now), false);
});
