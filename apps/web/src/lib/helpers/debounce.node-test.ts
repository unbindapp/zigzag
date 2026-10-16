import assert from "node:assert/strict";
import { mock, test } from "node:test";

import { createDebounce } from "./debounce.ts";

function setup() {
  mock.timers.enable({ apis: ["setTimeout"] });
  const calls: string[] = [];
  const debounced = createDebounce((value: string) => calls.push(value), 800);
  return { calls, debounced };
}

test("calls once with the last arguments after the delay", (t) => {
  const { calls, debounced } = setup();
  t.after(() => mock.timers.reset());
  debounced.call("a");
  mock.timers.tick(500);
  debounced.call("ab");
  mock.timers.tick(799);
  assert.deepEqual(calls, []);
  mock.timers.tick(1);
  assert.deepEqual(calls, ["ab"]);
  assert.equal(debounced.isPending(), false);
});

test("flush runs a pending call immediately and only once", (t) => {
  const { calls, debounced } = setup();
  t.after(() => mock.timers.reset());
  debounced.call("draft");
  assert.equal(debounced.isPending(), true);
  debounced.flush();
  assert.deepEqual(calls, ["draft"]);
  mock.timers.tick(1000);
  assert.deepEqual(calls, ["draft"]);
});

test("flush without a pending call does nothing", (t) => {
  const { calls, debounced } = setup();
  t.after(() => mock.timers.reset());
  debounced.flush();
  assert.deepEqual(calls, []);
});

test("cancel drops a pending call", (t) => {
  const { calls, debounced } = setup();
  t.after(() => mock.timers.reset());
  debounced.call("draft");
  debounced.cancel();
  mock.timers.tick(1000);
  debounced.flush();
  assert.deepEqual(calls, []);
});
