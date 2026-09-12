<script lang="ts">
  /**
   * Add or edit an income/spending entry: name, amount in €, whether the
   * amount is per month or per year (the "master" period) and the category.
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
    kindLabel,
    notify,
    openDialog,
  } from "../lib/store.svelte";
  import { centsToInput, formatEuro, parseEuro } from "../lib/money";
  import { untrack } from "svelte";

  let { kind, entry, categoryId }: { kind: Kind; entry?: EntryView; categoryId?: string } = $props();

  const categories = $derived(categoriesOf(kind));

  // The dialog is created fresh each time it opens, so the form is seeded
  // once from the initial props (untrack: no reactive dependency intended).
  const initial = untrack(() => ({ entry, categoryId }));
  const isEdit = !!initial.entry;

  let name = $state(initial.entry?.name ?? "");
  let amount = $state(initial.entry ? centsToInput(initial.entry.amountCents) : "");
  let period = $state<Period>(initial.entry?.period ?? Period.PeriodMonthly);
  let selectedCategory = $state(initial.entry?.categoryId ?? initial.categoryId ?? "");
  let error = $state("");
  let working = $state(false);

  // Default to the first category when none was preselected.
  $effect(() => {
    if (!selectedCategory && categories.length > 0) selectedCategory = categories[0].id;
  });

  /** Live preview of the derived value while typing. */
  const preview = $derived.by(() => {
    const cents = parseEuro(amount);
    if (cents === null || cents <= 0) return "";
    return period === Period.PeriodMonthly
      ? `= ${formatEuro(cents * 12)} per year`
      : `= ${formatEuro(Math.round(cents / 12))} per month`;
  });

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    const cents = parseEuro(amount);
    if (cents === null) {
      error = "Please enter a valid amount, e.g. 12,50.";
      return;
    }
    working = true;
    try {
      if (initial.entry) {
        await apply(Service.UpdateEntry(initial.entry.id, selectedCategory, name, cents, period));
        notify("success", `Saved "${name.trim()}".`);
      } else {
        await apply(Service.AddEntry(selectedCategory, name, cents, period));
        notify("success", `Added "${name.trim()}".`);
      }
      closeDialog();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      working = false;
    }
  }

  function addCategoryFirst() {
    openDialog({ type: "category", kind });
  }
</script>

<Modal title={isEdit ? `Edit ${kindLabel(kind).toLowerCase()}` : `New ${kindLabel(kind).toLowerCase()}`} width={460} onclose={closeDialog}>
  {#if categories.length === 0}
    <p class="empty">
      There are no {kindLabel(kind).toLowerCase()} categories yet. Categories have to be added before entries can be entered.
    </p>
    <div class="empty-actions">
      <button class="btn btn-primary" type="button" onclick={addCategoryFirst}>Add a category first</button>
    </div>
  {:else}
    <form id="entry-form" onsubmit={submit}>
      {#if error}<p class="form-error">{error}</p>{/if}
      <div class="field">
        <label for="entry-name">Name</label>
        <input id="entry-name" class="input" type="text" bind:value={name} placeholder="e.g. Rent" autocomplete="off" />
      </div>
      <div class="row">
        <div class="field grow">
          <label for="entry-amount">Amount</label>
          <div class="input-suffix">
            <input id="entry-amount" class="input" type="text" inputmode="decimal" bind:value={amount} placeholder="0,00" autocomplete="off" />
            <span>€</span>
          </div>
          <span class="hint">{preview || " "}</span>
        </div>
        <div class="field">
          <label for="entry-period">Paid</label>
          <div class="segmented" id="entry-period" role="radiogroup" aria-label="Period">
            <button type="button" class:active={period === Period.PeriodMonthly} onclick={() => (period = Period.PeriodMonthly)}>per month</button>
            <button type="button" class:active={period === Period.PeriodYearly} onclick={() => (period = Period.PeriodYearly)}>per year</button>
          </div>
        </div>
      </div>
      <div class="field">
        <label for="entry-category">Category</label>
        <select id="entry-category" class="select" bind:value={selectedCategory}>
          {#each categories as c (c.id)}
            <option value={c.id}>{c.name}</option>
          {/each}
        </select>
      </div>
    </form>
  {/if}
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>Cancel</button>
    {#if categories.length > 0}
      <button class="btn btn-primary" type="submit" form="entry-form" disabled={working}>
        {isEdit ? "Save" : "Add"}
      </button>
    {/if}
  {/snippet}
</Modal>

<style>
  .row {
    display: flex;
    gap: 16px;
  }
  .grow {
    flex: 1;
  }
  .empty {
    margin: 4px 0 12px;
    color: var(--text-2);
  }
  .empty-actions {
    margin-bottom: 12px;
  }
</style>
