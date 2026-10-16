import assert from "node:assert/strict";
import { test } from "node:test";

import { nextTheme } from "./themes.ts";

test("toggling cycles system, light, dark and back", () => {
  assert.equal(nextTheme("system"), "light");
  assert.equal(nextTheme("light"), "dark");
  assert.equal(nextTheme("dark"), "system");
});

test("an unset or unknown theme toggles as if it were system", () => {
  assert.equal(nextTheme(undefined), "light");
  assert.equal(nextTheme("sepia"), "light");
});
