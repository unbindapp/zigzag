import assert from "node:assert/strict";
import { test } from "node:test";

import { estimateOptions, formatEstimate } from "./estimates.ts";

test("no options when estimates are off", () => {
  assert.deepEqual(estimateOptions({ type: "none", allowZero: true, extended: true }), []);
});

test("scales follow the type and the extended flag", () => {
  assert.deepEqual(
    estimateOptions({ type: "fibonacci", allowZero: false, extended: false }),
    [1, 2, 3, 5, 8],
  );
  assert.deepEqual(
    estimateOptions({ type: "exponential", allowZero: true, extended: true }),
    [0, 1, 2, 4, 8, 16, 32, 64],
  );
});

test("t-shirt estimates read as sizes", () => {
  assert.equal(formatEstimate(3, "tshirt"), "M");
  assert.equal(formatEstimate(21, "tshirt"), "XXXL");
  assert.equal(formatEstimate(3, "linear"), "3");
});
