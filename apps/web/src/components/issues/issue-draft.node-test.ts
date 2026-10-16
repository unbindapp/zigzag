import assert from "node:assert/strict";
import { test } from "node:test";

import { applyPatch, draftToCreateInput, EMPTY_DRAFT, patchToOptimistic } from "./issue-draft.ts";

const project = { id: "p1", name: "Launch", slug: "launch", icon: "", color: "" };

test("changing the project drops the milestone of the old project", () => {
  const draft = applyPatch(
    { ...EMPTY_DRAFT, project, milestone: { id: "m1", name: "Beta" } },
    { project: null },
  );
  assert.equal(draft.milestone, null);
  assert.equal(patchToOptimistic({ project: null }).milestone, undefined);
});

test("a new issue only sends the properties that were set", () => {
  const input = draftToCreateInput({
    teamId: "t1",
    title: "  Fix login  ",
    description: " ",
    draft: {
      ...EMPTY_DRAFT,
      priority: 1,
      labels: [{ id: "l1", name: "Bug", color: "", parent_id: null }],
    },
  });
  assert.deepEqual(input, { team_id: "t1", title: "Fix login", priority: 1, label_ids: ["l1"] });
});
