import assert from "node:assert/strict";
import { test } from "node:test";

import {
  githubCommentAuthor,
  githubLoginSchema,
  githubTokenSchema,
  importSummary,
  isImporting,
  linkedIssuesLabel,
  repositoryNameSchema,
} from "./github-sync.ts";

test("GitHub usernames are lowercased, lose the @ and may be cleared", () => {
  assert.equal(githubLoginSchema.parse(" @Octo-Cat "), "octo-cat");
  assert.equal(githubLoginSchema.parse(""), "");
  for (const bad of ["-octo", "octo-", "octo--cat", "octo_cat", "a".repeat(40)]) {
    assert.equal(githubLoginSchema.safeParse(bad).success, false, bad);
  }
});

test("repository names accept owner/name and GitHub URLs", () => {
  assert.equal(repositoryNameSchema.parse("unbindapp/zigzag"), "unbindapp/zigzag");
  assert.equal(
    repositoryNameSchema.parse("https://github.com/zigzagapp-org/zigzag.git"),
    "unbindapp/zigzag",
  );
  assert.equal(repositoryNameSchema.safeParse("zigzag").success, false);
  assert.equal(repositoryNameSchema.safeParse("unbindapp/zigzag/issues").success, false);
});

test("tokens must be whole and without spaces", () => {
  assert.equal(githubTokenSchema.safeParse("ghp_short").success, false);
  assert.equal(githubTokenSchema.safeParse("github_pat_ 11ABCDEFGHIJKLMNOPQRS").success, false);
  assert.equal(githubTokenSchema.safeParse(" github_pat_11ABCDEFGHIJKLMNOPQRS ").success, true);
});

test("import summaries read the counts of the last import", () => {
  const report = { issues_created: 3, issues_updated: 1, comments_created: 1, skipped: 2 };
  assert.equal(
    importSummary({ backfill_status: "completed", backfill_report: report }),
    "3 created, 1 updated, 1 comment, 2 skipped",
  );
  assert.equal(importSummary({ backfill_status: "none" }), null);
  assert.equal(isImporting({ backfill_status: "running", backfill_report: report }), true);
  assert.equal(isImporting({ backfill_status: "failed" }), false);
});

test("linked issue counts read naturally", () => {
  assert.equal(linkedIssuesLabel(1), "1 issue");
  assert.equal(linkedIssuesLabel(0), "0 issues");
});

test("comments synced from GitHub users without an account name their author", () => {
  assert.deepEqual(githubCommentAuthor("**@octocat on GitHub**\n\nSeen it\ntoo"), {
    login: "octocat",
    body: "Seen it\ntoo",
  });
  assert.equal(githubCommentAuthor("**Eve on Zigzag**\n\nOn it"), null);
  assert.equal(githubCommentAuthor("I saw **@octocat on GitHub**\n"), null);
});
