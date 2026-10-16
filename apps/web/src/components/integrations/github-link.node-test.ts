import assert from "node:assert/strict";
import { test } from "node:test";

import { githubLinkOf } from "./github-link.ts";

test("pull request attachments read their repository, number and state", () => {
  assert.deepEqual(
    githubLinkOf({
      source_type: "github",
      metadata: { repo: "unbindapp/unbind", number: 42, state: "merged", branch: "unb-1" },
    }),
    { kind: "pull_request", reference: "unbindapp/unbind#42", state: "merged" },
  );
});

test("an unknown pull request state shows no chip", () => {
  assert.deepEqual(
    githubLinkOf({ source_type: "github", metadata: { repo: "a/b", number: 1, state: "weird" } }),
    { kind: "pull_request", reference: "a/b#1", state: null },
  );
});

test("branch attachments read their repository", () => {
  assert.deepEqual(
    githubLinkOf({ source_type: "github_branch", metadata: { repo: "a/b", branch: "unb-1" } }),
    { kind: "branch", reference: "a/b" },
  );
});

test("links and uploads are not GitHub attachments", () => {
  assert.equal(githubLinkOf({ source_type: "link", metadata: {} }), null);
  assert.equal(githubLinkOf({ source_type: "upload", metadata: {} }), null);
});
