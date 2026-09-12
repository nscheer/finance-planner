import { test } from "node:test";
import assert from "node:assert/strict";
import { parseEuro, centsToInput, formatEuro } from "./money.ts";

test("parses plain and decimal amounts with comma or dot", () => {
  assert.equal(parseEuro("12"), 1200);
  assert.equal(parseEuro("12,5"), 1250);
  assert.equal(parseEuro("12.50"), 1250);
  assert.equal(parseEuro("0,01"), 1);
  assert.equal(parseEuro(",5"), 50);
  assert.equal(parseEuro(" 12,50 € "), 1250);
});

test("parses thousands separators", () => {
  assert.equal(parseEuro("1.234,56"), 123456);
  assert.equal(parseEuro("1,234.56"), 123456);
  assert.equal(parseEuro("1.234"), 123400);
  assert.equal(parseEuro("1.234.567"), 123456700);
  assert.equal(parseEuro("1 234,56"), 123456);
});

test("rejects invalid input", () => {
  assert.equal(parseEuro(""), null);
  assert.equal(parseEuro("abc"), null);
  assert.equal(parseEuro("12,3456"), null); // more than 2 decimals
  assert.equal(parseEuro("12,345"), 1234500); // exactly 3 digits = thousands separator
  assert.equal(parseEuro("1,2,3"), null);
  assert.equal(parseEuro("12,5x"), null);
});

test("negative amounts keep their sign", () => {
  assert.equal(parseEuro("-3,50"), -350);
});

test("round trip between input text and cents", () => {
  for (const cents of [0, 1, 99, 100, 123456, 100000000]) {
    assert.equal(parseEuro(centsToInput(cents)), cents);
  }
  assert.equal(centsToInput(1250), "12,50");
  assert.equal(centsToInput(5), "0,05");
});

test("formats as German euro amounts", () => {
  // Intl uses a non-breaking space before the € sign.
  assert.equal(formatEuro(123456).replace(/ /g, " "), "1.234,56 €");
  assert.equal(formatEuro(0).replace(/ /g, " "), "0,00 €");
});
