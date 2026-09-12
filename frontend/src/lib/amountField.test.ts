import { test } from "node:test";
import assert from "node:assert/strict";
import { amountFieldState } from "./amountField.ts";

test("entry amounts: must be a number greater than zero", () => {
  assert.deepEqual(amountFieldState("12,50", ",", false), { cents: 1250, errorKey: null });
  assert.deepEqual(amountFieldState("12.50", ".", false), { cents: 1250, errorKey: null });
  assert.equal(amountFieldState("", ",", false).errorKey, "entryDialog.invalidAmount");
  assert.equal(amountFieldState("abc", ",", false).errorKey, "entryDialog.invalidAmount");
  assert.equal(amountFieldState("10.123", ".", false).errorKey, "entryDialog.invalidAmount");
  assert.equal(amountFieldState("0", ",", false).errorKey, "errors.entry.amountPositive");
  assert.equal(amountFieldState("-5", ",", false).errorKey, "errors.entry.amountPositive");
});

test("savings goal: empty or zero removes the goal, negative is refused", () => {
  assert.deepEqual(amountFieldState("", ",", true), { cents: 0, errorKey: null });
  assert.deepEqual(amountFieldState("0", ",", true), { cents: 0, errorKey: null });
  assert.deepEqual(amountFieldState("300", ".", true), { cents: 30000, errorKey: null });
  assert.equal(amountFieldState("-1", ",", true).errorKey, "errors.settings.savingsGoalNegative");
  assert.equal(amountFieldState("x", ",", true).errorKey, "entryDialog.invalidAmount");
});
