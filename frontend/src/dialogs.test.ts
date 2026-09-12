/**
 * Regression test for the "dialog closes itself, then reads its props" bug:
 * Svelte 5 compiles child props to getters that re-evaluate the parent's
 * expression. If App.svelte passed `app.dialog.preview` directly, the getter
 * would throw after closeDialog() set app.dialog to null, and the import and
 * delete confirmations would silently do nothing.
 */
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { compile } from "svelte/compiler";

test("dialog props in App.svelte are bound to a local constant, not app.dialog", () => {
  const source = readFileSync(new URL("./App.svelte", import.meta.url), "utf8");
  const { js } = compile(source, { generate: "client", filename: "App.svelte" });
  const getters = [...js.code.matchAll(/get \w+\(\) \{\s*return ([^;]+);/g)].map((m) => m[1].trim());
  assert.ok(getters.length > 0, "expected dialog prop getters");
  for (const expr of getters) {
    assert.ok(!expr.startsWith("app.dialog"), `prop getter reads ${expr} which is null after the dialog closes`);
  }
});
