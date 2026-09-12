/**
 * The reference copies of the language files in harness/ are part of the
 * specification and must be byte-identical to the files used by the app.
 * See harness/specs.md, section 7.10.
 */
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";

const files = ["en.ts", "de.ts"];

test("language files in harness/ are in sync with frontend/src/i18n", () => {
  for (const name of files) {
    const app = readFileSync(new URL(`./${name}`, import.meta.url), "utf8");
    const reference = readFileSync(new URL(`../../../harness/${name}`, import.meta.url), "utf8");
    assert.equal(
      reference,
      app,
      `harness/${name} differs from frontend/src/i18n/${name} – run: cp frontend/src/i18n/${name} harness/`,
    );
  }
});
