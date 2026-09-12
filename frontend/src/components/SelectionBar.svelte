<script lang="ts">
  /** Bulk actions for the selected entries; shown above the blocks. */
  import Icon from "./Icon.svelte";
  import {
    Kind,
    categoriesOf,
    selectionCount,
    selectionKind,
    clearSelection,
    moveSelection,
    pauseSelection,
    confirmDeleteSelection,
  } from "../lib/store.svelte";
  import { t, plural } from "../lib/i18n.svelte";

  const count = $derived(selectionCount());
  const kind = $derived(selectionKind());
  const targets = $derived(kind === Kind.KindIncome || kind === Kind.KindSpending ? categoriesOf(kind) : []);
  let target = $state("");
</script>

<div class="selection-bar" role="toolbar" aria-label={plural("selection.count", count)}>
  <span class="count">{plural("selection.count", count)}</span>
  <span class="divider"></span>
  <select class="select select-sm" bind:value={target} disabled={kind === "mixed"} title={kind === "mixed" ? t("selection.mixed") : undefined} aria-label={t("selection.moveTo")}>
    <option value="">{t("selection.moveTo")}</option>
    {#each targets as c (c.id)}
      <option value={c.id}>{c.name}</option>
    {/each}
  </select>
  <button class="btn btn-sm" type="button" disabled={!target || kind === "mixed"} onclick={() => { const id = target; target = ""; moveSelection(id); }}>{t("selection.move")}</button>
  <span class="divider"></span>
  <button class="btn btn-sm" type="button" onclick={() => pauseSelection(true)}><Icon name="pause" size={14} /> {t("selection.pause")}</button>
  <button class="btn btn-sm" type="button" onclick={() => pauseSelection(false)}><Icon name="play" size={14} /> {t("selection.resume")}</button>
  <button class="btn btn-sm btn-danger" type="button" onclick={confirmDeleteSelection}><Icon name="trash" size={14} /> {t("selection.delete")}</button>
  <span class="grow"></span>
  <button class="btn btn-sm" type="button" onclick={clearSelection}><Icon name="close" size={14} /> {t("selection.clear")}</button>
</div>

<style>
  /* Floats over the content at the bottom center, so selecting an entry
     never moves the table. Dialogs (z-index 100) stay above it. */
  .selection-bar {
    position: fixed;
    left: 50%;
    bottom: 20px;
    z-index: 90;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 8px;
    max-width: min(760px, calc(100vw - 40px));
    padding: 8px 12px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface);
    box-shadow: var(--shadow-lg);
    animation: rise 0.16s ease-out;
  }
  .count {
    padding: 2px 10px;
    border-radius: 999px;
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-weight: 600;
    white-space: nowrap;
  }
  @keyframes rise {
    from {
      transform: translate(-50%, 8px);
      opacity: 0;
    }
  }
  .divider {
    width: 1px;
    height: 20px;
    background: var(--border-strong);
  }
  .grow {
    flex: 1;
  }
</style>
