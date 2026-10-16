import assert from "node:assert/strict";
import { test } from "node:test";

import { gitBranchName, titleSlug } from "./git-branch.ts";

test("branch names use the lowercased identifier and a title slug", () => {
  assert.equal(
    gitBranchName({ identifier: "ENG-42", title: "Fix login redirect" }),
    "eng-42-fix-login-redirect",
  );
});

test("assigned callers get their display name as a prefix", () => {
  assert.equal(
    gitBranchName({ identifier: "ENG-42", title: "Fix login", displayName: "ada" }),
    "ada/eng-42-fix-login",
  );
});

test("title slugs are capped at 50 characters without a trailing dash", () => {
  const slug = titleSlug("word ".repeat(30));
  assert.ok(slug.length <= 50);
  assert.ok(!slug.endsWith("-"));
});

test("titles without usable characters fall back to the identifier", () => {
  assert.equal(gitBranchName({ identifier: "OPS-1", title: "!!!" }), "ops-1");
});

test("accents are folded", () => {
  assert.equal(titleSlug("Crème brûlée"), "creme-brulee");
});
