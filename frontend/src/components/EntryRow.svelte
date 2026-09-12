<script lang="ts">
  /**
   * One table row. The amount the user entered (the "master") is shown bold
   * with a period badge, the derived amount is shown muted next to it.
   */
  import Icon from "./Icon.svelte";
  import { Period, type CategoryView, type EntryView, openDialog, confirmDeleteEntry } from "../lib/store.svelte";
  import { t, formatEuro } from "../lib/i18n.svelte";
  import { startDrag, endDrag, hoverEntryTarget, isLowerHalf, isDraggedEntry, dnd } from "../lib/dnd.svelte";

  let { entry, category, index }: { entry: EntryView; category: CategoryView; index: number } = $props();

  let rowEl: HTMLElement;
  const monthlyIsMaster = $derived(entry.period === Period.PeriodMonthly);

  function onDragStart(event: DragEvent) {
    startDrag(event, { type: "entry", id: entry.id, kind: category.kind, categoryId: category.id, index });
  }

  function onDragOver(event: DragEvent) {
    if (dnd.source?.type !== "entry") return; // category drags are handled by the group
    hoverEntryTarget(event, category, isLowerHalf(event, rowEl) ? index + 1 : index);
  }

  function edit() {
    openDialog({ type: "entry", kind: category.kind, entry });
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="row"
  class:dragging={isDraggedEntry(entry)}
  bind:this={rowEl}
  draggable="true"
  ondragstart={onDragStart}
  ondragend={endDrag}
  ondragover={onDragOver}
  ondblclick={edit}
>
  <span class="handle" title={t("entry.dragHint")}><Icon name="grip" size={14} /></span>
  <span class="name">{entry.name}</span>
  <span class="period">
    <span class="badge {monthlyIsMaster ? 'badge-monthly' : 'badge-yearly'}">{t(monthlyIsMaster ? "entry.monthly" : "entry.yearly")}</span>
  </span>
  <span class="money amount" class:master={monthlyIsMaster} class:derived={!monthlyIsMaster}>{formatEuro(entry.monthlyCents)}</span>
  <span class="money amount" class:master={!monthlyIsMaster} class:derived={monthlyIsMaster}>{formatEuro(entry.yearlyCents)}</span>
  <span class="actions">
    <button class="icon-btn" type="button" title={t("entry.edit")} aria-label="{t('entry.edit')}: {entry.name}" onclick={edit}><Icon name="edit" /></button>
    <button class="icon-btn danger" type="button" title={t("entry.delete")} aria-label="{t('entry.delete')}: {entry.name}" onclick={() => confirmDeleteEntry(entry)}><Icon name="trash" /></button>
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
  .handle {
    display: flex;
    justify-content: center;
    color: var(--text-3);
    cursor: grab;
  }
  .name {
    padding-left: 4px;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
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
