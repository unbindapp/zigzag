import assert from "node:assert/strict";
import { test } from "node:test";

import { formatProjectDate, snapToResolution } from "./project-dates.ts";

test("dates read at their resolution", () => {
  assert.equal(formatProjectDate("2026-10-16"), "Oct 16, 2026");
  assert.equal(formatProjectDate("2026-10-16", "month"), "Oct 2026");
  assert.equal(formatProjectDate("2026-10-16", "quarter"), "Q4 2026");
  assert.equal(formatProjectDate("2026-05-01", "half_year"), "H1 2026");
  assert.equal(formatProjectDate("2026-07-01", "half_year"), "H2 2026");
  assert.equal(formatProjectDate("2026-02-28", "year"), "2026");
});

test("coarse targets snap to the end of their period", () => {
  assert.equal(snapToResolution("2026-10-16", "day"), "2026-10-16");
  assert.equal(snapToResolution("2026-02-03", "month"), "2026-02-28");
  assert.equal(snapToResolution("2028-02-03", "month"), "2028-02-29");
  assert.equal(snapToResolution("2026-08-10", "quarter"), "2026-09-30");
  assert.equal(snapToResolution("2026-03-10", "half_year"), "2026-06-30");
  assert.equal(snapToResolution("2026-03-10", "year"), "2026-12-31");
});
