import { test } from "node:test";
import assert from "node:assert/strict";
import { locales, interpolate, translate, translatePlural, placeholders, findLocale } from "./index.ts";
import { en } from "./en.ts";

test("every language defines exactly the English keys with the same placeholders", () => {
  const enKeys = Object.keys(en).sort();
  for (const locale of locales) {
    const keys = Object.keys(locale.messages).sort();
    assert.deepEqual(keys, enKeys, `${locale.code}: key set differs from en`);
    for (const key of enKeys) {
      const value = (locale.messages as Record<string, string>)[key];
      assert.equal(typeof value, "string", `${locale.code}: ${key} is not a string`);
      assert.notEqual(value.trim(), "", `${locale.code}: ${key} is empty`);
      assert.deepEqual(
        placeholders(value),
        placeholders((en as Record<string, string>)[key]),
        `${locale.code}: placeholders of ${key} differ from en`,
      );
    }
  }
});

test("plural keys come in one/other pairs", () => {
  for (const key of Object.keys(en)) {
    if (key.endsWith(".one")) {
      assert.ok(key.replace(/\.one$/, ".other") in en, `${key} has no .other form`);
    }
    if (key.endsWith(".other")) {
      assert.ok(key.replace(/\.other$/, ".one") in en, `${key} has no .one form`);
    }
  }
});

test("German and English are available and German is the default", () => {
  assert.deepEqual(
    locales.map((l) => l.code).sort(),
    ["de", "en"],
  );
  assert.equal(findLocale("en").label, "English");
  assert.equal(findLocale("de").label, "Deutsch");
  // No choice saved yet (empty / missing) or an unknown code -> German.
  assert.equal(findLocale("").code, "de");
  assert.equal(findLocale(undefined).code, "de");
  assert.equal(findLocale("xx").code, "de");
});

test("interpolation replaces placeholders and keeps unknown ones", () => {
  assert.equal(interpolate("Hello {name}, {count} items", { name: "Ann", count: 3 }), "Hello Ann, 3 items");
  assert.equal(interpolate("{missing} stays", {}), "{missing} stays");
  assert.equal(interpolate("no params"), "no params");
});

test("translate and plural forms", () => {
  const de = findLocale("de").messages;
  assert.equal(translate(de, "kind.income"), "Einnahmen");
  assert.equal(translate(de, "toast.entryAdded", { name: "Miete" }), "„Miete“ hinzugefügt.");
  assert.equal(translatePlural(en, "importDialog.entries", 1), "1 entry");
  assert.equal(translatePlural(en, "importDialog.entries", 2), "2 entries");
  assert.equal(translatePlural(de, "importDialog.entries", 0), "0 Einträge");
});
