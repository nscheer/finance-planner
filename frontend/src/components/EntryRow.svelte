<script lang="ts">
  /**
   * One table row. The amount the user entered (the "master") is shown bold
   * with a period badge, the derived amount is shown muted next to it.
   * Paused entries are greyed out and carry a "paused" badge.
   */
  import Icon from "./Icon.svelte";
  import {
    Period,
    type CategoryView,
    type EntryView,
    openDialog,
    confirmDeleteEntry,
    pauseEntry,
    periodKey,
  } from "../lib/store.svelte";
  import { t, formatEuro, monthShort } from "../lib/i18n.svelte";
  import { startDrag, endDrag, hoverEntryTarget, isLowerHalf, isDraggedEntry, dnd } from "../lib/dnd.svelte";

  let {
    entry,
    category,
    index,
    draggable = true,
  }: { entry: EntryView; category: CategoryView; index: number; draggable?: boolean } = $props();

  let rowEl: HTMLElement;
  const monthlyIsMaster = $derived(entry.period === Period.PeriodMonthly);

  function onDragStart(event: DragEvent) {
    if (!draggable) {
      event.preventDefault();
      return;
    }
    startDrag(event, { type: "entry", id: entry.id, kind: category.kind, categoryId: category.id, index });
  }

  function onDragOver(event: DragEvent) {
    if (dnd.source?.type !== "entry") return; // category drags are handled by the group
    hoverEntryTarget(event, category, isLowerHalf(event, rowEl) ? index + 1 : index);
  }

  function edit() {
    openDialog({ type: "entry", kind: category.kind, entry });
  }

  function duplicate() {
    openDialog({ type: "entry", kind: category.kind, duplicateOf: entry });
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="row"
  class:dragging={isDraggedEntry(entry)}
  class:paused={entry.paused}
  class:locked={!draggable}
  bind:this={rowEl}
  draggable="true"
  ondragstart={onDragStart}
  ondragend={endDrag}
  ondragover={onDragOver}
  ondblclick={edit}
>
  <span class="handle" title={draggable ? t("entry.dragHint") : ""}><Icon name="grip" size={14} /></span>
  <span class="name" title={entry.notes || undefined}>
    <span class="name-text">{entry.name}</span>
    {#if entry.notes}<span class="note-icon" aria-label={t("entry.notes")}><Icon name="note" size={13} /></span>{/if}
  </span>
  <span class="period">
    <span class="badge badge-{entry.period}">{t(periodKey(entry.period))}</span>
    {#if entry.paused}<span class="badge badge-paused">{t("entry.paused")}</span>{/if}
    {#if !monthlyIsMaster && (entry.dueMonth ?? 0) > 0}
      <span class="due">{t("entry.due", { month: monthShort(entry.dueMonth ?? 0) })}</span>
    {/if}
  </span>
  <span class="money amount" class:master={monthlyIsMaster} class:derived={!monthlyIsMaster}>{formatEuro(entry.monthlyCents)}</span>
  <span class="money amount" class:master={entry.period === Period.PeriodYearly} class:derived={entry.period !== Period.PeriodYearly}>{formatEuro(entry.yearlyCents)}</span>
  <span class="actions">
    <button class="icon-btn" type="button" title={entry.paused ? t("entry.resume") : t("entry.pause")} aria-label="{entry.paused ? t('entry.resume') : t('entry.pause')}: {entry.name}" onclick={() => pauseEntry(entry, !entry.paused)}><Icon name={entry.paused ? "play" : "pause"} /></button>
    <button class="icon-btn" type="button" title={t("entry.duplicate")} aria-label="{t('entry.duplicate')}: {entry.name}" onclick={duplicate}><Icon name="copy" /></button>
    <button class="icon-btn" type="button" title={t("entry.edit")} aria-label="{t('entry.edit')}: {entry.name}" onclick={edit}><Icon name="edit" /></button>
    <button class="icon-btn danger" type="button" title={t("entry.delete")} aria-label="{t('entry.delete')}: {entry.name}" onclick={() => confirmDeleteEntry(entry, index)}><Icon name="trash" /></button>
  </span>
</div>

<style>
  .row {
    display: grid;
    grid-template-columns: var(--cols);
    align-items: center;
    min-height: 38px;
    padding: 0 8px 0 4px;
    border-top: 1px solid var(--border);
    background: var(--surface);
    transition: background 0.1s;
  }
  .row:hover {
    background: var(--surface-2);
  }
  .row:hover .actions {
    opacity: 1;
  }
  .row.dragging {
    opacity: 0.35;
  }
  .row.paused .name,
  .row.paused .amount {
    color: var(--text-3);
    text-decoration: line-through;
    text-decoration-color: var(--border-strong);
  }
  .row.locked .handle {
    opacity: 0.3;
    cursor: default;
  }
  .handle {
    display: flex;
    justify-content: center;
    color: var(--text-3);
    cursor: grab;
  }
  .name {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    padding-left: 4px;
    font-weight: 500;
  }
  .name-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .note-icon {
    display: inline-flex;
    flex-shrink: 0;
    color: var(--text-3);
  }
  .period {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
  }
  .due {
    font-size: 11px;
    color: var(--text-3);
    white-space: nowrap;
  }
  .amount {
    text-align: right;
    padding-right: 12px;
  }
  .amount.master {
    font-weight: 600;
  }
  .amount.derived {
    color: var(--text-3);
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 2px;
    opacity: 0;
    transition: opacity 0.1s;
  }
</style>
