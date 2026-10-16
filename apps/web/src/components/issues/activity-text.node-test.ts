import assert from "node:assert/strict";
import { test } from "node:test";

import { describeChange } from "./activity-text.ts";

const change = (field: string, from = "", to = "", fromLabel = "", toLabel = "") => ({
  field,
  from,
  to,
  from_label: fromLabel,
  to_label: toLabel,
});

test("state changes read with the state names", () => {
  assert.equal(
    describeChange(change("state", "id1", "id2", "Todo", "In Progress")),
    "changed status from Todo to In Progress",
  );
});

test("priority values turn into their names", () => {
  assert.equal(
    describeChange(change("priority", "0", "1")),
    "changed priority from No priority to Urgent",
  );
});

test("clearing the assignee reads as unassigning", () => {
  assert.equal(describeChange(change("assignee", "u1", "", "Ada")), "unassigned Ada");
  assert.equal(
    describeChange(change("assignee", "", "u1", "", "Ada")),
    "assigned the issue to Ada",
  );
});

test("other fields fall back to set and removed", () => {
  assert.equal(describeChange(change("due_date", "", "2026-03-01")), "set due date to 2026-03-01");
  assert.equal(describeChange(change("project", "p1", "", "Launch")), "removed the project");
});

test("GitHub events read with their label", () => {
  assert.equal(
    describeChange(
      change("github", "", "https://github.com/a/b/pull/42", "", "Pull request #42 opened"),
    ),
    "on GitHub: Pull request #42 opened",
  );
  assert.equal(describeChange(change("github")), "updated from GitHub");
});
