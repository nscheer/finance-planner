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
    periodMonths,
    app,
    isSelected,
    setSelected,
    toggleSelected,
    selectRange,
  } from "../lib/store.svelte";
  import { t, formatEuro, monthName } from "../lib/i18n.svelte";
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

  /** Ctrl/Cmd+click toggles the selection, Shift+click selects a range. */
  function onClick(event: MouseEvent) {
    if (event.shiftKey) {
      selectRange(category, entry.id);
      event.preventDefault();
    } else if (event.ctrlKey || event.metaKey) {
      toggleSelected(entry.id);
      event.preventDefault();
    }
  }

  /** Tooltips that explain the entered and the calculated amount. */
  const entered = $derived(t("tip.entryMaster", { amount: formatEuro(entry.amountCents), period: t(periodKey(entry.period)) }));
  const months = $derived(periodMonths(entry.period));
  const monthlyTip = $derived(
    monthlyIsMaster ? entered : t("tip.entryMonthlyDerived", { amount: formatEuro(entry.amountCents), period: t(periodKey(entry.period)), months }),
  );
  const yearlyTip = $derived(
    entry.period === Period.PeriodYearly
      ? entered
      : t("tip.entryYearlyDerived", { amount: formatEuro(entry.amountCents), period: t(periodKey(entry.period)), payments: 12 / months }),
  );

  const selected = $derived(isSelected(entry.id));
  const showCheckbox = $derived(app.selectMode);

  function duplicate() {
    openDialog({ type: "entry", kind: category.kind, duplicateOf: entry });
  }
</script>

<!-- Keyboard users select via the checkbox; the row click is a mouse shortcut. -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
  class="row"
  data-testid="entry-row"
  class:dragging={isDraggedEntry(entry)}
  class:paused={entry.paused}
  class:locked={!draggable}
  class:selected
  class:show-checkbox={showCheckbox}
  bind:this={rowEl}
  draggable="true"
  ondragstart={onDragStart}
  ondragend={endDrag}
  ondragover={onDragOver}
  ondblclick={edit}
  onclick={onClick}
>
  <span class="handle" title={draggable ? t("entry.dragHint") : ""}>
    <span class="grip"><Icon name="grip" size={14} /></span>
    <input
      class="check"
      type="checkbox"
      checked={selected}
      title={t("entry.select")}
      aria-label="{t('entry.select')}: {entry.name}"
      onclick={(e) => { e.stopPropagation(); if (e.shiftKey) selectRange(category, entry.id); else setSelected(entry.id, !selected); }}
      ondblclick={(e) => e.stopPropagation()}
    />
  </span>
  <span class="name" title={entry.notes || undefined}>
    <span class="name-text">{entry.name}</span>
    {#if entry.notes}<span class="name-icon" aria-label={t("entry.notes")} role="img"><Icon name="note" size={13} /></span>{/if}
  </span>
  <span class="period">
    <span class="badge badge-{entry.period}">{t(periodKey(entry.period))}</span>
  </span>
  <span class="due">
    {#if !monthlyIsMaster && (entry.dueMonth ?? 0) > 0}{monthName(entry.dueMonth ?? 0)}{/if}
  </span>
  <span class="money amount" class:master={monthlyIsMaster} class:derived={!monthlyIsMaster} title={monthlyTip}>{formatEuro(entry.monthlyCents)}</span>
  <span class="money amount" class:master={entry.period === Period.PeriodYearly} class:derived={entry.period !== Period.PeriodYearly} title={yearlyTip}>{formatEuro(entry.yearlyCents)}</span>
  <span class="actions">
    <button class="icon-btn" type="button" title={t("entry.edit")} aria-label="{t('entry.edit')}: {entry.name}" onclick={edit}><Icon name="edit" /></button>
    <button class="icon-btn" type="button" title={t("entry.duplicate")} aria-label="{t('entry.duplicate')}: {entry.name}" onclick={duplicate}><Icon name="copy" /></button>
    <button class="icon-btn" type="button" title={entry.paused ? t("entry.resume") : t("entry.pause")} aria-label="{entry.paused ? t('entry.resume') : t('entry.pause')}: {entry.name}" onclick={() => pauseEntry(entry, !entry.paused)}><Icon name={entry.paused ? "play" : "pause"} /></button>
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
  /* Paused: diagonal stripes over the row background, text and badge muted. */
  .row.paused {
    background-image: repeating-linear-gradient(
      135deg,
      var(--stripe) 0 3px,
      transparent 3px 10px
    );
  }
  .row.paused .name,
  .row.paused .amount,
  .row.paused .due {
    color: var(--text-3);
  }
  .row.paused .badge {
    background: var(--surface-3);
    color: var(--text-3);
  }
  .row.locked .handle {
    opacity: 0.3;
    cursor: default;
  }
  .handle {
    display: flex;
    justify-content: center;
    align-items: center;
    color: var(--text-3);
    cursor: grab;
  }
  /* The handle cell shows the grip, or a checkbox in selection mode. */
  .handle .check {
    display: none;
    margin: 0;
    cursor: pointer;
  }
  .row.show-checkbox .handle .grip {
    display: none;
  }
  .row.show-checkbox .handle .check {
    display: block;
  }
  .row.selected {
    background: var(--accent-soft);
  }
  .row.selected:hover {
    background: var(--accent-soft);
  }
  .name {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    padding-left: 4px;
    font-weight: 400;
  }
  .name-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .name-icon {
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
    color: var(--text-2);
    text-align: center;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
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
