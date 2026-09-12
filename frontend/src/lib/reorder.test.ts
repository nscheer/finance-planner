import { test } from "node:test";
import assert from "node:assert/strict";
import { resolveMoveIndex } from "./reorder.ts";

test("moving an item downwards in the same list shifts the index", () => {
  // list [a,b,c,d], drag a (0) to "before d" (3) -> index 2 after removal
  assert.equal(resolveMoveIndex(true, 0, 3), 2);
  // drag a (0) to the end (4) -> 3
  assert.equal(resolveMoveIndex(true, 0, 4), 3);
});

test("moving an item upwards keeps the index", () => {
  assert.equal(resolveMoveIndex(true, 3, 0), 0);
  assert.equal(resolveMoveIndex(true, 3, 1), 1);
});

test("dropping an item on its own position is a no-op", () => {
  assert.equal(resolveMoveIndex(true, 1, 1), null); // before itself
  assert.equal(resolveMoveIndex(true, 1, 2), null); // after itself
});

test("moving to another list uses the index as-is", () => {
  assert.equal(resolveMoveIndex(false, 1, 1), 1);
  assert.equal(resolveMoveIndex(false, 0, 5), 5);
});
