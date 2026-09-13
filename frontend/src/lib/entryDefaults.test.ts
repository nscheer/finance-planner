import { test } from "node:test";
import assert from "node:assert/strict";
import { pickCategory } from "./entryDefaults.ts";

const cats = [{ id: "a" }, { id: "b" }, { id: "c" }];

test("a new entry starts in the category last used", () => {
  assert.equal(pickCategory(cats, "b"), "b");
});

test("falls back to the first category when the remembered one is gone", () => {
  assert.equal(pickCategory(cats, "deleted"), "a");
  assert.equal(pickCategory(cats, undefined), "a");
  assert.equal(pickCategory(cats, ""), "a");
});

test("no categories means no selection", () => {
  assert.equal(pickCategory([], "b"), "");
  assert.equal(pickCategory([]), "");
});
