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

<div class="bar" role="toolbar" aria-label={plural("selection.count", count)}>
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
  .bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border: 1px solid var(--accent);
    border-radius: var(--radius-sm);
    background: var(--accent-soft);
  }
  .count {
    font-weight: 600;
    white-space: nowrap;
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
