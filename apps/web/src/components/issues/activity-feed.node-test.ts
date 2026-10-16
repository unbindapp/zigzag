import assert from "node:assert/strict";
import { test } from "node:test";

import { mergeFeed } from "./activity-feed.ts";

const item = (id: string, at: string) => ({ id, at });

test("interleaves complete streams by time", () => {
  const feed = mergeFeed([
    {
      items: [item("h1", "2026-01-01T10:00:00Z"), item("h2", "2026-01-01T12:00:00Z")],
      hasMore: false,
    },
    { items: [item("c1", "2026-01-01T11:00:00Z")], hasMore: false },
  ]);
  assert.deepEqual(
    feed.map((entry) => entry.id),
    ["h1", "c1", "h2"],
  );
});

test("hides items newer than the last loaded item of a stream with more pages", () => {
  const feed = mergeFeed([
    {
      items: [item("h1", "2026-01-01T10:00:00Z"), item("h2", "2026-01-01T11:00:00Z")],
      hasMore: true,
    },
    {
      items: [item("c1", "2026-01-01T10:30:00Z"), item("c2", "2026-01-01T13:00:00Z")],
      hasMore: false,
    },
  ]);
  assert.deepEqual(
    feed.map((entry) => entry.id),
    ["h1", "c1", "h2"],
  );
});

test("orders by instant, not by string, when fractional seconds differ in length", () => {
  const feed = mergeFeed([
    { items: [item("late", "2026-01-01T10:00:00.89Z")], hasMore: false },
    { items: [item("later", "2026-01-01T10:00:00.893843Z")], hasMore: false },
  ]);
  assert.deepEqual(
    feed.map((entry) => entry.id),
    ["late", "later"],
  );
});
