import assert from "node:assert/strict";
import { test } from "node:test";

import { groupIssues } from "./grouping.ts";
import type { TIssue, TWorkflowStateSummary } from "@/lib/queries/issues";

const todo: TWorkflowStateSummary = {
  id: "s-todo",
  team_id: "t",
  name: "Todo",
  type: "unstarted",
  color: "#e2e2e2",
  description: "",
  position: 0,
};
const doing: TWorkflowStateSummary = {
  ...todo,
  id: "s-doing",
  name: "In Progress",
  type: "started",
};
const backlog: TWorkflowStateSummary = {
  ...todo,
  id: "s-backlog",
  name: "Backlog",
  type: "backlog",
};

function issue(id: string, overrides: Partial<TIssue> = {}): TIssue {
  return {
    id,
    identifier: `ENG-${id}`,
    number: 1,
    workspace_id: "w",
    team_id: "t",
    team_key: "ENG",
    title: id,
    description: "",
    priority: 0,
    priority_label: "No priority",
    estimate: null,
    state: todo,
    labels: [],
    child_count: 0,
    completed_child_count: 0,
    comment_count: 0,
    attachment_count: 0,
    sort_order: 0,
    sub_issue_sort_order: 0,
    due_date: null,
    started_at: null,
    completed_at: null,
    canceled_at: null,
    triaged_at: null,
    snoozed_until: null,
    archived_at: null,
    trashed_at: null,
    git_branch_name: "",
    url: "",
    import_url: "",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

test("state groups follow the workflow order and keep empty team states", () => {
  const groups = groupIssues({
    issues: [issue("a", { state: doing }), issue("b")],
    groupBy: "state",
    orderBy: "manual",
    order: "asc",
    states: [doing, todo, backlog],
  });
  assert.deepEqual(
    groups.map((group) => [group.label, group.issues.length]),
    [
      ["Backlog", 0],
      ["Todo", 1],
      ["In Progress", 1],
    ],
  );
});

test("issues inside a group are ordered by the display ordering", () => {
  const [group] = groupIssues({
    issues: [issue("a", { sort_order: 3 }), issue("b", { sort_order: 1 })],
    groupBy: "none",
    orderBy: "manual",
    order: "asc",
  });
  assert.deepEqual(
    group.issues.map((item) => item.id),
    ["b", "a"],
  );
});

test("priority groups put urgent first and no priority last", () => {
  const groups = groupIssues({
    issues: [issue("a", { priority: 0 }), issue("b", { priority: 1 })],
    groupBy: "priority",
    orderBy: "manual",
    order: "asc",
  });
  assert.deepEqual(
    groups.map((group) => group.label),
    ["Urgent", "High", "Medium", "Low", "No priority"],
  );
});

test("an issue with two labels shows up under both", () => {
  const groups = groupIssues({
    issues: [
      issue("a", {
        labels: [
          { id: "l1", name: "Bug", color: "#f00", parent_id: null },
          { id: "l2", name: "UI", color: "#0f0", parent_id: null },
        ],
      }),
      issue("b"),
    ],
    groupBy: "label",
    orderBy: "manual",
    order: "asc",
  });
  assert.deepEqual(
    groups.map((group) => [group.label, group.issues.map((item) => item.id)]),
    [
      ["Bug", ["a"]],
      ["UI", ["a"]],
      ["No label", ["b"]],
    ],
  );
});

test("unassigned issues group last under no assignee", () => {
  const groups = groupIssues({
    issues: [
      issue("a"),
      issue("b", { assignee: { id: "u1", name: "Zed", display_name: "zed", avatar_url: null } }),
    ],
    groupBy: "assignee",
    orderBy: "manual",
    order: "asc",
  });
  assert.deepEqual(
    groups.map((group) => group.label),
    ["Zed", "No assignee"],
  );
});
