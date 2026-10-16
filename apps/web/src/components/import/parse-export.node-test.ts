import assert from "node:assert/strict";
import { test } from "node:test";

import { countRows, parseLinearExport, reportSummary } from "./parse-export.ts";

test("a Linear export parses with a summary of what it holds", () => {
  const text = JSON.stringify({
    workspace: { name: "Unbind" },
    users: [{}, {}],
    teams: [{}],
    projects: [{}],
    issues: [{}, {}, {}],
  });
  const result = parseLinearExport("export.json", text);
  assert.equal(result.ok, true);
  if (!result.ok) return;
  assert.equal(result.parsed.fileName, "export.json");
  assert.equal(result.parsed.summary, "3 issues, 1 team, 1 project, 2 users from Unbind");
});

test("files that are not Linear exports are refused with a reason", () => {
  assert.deepEqual(parseLinearExport("notes.txt", "hello"), {
    ok: false,
    error: "notes.txt is not a JSON file.",
  });
  assert.deepEqual(parseLinearExport("list.json", "[]"), {
    ok: false,
    error: "list.json is not a Linear export.",
  });
  const missing = parseLinearExport("other.json", JSON.stringify({ issues: [] }));
  assert.equal(missing.ok, false);
});

test("count rows cover every entity in a fixed order, zero when missing", () => {
  const rows = countRows({ counts: { issues: { created: 456, updated: 0, skipped: 0 } } });
  assert.equal(rows.length, 12);
  assert.equal(rows[0].key, "users");
  assert.deepEqual(rows[0], { key: "users", label: "Users", created: 0, updated: 0, skipped: 0 });
  assert.equal(rows.find((row) => row.key === "issues")?.created, 456);
});

test("the summary names what happened to issues", () => {
  assert.equal(
    reportSummary({ counts: { issues: { created: 3, updated: 2, skipped: 1 } } }),
    "Issues: 3 created, 2 updated, 1 skipped",
  );
  assert.equal(reportSummary({ counts: {} }), "No issues");
});
