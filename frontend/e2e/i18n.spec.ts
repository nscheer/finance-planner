/**
 * The second language (specification 5.1 and 7.5).
 *
 * The rest of the suite runs in English, the default. Here the whole
 * application is driven in German, because the bug this catches is a string
 * that does not switch: a t() call with the wrong key, or a literal that
 * never went through the language files at all. The unit test in
 * src/i18n/i18n.test.ts proves both files define the same keys — it cannot
 * see a key that is never used, or one used in the wrong place.
 */
import { test, expect, type Page } from "@playwright/test";
import { openApp, nextDialog, row, block, settled, de, en } from "./app.ts";

const german = { lang: "de", sample: true } as const;

/** The dialog that is currently open. */
const dialog = (page: Page) => page.getByRole("dialog");

test("the entry dialog is German", async ({ page }) => {
  await openApp(page, german);
  await block(page, "spending", "de").getByRole("button", { name: de["block.addEntry.spending"] }).click();

  const d = dialog(page);
  await expect(d).toHaveAttribute("aria-label", de["entryDialog.titleNew.spending"]);
  await expect(d.getByText(de["entryDialog.name"], { exact: true })).toBeVisible();
  await expect(d.getByText(de["entryDialog.amount"], { exact: true })).toBeVisible();
  await expect(d.getByRole("button", { name: de["dialog.add"], exact: true })).toBeVisible();
  await expect(d.getByRole("button", { name: de["entryDialog.saveAndNext"] })).toBeVisible();
});

test("a coded error from the backend is German", async ({ page }) => {
  await openApp(page, german);
  await page.keyboard.press("c");

  const d = dialog(page);
  await expect(d).toHaveAttribute("aria-label", de["categoryDialog.titleNew.spending"]);
  await d.getByLabel(de["categoryDialog.name"]).fill("Wohnen"); // exists in the example plan
  await d.getByRole("button", { name: de["categoryDialog.submit"] }).click();

  // errors.category.exists, translated in the frontend from the code the
  // service returned; the kind inside the message is translated too.
  await expect(d.locator(".form-error")).toHaveText(
    de["errors.category.exists"].replace("{kind}", de["kind.spending"]).replace("{name}", "Wohnen"),
  );
});

test("the savings goal, backups and shortcut dialogs are German", async ({ page }) => {
  await openApp(page, german);

  await page.getByRole("button", { name: de["stats.goalEdit"] }).click();
  await expect(dialog(page)).toHaveAttribute("aria-label", de["goalDialog.title"]);
  await expect(dialog(page).getByText(de["goalDialog.hint"])).toBeVisible();
  await page.keyboard.press("Escape");

  await page.getByRole("button", { name: de["app.data"], exact: true }).click();
  await page.getByRole("menuitem", { name: de["app.backups"] }).click();
  await expect(dialog(page)).toHaveAttribute("aria-label", de["backups.title"]);
  await expect(dialog(page).getByText(de["backups.intro"])).toBeVisible();
  await page.keyboard.press("Escape");

  await page.keyboard.press("?");
  await expect(dialog(page)).toHaveAttribute("aria-label", de["shortcuts.title"]);
});

test("the command palette and the delete confirmation are German", async ({ page }) => {
  await openApp(page, german);
  await page.keyboard.press("Control+k");
  await expect(page.getByRole("textbox", { name: de["palette.placeholder"] })).toBeVisible();
  await expect(page.getByText(de["palette.hint"])).toBeVisible();
  await page.keyboard.press("Escape");

  await row(page, "Strom").getByRole("button", { name: `${de["entry.delete"]}: Strom` }).click();
  await expect(dialog(page)).toHaveAttribute("aria-label", de["confirm.deleteEntry.title"]);
  await expect(dialog(page).getByRole("button", { name: de["dialog.delete"] })).toBeVisible();
});

test("the import preview is German", async ({ page }, testInfo) => {
  await openApp(page, german);
  const file = testInfo.outputPath("export.json");

  await nextDialog(page, { path: file });
  await page.getByRole("button", { name: de["app.data"], exact: true }).click();
  await page.getByRole("menuitem", { name: de["app.export"], exact: true }).click();
  await expect(page.getByText(de["toast.exported"].split("{")[0].trim())).toBeVisible();

  await nextDialog(page, { path: file });
  await page.getByRole("button", { name: de["app.data"], exact: true }).click();
  await page.getByRole("menuitem", { name: de["app.import"], exact: true }).click();

  const d = dialog(page);
  await expect(d).toHaveAttribute("aria-label", de["importDialog.title"]);
  await expect(d.getByText(de["importDialog.question"])).toBeVisible();
  await expect(d.getByText(de["importDialog.merge.title"])).toBeVisible();
  await expect(d.getByText(de["importDialog.replace.title"])).toBeVisible();
});

/**
 * Texts that exist in one language only, so finding one on the other
 * language's page means it never switched. Dropped from the list:
 *
 *  - keys both languages spell the same way ("Name", "OK"),
 *  - values with a placeholder, which are interpolated before they appear,
 *  - values short enough to hide inside another word — "Data" sits in
 *    "data.json", and German compounds swallow short English words whole,
 *  - values that are part of their own counterpart, where the check cannot
 *    tell the two apart: "Import" is the start of "Importieren", "Januar"
 *    the start of "January".
 */
function onlyIn(source: typeof de, other: typeof de): string[] {
  return (Object.keys(source) as (keyof typeof de)[])
    .filter((k) => {
      const value = source[k];
      const counterpart = other[k];
      if (value === counterpart || value.includes("{") || value.length < 6) return false;
      return !counterpart.includes(value) && !value.includes(counterpart);
    })
    .map((k) => source[k]);
}

/**
 * Everything the user can read on the page: the rendered text plus the
 * attributes that carry copy. Without the attributes the tooltips — the
 * whole `tip.*` half of the language files — would never be looked at.
 */
async function visibleText(page: Page): Promise<string> {
  const attributes = await page.evaluate(() =>
    [...document.querySelectorAll("[title], [aria-label], [placeholder]")]
      .flatMap((el) => [el.getAttribute("title"), el.getAttribute("aria-label"), el.getAttribute("placeholder")])
      .filter(Boolean)
      .join("\n"),
  );
  return (await page.locator("body").innerText()) + "\n" + attributes;
}

test("no English text survives the switch to German", async ({ page }) => {
  await openApp(page, german);
  const text = await visibleText(page);

  const leftovers = onlyIn(en, de).filter((value) => text.includes(value));
  expect(leftovers, "these are English strings on a German page").toEqual([]);
});

test("no German text shows up in English", async ({ page }) => {
  await openApp(page, { sample: true });
  const text = await visibleText(page);

  const leftovers = onlyIn(de, en).filter((value) => text.includes(value));
  expect(leftovers, "these are German strings on an English page").toEqual([]);
});

test("the sweep also covers what is inside a dialog", async ({ page }) => {
  await openApp(page, german);
  await block(page, "spending", "de").getByRole("button", { name: de["block.addEntry.spending"] }).click();
  await settled(dialog(page));

  const text = await visibleText(page);
  const leftovers = onlyIn(en, de).filter((value) => text.includes(value));
  expect(leftovers, "these are English strings in a German dialog").toEqual([]);
});
