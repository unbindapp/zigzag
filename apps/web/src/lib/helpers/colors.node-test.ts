import assert from "node:assert/strict";
import { test } from "node:test";

import { colorFromId, initials, isHexColor, PRESET_COLORS } from "./colors.ts";

test("the same id always gets the same preset color", () => {
  const id = "0b6c5e1e-6a7d-4f3f-9a3c-1f2d3e4f5a6b";
  assert.equal(colorFromId(id), colorFromId(id));
  assert.ok(PRESET_COLORS.includes(colorFromId(id)));
});

test("initials use first and last word", () => {
  assert.equal(initials("Ada Lovelace"), "AL");
  assert.equal(initials("Grace Brewster Hopper"), "GH");
  assert.equal(initials("plato"), "PL");
  assert.equal(initials("   "), "?");
});

test("hex colors need six digits", () => {
  assert.equal(isHexColor("#5e6ad2"), true);
  assert.equal(isHexColor("#fff"), false);
  assert.equal(isHexColor("5e6ad2"), false);
  assert.equal(isHexColor(null), false);
});
