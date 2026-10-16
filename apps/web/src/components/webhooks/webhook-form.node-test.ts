import assert from "node:assert/strict";
import { test } from "node:test";

import { ALL_TEAMS, toCreateInput, webhookFormSchema } from "./webhook-form.ts";

const valid = {
  url: " https://hooks.example.com/zigzag ",
  label: " Slack bridge ",
  events: ["issue.created"],
  team: ALL_TEAMS,
  enabled: true,
};

test("a complete form passes and maps to the create input", () => {
  const parsed = webhookFormSchema.parse(valid);
  assert.deepEqual(toCreateInput("ws", parsed), {
    workspace_id: "ws",
    url: "https://hooks.example.com/zigzag",
    label: "Slack bridge",
    events: ["issue.created"],
    enabled: true,
    team_id: undefined,
  });
  assert.equal(toCreateInput("ws", { ...parsed, team: "team-1" }).team_id, "team-1");
});

test("urls must be http or https and events are required", () => {
  assert.equal(webhookFormSchema.safeParse({ ...valid, url: "ftp://x" }).success, false);
  assert.equal(webhookFormSchema.safeParse({ ...valid, url: "hooks.example.com" }).success, false);
  assert.equal(webhookFormSchema.safeParse({ ...valid, events: [] }).success, false);
  assert.equal(
    webhookFormSchema.safeParse({ ...valid, events: ["issue.exploded"] }).success,
    false,
  );
});
