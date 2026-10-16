import assert from "node:assert/strict";
import { test } from "node:test";

import {
  deriveTeamKey,
  formatIdentifier,
  parseIdentifier,
  sanitizeTeamKey,
  slugify,
  teamKeyPattern,
  workspaceSlugPattern,
} from "./identifiers.ts";

test("identifiers parse into key and number", () => {
  assert.deepEqual(parseIdentifier("ENG-123"), { key: "ENG", number: 123 });
  assert.deepEqual(parseIdentifier(" eng-7 "), { key: "ENG", number: 7 });
  assert.equal(parseIdentifier("ENG123"), null);
  assert.equal(parseIdentifier("1NG-1"), null);
  assert.equal(parseIdentifier("ENG-"), null);
});

test("identifiers format uppercase", () => {
  assert.equal(formatIdentifier("eng", 42), "ENG-42");
});

test("team keys derive from the name", () => {
  assert.equal(deriveTeamKey("Engineering"), "ENG");
  assert.equal(deriveTeamKey("Product design"), "PD");
  assert.equal(deriveTeamKey("Mobile iOS and Android apps team rocks"), "MIAAATR");
  assert.equal(deriveTeamKey("3D printing"), "PRI");
  assert.equal(deriveTeamKey("  "), "");
});

test("derived keys always match the key pattern", () => {
  for (const name of ["Engineering", "a b c d e f g h i", "x", "Café crème", "Q3 goals"]) {
    assert.match(deriveTeamKey(name), teamKeyPattern);
  }
});

test("typed keys are sanitized", () => {
  assert.equal(sanitizeTeamKey("eng-1"), "ENG1");
  assert.equal(sanitizeTeamKey("12ab"), "AB");
  assert.equal(sanitizeTeamKey("abcdefghij"), "ABCDEFG");
});

test("slugs are lowercase dashed and bounded", () => {
  assert.equal(slugify("Acme Corp!"), "acme-corp");
  assert.equal(slugify("  --Hello__World--  "), "hello-world");
  assert.match(slugify("Zigzag Labs"), workspaceSlugPattern);
  assert.equal(slugify("a".repeat(40)).length, 32);
  assert.equal(slugify("abc def", 4), "abc");
});
