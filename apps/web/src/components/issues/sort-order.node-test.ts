import assert from "node:assert/strict";
import { test } from "node:test";

import { sortOrderAtIndex, sortOrderBetween } from "./sort-order.ts";

test("a drop between two neighbors lands on their midpoint", () => {
  assert.equal(sortOrderBetween(1, 2), 1.5);
  assert.equal(sortOrderAtIndex([1, 2, 4], 2), 3);
});

test("a drop at either end steps past the outer neighbor", () => {
  assert.equal(sortOrderAtIndex([5, 6], 0), 4);
  assert.equal(sortOrderAtIndex([5, 6], 2), 7);
  assert.equal(sortOrderAtIndex([5, 6], 9), 7);
});

test("a drop into an empty column starts at zero", () => {
  assert.equal(sortOrderAtIndex([], 0), 0);
});
