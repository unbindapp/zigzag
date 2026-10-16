import assert from "node:assert/strict";
import { test } from "node:test";

import { notificationSummary, notificationTitle } from "./notification-text.ts";

test("the body wins when the API sent one", () => {
  assert.equal(
    notificationSummary({ type: "issue_commented", title: "", body: "  Looks good to me  " }),
    "Looks good to me",
  );
});

test("without a body the type is described, with the actor when known", () => {
  assert.equal(notificationSummary({ type: "issue_due_soon", title: "", body: "" }), "Due soon");
  assert.equal(
    notificationSummary({
      type: "issue_assigned",
      title: "",
      body: "",
      actor: { name: "Ada Lovelace", display_name: "ada" },
    }),
    "Ada Lovelace: assigned to you",
  );
  assert.equal(
    notificationSummary({
      type: "reaction",
      title: "",
      body: "",
      actor: { name: "", display_name: "ada" },
    }),
    "ada: reacted",
  );
});

test("the issue identifier is not repeated in the title", () => {
  const issue = { identifier: "ENG-12" };
  assert.equal(notificationTitle({ title: "ENG-12 Fix login", issue }), "Fix login");
  assert.equal(notificationTitle({ title: "Fix login", issue }), "Fix login");
  assert.equal(notificationTitle({ title: "ENG-123 Other", issue }), "ENG-123 Other");
  assert.equal(notificationTitle({ title: "Weekly update" }), "Weekly update");
});
