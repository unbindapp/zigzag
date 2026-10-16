import assert from "node:assert/strict";
import { test } from "node:test";

import { snoozeOptions } from "./snooze.ts";

const byId = (now: Date) =>
  Object.fromEntries(snoozeOptions(now).map((option) => [option.id, option.until]));

test("in 1 hour is exactly an hour later", () => {
  const now = new Date(2026, 9, 16, 22, 30);
  assert.equal(byId(now).hour.getTime() - now.getTime(), 60 * 60 * 1000);
});

test("tomorrow is 9:00 the next calendar day, even late at night", () => {
  const until = byId(new Date(2026, 9, 16, 23, 50)).tomorrow;
  assert.deepEqual(
    [until.getFullYear(), until.getMonth(), until.getDate(), until.getHours(), until.getMinutes()],
    [2026, 9, 17, 9, 0],
  );
});

test("next week is the coming Monday at 9:00", () => {
  const fromFriday = byId(new Date(2026, 9, 16, 10, 0))["next-week"];
  assert.deepEqual([fromFriday.getDate(), fromFriday.getDay(), fromFriday.getHours()], [19, 1, 9]);
  const fromMonday = byId(new Date(2026, 9, 19, 8, 0))["next-week"];
  assert.equal(fromMonday.getDate(), 26, "a Monday snoozes to the following Monday");
});
