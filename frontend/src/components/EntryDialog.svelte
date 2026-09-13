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
    recentFor,
    rememberEntryDefaults,
  } from "../lib/store.svelte";
  import { pickCategory } from "../lib/entryDefaults";
  import { t, formatEuro, monthName, amountInput, amountPlaceholder } from "../lib/i18n.svelte";
  import { AmountField } from "../lib/validate.svelte";
  import { onDestroy, untrack } from "svelte";

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

  // Where the last entry of this kind went; a new entry starts there.
  const remembered = untrack(() => recentFor(kind));

  let name = $state(initial.entry?.name ?? (initial.duplicateOf ? initial.duplicateOf.name + t("entryDialog.copySuffix") : ""));
  let nameEl: HTMLInputElement | undefined = $state();
  const amount = new AmountField(source ? amountInput(source.amountCents) : "");
  let amountEl: HTMLInputElement | undefined = $state();
  onDestroy(() => amount.dispose());
  let period = $state<Period>(source?.period ?? remembered.period);
  let dueMonth = $state(source?.dueMonth ?? 0);
  let notes = $state(source?.notes ?? "");
  let paused = $state(source?.paused ?? false);
  let selectedCategory = $state(source?.categoryId ?? initial.categoryId ?? "");
  let error = $state("");
  let working = $state(false);

  // Nothing preselected: continue in the category last used, else the first
  // one. Runs again when a category is created from inside this dialog.
  $effect(() => {
    if (!selectedCategory && categories.length > 0) {
      selectedCategory = pickCategory(categories, remembered.categoryId);
    }
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
    const cents = amount.cents;
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

  /**
   * Saves the entry. With keepOpen the dialog stays open for the next one:
   * category, period and due month are kept, everything that belongs to the
   * single entry is cleared and the cursor returns to the name field.
   */
  async function save(keepOpen: boolean) {
    error = "";
    amount.touch();
    const cents = amount.cents;
    if (cents === null) {
      amountEl?.focus();
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
      rememberEntryDefaults(kind, selectedCategory, period);
      if (keepOpen && !initial.entry) {
        name = "";
        amount.reset();
        notes = "";
        paused = false;
        nameEl?.focus();
        return;
      }
      closeDialog();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      working = false;
    }
  }

  /** Ctrl/Cmd+Enter saves; while adding it goes straight to the next entry. */
  function onFormKeydown(event: KeyboardEvent) {
    if ((event.ctrlKey || event.metaKey) && event.key === "Enter") {
      event.preventDefault();
      save(!isEdit);
    }
  }

  /** Arrow keys move the period selection; the group is one tab stop. */
  let periodEls: (HTMLButtonElement | undefined)[] = [];
  function onPeriodKeydown(event: KeyboardEvent) {
    const current = periods.findIndex((p) => p.value === period);
    let next: number;
    switch (event.key) {
      case "ArrowRight":
      case "ArrowDown":
        next = (current + 1) % periods.length;
        break;
      case "ArrowLeft":
      case "ArrowUp":
        next = (current - 1 + periods.length) % periods.length;
        break;
      case "Home":
        next = 0;
        break;
      case "End":
        next = periods.length - 1;
        break;
      default:
        return;
    }
    event.preventDefault();
    period = periods[next].value;
    periodEls[next]?.focus();
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
    <!-- The keydown carries the Ctrl+Enter shortcut for the whole form. -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <form id="entry-form" onsubmit={(e) => { e.preventDefault(); save(false); }} onkeydown={onFormKeydown}>
      {#if error}<p class="form-error">{error}</p>{/if}
      <div class="field">
        <label for="entry-name">{t("entryDialog.name")}</label>
        <input id="entry-name" class="input" type="text" bind:value={name} bind:this={nameEl} placeholder={t("entryDialog.namePlaceholder")} autocomplete="off" />
      </div>
      <div class="field">
        <label for="entry-amount">{t("entryDialog.amount")}</label>
        <div class="row">
          <div class="input-suffix grow">
            <input
              id="entry-amount"
              class="input"
              class:invalid={!!amount.error}
              type="text"
              inputmode="decimal"
              bind:value={amount.value}
              bind:this={amountEl}
              oninput={() => amount.changed()}
              onblur={() => amount.touch()}
              aria-invalid={!!amount.error}
              placeholder={amountPlaceholder()}
              autocomplete="off"
            />
            <span>€</span>
          </div>
          <!-- tabindex="-1": the group itself is skipped, the selected radio is the tab stop. -->
          <div class="segmented" role="radiogroup" tabindex="-1" aria-label={t("entryDialog.paid")} onkeydown={onPeriodKeydown}>
            {#each periods as p, i (p.value)}
              <button
                type="button"
                role="radio"
                aria-checked={period === p.value}
                tabindex={period === p.value ? 0 : -1}
                class:active={period === p.value}
                bind:this={periodEls[i]}
                onclick={() => (period = p.value)}
              >{t(p.label)}</button>
            {/each}
          </div>
        </div>
        <span class="hint" class:error={!!amount.error}>{amount.error || preview || " "}</span>
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
      {#if !isEdit}
        <button class="btn btn-tonal" type="button" disabled={working} title={t("entryDialog.saveAndNextTip")} onclick={() => save(true)}>
          {t("entryDialog.saveAndNext")}
        </button>
      {/if}
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
