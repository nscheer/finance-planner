<script lang="ts">
  /**
   * Add, edit or duplicate an income/spending entry: name, amount in €, the
   * period the amount is paid in (the "master" period), the due month for
   * non-monthly periods, category, notes and the paused flag.
   */
  import Modal from "./Modal.svelte";
  import {
    Service,
    Kind,
    Period,
    type EntryView,
    apply,
    categoriesOf,
    closeDialog,
    errorMessage,
    kindKey,
    notify,
    openDialog,
    periodMonths,
  } from "../lib/store.svelte";
  import { t, formatEuro, monthName } from "../lib/i18n.svelte";
  import { centsToInput, parseEuro } from "../lib/money";
  import { untrack } from "svelte";

  let {
    kind,
    entry,
    categoryId,
    duplicateOf,
  }: { kind: Kind; entry?: EntryView; categoryId?: string; duplicateOf?: EntryView } = $props();

  const categories = $derived(categoriesOf(kind));

  // The dialog is created fresh each time it opens, so the form is seeded
  // once from the initial props (untrack: no reactive dependency intended).
  const initial = untrack(() => ({ entry, categoryId, duplicateOf }));
  const isEdit = !!initial.entry;
  const source = initial.entry ?? initial.duplicateOf;

  let name = $state(initial.entry?.name ?? (initial.duplicateOf ? initial.duplicateOf.name + t("entryDialog.copySuffix") : ""));
  let amount = $state(source ? centsToInput(source.amountCents) : "");
  let period = $state<Period>(source?.period ?? Period.PeriodMonthly);
  let dueMonth = $state(source?.dueMonth ?? 0);
  let notes = $state(source?.notes ?? "");
  let paused = $state(source?.paused ?? false);
  let selectedCategory = $state(source?.categoryId ?? initial.categoryId ?? "");
  let error = $state("");
  let working = $state(false);

  // Default to the first category when none was preselected.
  $effect(() => {
    if (!selectedCategory && categories.length > 0) selectedCategory = categories[0].id;
  });

  const periods: { value: Period; label: "entryDialog.perMonth" | "entryDialog.perQuarter" | "entryDialog.perHalfYear" | "entryDialog.perYear" }[] = [
    { value: Period.PeriodMonthly, label: "entryDialog.perMonth" },
    { value: Period.PeriodQuarterly, label: "entryDialog.perQuarter" },
    { value: Period.PeriodHalfYearly, label: "entryDialog.perHalfYear" },
    { value: Period.PeriodYearly, label: "entryDialog.perYear" },
  ];
  const months = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12];

  /** Live preview of the derived values while typing. */
  const preview = $derived.by(() => {
    const cents = parseEuro(amount);
    if (cents === null || cents <= 0) return "";
    const n = periodMonths(period);
    return t("entryDialog.preview", {
      monthly: formatEuro(Math.round(cents / n)),
      yearly: formatEuro((cents * 12) / n),
    });
  });

  const title = $derived(
    t(kindKey(isEdit ? "entryDialog.titleEdit" : initial.duplicateOf ? "entryDialog.titleDuplicate" : "entryDialog.titleNew", kind)),
  );

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    const cents = parseEuro(amount);
    if (cents === null) {
      error = t("entryDialog.invalidAmount");
      return;
    }
    const input = {
      categoryId: selectedCategory,
      name,
      amountCents: cents,
      period,
      dueMonth: period === Period.PeriodMonthly ? 0 : dueMonth,
      paused,
      notes,
    };
    working = true;
    try {
      if (initial.entry) {
        await apply(Service.UpdateEntry(initial.entry.id, input));
        notify("success", t("toast.entrySaved", { name: name.trim() }));
      } else {
        await apply(Service.AddEntry(input));
        notify("success", t("toast.entryAdded", { name: name.trim() }));
      }
      closeDialog();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      working = false;
    }
  }

  function addCategoryFirst() {
    openDialog({ type: "category", kind, returnToEntry: true });
  }
</script>

<Modal {title} width={520} onclose={closeDialog}>
  {#if categories.length === 0}
    <p class="empty">{t(kindKey("entryDialog.noCategories", kind))}</p>
    <div class="empty-actions">
      <button class="btn btn-primary" type="button" onclick={addCategoryFirst}>{t("entryDialog.addCategoryFirst")}</button>
    </div>
  {:else}
    <form id="entry-form" onsubmit={submit}>
      {#if error}<p class="form-error">{error}</p>{/if}
      <div class="field">
        <label for="entry-name">{t("entryDialog.name")}</label>
        <input id="entry-name" class="input" type="text" bind:value={name} placeholder={t("entryDialog.namePlaceholder")} autocomplete="off" />
      </div>
      <div class="field">
        <label for="entry-amount">{t("entryDialog.amount")}</label>
        <div class="row">
          <div class="input-suffix grow">
            <input id="entry-amount" class="input" type="text" inputmode="decimal" bind:value={amount} placeholder="0,00" autocomplete="off" />
            <span>€</span>
          </div>
          <div class="segmented" role="radiogroup" aria-label={t("entryDialog.paid")}>
            {#each periods as p (p.value)}
              <button type="button" class:active={period === p.value} onclick={() => (period = p.value)}>{t(p.label)}</button>
            {/each}
          </div>
        </div>
        <span class="hint">{preview || " "}</span>
      </div>
      <div class="row">
        <div class="field grow">
          <label for="entry-category">{t("entryDialog.category")}</label>
          <select id="entry-category" class="select" bind:value={selectedCategory}>
            {#each categories as c (c.id)}
              <option value={c.id}>{c.name}</option>
            {/each}
          </select>
        </div>
        {#if period !== Period.PeriodMonthly}
          <div class="field grow">
            <label for="entry-due">{t("entryDialog.dueMonth")}</label>
            <select id="entry-due" class="select" bind:value={dueMonth}>
              <option value={0}>{t("entryDialog.dueMonthNone")}</option>
              {#each months as m (m)}
                <option value={m}>{monthName(m)}</option>
              {/each}
            </select>
            <span class="hint">{t("entryDialog.dueMonthHint", { months: periodMonths(period) })}</span>
          </div>
        {/if}
      </div>
      <div class="field">
        <label for="entry-notes">{t("entryDialog.notes")}</label>
        <textarea id="entry-notes" class="input" rows="2" bind:value={notes} placeholder={t("entryDialog.notesPlaceholder")}></textarea>
      </div>
      <label class="check">
        <input type="checkbox" bind:checked={paused} />
        <span>{t("entryDialog.paused")}</span>
      </label>
    </form>
  {/if}
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>{t("dialog.cancel")}</button>
    {#if categories.length > 0}
      <button class="btn btn-primary" type="submit" form="entry-form" disabled={working}>
        {isEdit ? t("dialog.save") : t("dialog.add")}
      </button>
    {/if}
  {/snippet}
</Modal>

<style>
  .row {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }
  .grow {
    flex: 1;
    min-width: 0;
  }
  .segmented {
    flex-shrink: 0;
  }
  .segmented button {
    padding: 7px 10px;
    font-size: 13px;
  }
  textarea.input {
    resize: vertical;
    min-height: 40px;
    font-family: inherit;
  }
  .check {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    margin: 2px 0 14px;
    color: var(--text-2);
    cursor: pointer;
  }
  .check input {
    margin-top: 3px;
  }
  .empty {
    margin: 4px 0 12px;
    color: var(--text-2);
  }
  .empty-actions {
    margin-bottom: 12px;
  }
</style>
