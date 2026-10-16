import assert from "node:assert/strict";
import { test } from "node:test";

import { nestSubIssues } from "./nesting.ts";

const issue = (id: string, parentId?: string) => ({
  id,
  parent: parentId ? { id: parentId } : undefined,
});

const shape = (issues: ReturnType<typeof issue>[]) =>
  nestSubIssues(issues).map(({ issue: item, depth }) => `${item.id}:${depth}`);

test("sub-issues follow their parent, indented, keeping their order", () => {
  assert.deepEqual(shape([issue("c2", "p"), issue("a"), issue("p"), issue("c1", "p")]), [
    "a:0",
    "p:0",
    "c2:1",
    "c1:1",
  ]);
});

test("a sub-issue whose parent is not in the list stays a top-level row", () => {
  assert.deepEqual(shape([issue("a"), issue("orphan", "elsewhere")]), ["a:0", "orphan:0"]);
});

test("nests more than one level", () => {
  assert.deepEqual(shape([issue("g", "c"), issue("c", "p"), issue("p")]), ["p:0", "c:1", "g:2"]);
});

test("a parent cycle still lists every issue once", () => {
  assert.deepEqual(shape([issue("x", "y"), issue("y", "x")]).length, 2);
});
