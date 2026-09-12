import { test } from "node:test";
import assert from "node:assert/strict";
import { parseEuro, centsToInput, formatEuroIn, formatEuroSignedIn } from "./money.ts";

test("German decimal mark: comma is decimal, point groups thousands", () => {
  assert.equal(parseEuro("12"), 1200);
  assert.equal(parseEuro("12,5"), 1250);
  assert.equal(parseEuro("0,01"), 1);
  assert.equal(parseEuro(",5"), 50);
  assert.equal(parseEuro(" 12,50 € "), 1250);
  assert.equal(parseEuro("1.234,56"), 123456);
  assert.equal(parseEuro("1 234,56"), 123456);
  assert.equal(parseEuro("10.123"), 1012300);
  assert.equal(parseEuro("1.234.567"), 123456700);
  assert.equal(parseEuro("12.50"), null); // point is not a decimal mark in German
});

test("English decimal mark: point is decimal, comma groups thousands", () => {
  assert.equal(parseEuro("12.5", "."), 1250);
  assert.equal(parseEuro("10.12", "."), 1012);
  assert.equal(parseEuro("1,234.56", "."), 123456);
  assert.equal(parseEuro("10,123", "."), 1012300);
  assert.equal(parseEuro("10.123", "."), null); // three decimals
  assert.equal(parseEuro("12,50", "."), null); // comma is not a decimal mark in English
});

test("rejects invalid input", () => {
  assert.equal(parseEuro(""), null);
  assert.equal(parseEuro("abc"), null);
  assert.equal(parseEuro("12,3456"), null); // more than 2 decimals
  assert.equal(parseEuro("1,2,3"), null);
  assert.equal(parseEuro("12,5x"), null);
  assert.equal(parseEuro("1.23,45"), null); // bad grouping
  assert.equal(parseEuro("1234.567,89"), null); // bad grouping
});

test("negative amounts keep their sign", () => {
  assert.equal(parseEuro("-3,50"), -350);
  assert.equal(parseEuro("-3.50", "."), -350);
});

test("round trip between input text and cents", () => {
  for (const cents of [0, 1, 99, 100, 123456, 100000000]) {
    assert.equal(parseEuro(centsToInput(cents)), cents);
  }
  assert.equal(centsToInput(1250), "12,50");
  assert.equal(centsToInput(5), "0,05");
  assert.equal(centsToInput(1250, "."), "12.50");
  assert.equal(parseEuro(centsToInput(123456, "."), "."), 123456);
});

test("formats euro amounts per locale", () => {
  // Intl uses a non-breaking space before the € sign.
  const nbsp = (s: string) => s.replace(/\u00a0/g, " ");
  assert.equal(nbsp(formatEuroIn(123456, "de-DE")), "1.234,56 €");
  assert.equal(nbsp(formatEuroIn(0, "de-DE")), "0,00 €");
  assert.equal(formatEuroIn(123456, "en-IE"), "€1,234.56");
  assert.equal(nbsp(formatEuroSignedIn(-350, "de-DE")), "-3,50 €");
  assert.equal(nbsp(formatEuroSignedIn(350, "de-DE")), "+3,50 €");
});
