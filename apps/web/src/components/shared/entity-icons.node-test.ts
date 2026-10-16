import assert from "node:assert/strict";
import { test } from "node:test";

import { resolveEntityIcon } from "./entity-icons.ts";

test("known icon names resolve to their lucide icon", () => {
  assert.equal(resolveEntityIcon("rocket").kind, "lucide");
});

test("emoji render as emoji", () => {
  assert.deepEqual(resolveEntityIcon("🚀"), { kind: "emoji", emoji: "🚀" });
});

test("unknown icon names, such as imported Linear icon names, fall back", () => {
  assert.equal(resolveEntityIcon("Chip").kind, "none");
  assert.equal(resolveEntityIcon(null).kind, "none");
  assert.equal(resolveEntityIcon("").kind, "none");
});
