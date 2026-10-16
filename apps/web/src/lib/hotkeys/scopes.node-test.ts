import assert from "node:assert/strict";
import { test } from "node:test";

import { hotkeyOwners } from "./scopes.ts";

const scope = (id: string, keys: string[], modal = false) => ({ id, keys, modal });

test("the scope activated last owns a key both scopes bind", () => {
  const owners = hotkeyOwners([scope("inbox", ["E", "J"]), scope("issue", ["E", "S"])]);
  assert.equal(owners.get("e"), "issue");
  assert.equal(owners.get("j"), "inbox");
  assert.equal(owners.get("s"), "issue");
});

test("a modal scope hides every key of the scopes under it", () => {
  const owners = hotkeyOwners([
    scope("shell", ["C", "Mod+K"]),
    scope("issue", ["S"]),
    scope("dialog", ["Mod+Enter"], true),
  ]);
  assert.equal(owners.get("c"), undefined);
  assert.equal(owners.get("s"), undefined);
  assert.equal(owners.get("enter+mod"), "dialog");
});

test("keys compare without case or modifier order", () => {
  const owners = hotkeyOwners([scope("a", ["Shift+P"]), scope("b", ["p+shift"])]);
  assert.equal(owners.get("p+shift"), "b");
  assert.equal(owners.size, 1);
});
