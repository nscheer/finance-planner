<script lang="ts">
  /**
   * A collapsible category with its entries. Handles the drag & drop targets
   * for both category reordering (drop above/below the group) and entry
   * moves (drop between rows, or on the header/empty body to append).
   */
  import Icon from "./Icon.svelte";
  import EntryRow from "./EntryRow.svelte";
  import { app, type CategoryView, openDialog, setCollapsed, confirmDeleteCategory } from "../lib/store.svelte";
  import { t, formatEuro, formatPercent } from "../lib/i18n.svelte";
  import {
    dnd,
    startDrag,
    endDrag,
    hoverCategoryTarget,
    hoverEntryTarget,
    isLowerHalf,
    isEntryTarget,
    isCategoryHighlighted,
    isDraggedCategory,
  } from "../lib/dnd.svelte";

  let {
    category,
    index,
    share = null,
    draggable = true,
  }: {
    category: CategoryView;
    index: number;
    /** Share of all spending (0..1) for spending categories, null for income. */
    share?: number | null;
    /** false while a filter is active: no drag & drop. */
    draggable?: boolean;
  } = $props();

  let groupEl: HTMLElement;
  const entries = $derived(category.entries ?? []);
  const count = $derived(entries.length);

  function onDragStart(event: DragEvent) {
    if (!draggable) {
      event.preventDefault();
      return;
    }
    startDrag(event, { type: "category", id: category.id, kind: category.kind, index });
  }

  /**
   * Category drags target the group as a whole (insert before/after), entry
   * drags that no row handled (header, empty body, collapsed group) append
   * to this category.
   */
  function onGroupDragOver(event: DragEvent) {
    if (dnd.source?.type === "category") {
      hoverCategoryTarget(event, category.kind, isLowerHalf(event, groupEl) ? index + 1 : index);
    } else if (dnd.source?.type === "entry" && !event.defaultPrevented) {
      hoverEntryTarget(event, category, count);
    }
  }

  function toggle() {
    setCollapsed(category.id, !category.collapsed);
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<section
  class="group"
  id="category-{category.id}"
  class:collapsed={category.collapsed}
  class:dragging={isDraggedCategory(category)}
  class:drop-target={isCategoryHighlighted(category.id)}
  bind:this={groupEl}
  ondragover={onGroupDragOver}
>
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <header class="head" draggable="true" ondragstart={onDragStart} ondragend={endDrag}>
    <span class="handle" title={t("category.dragHint")}><Icon name="grip" size={14} /></span>
    <button class="title" type="button" onclick={toggle} aria-expanded={!category.collapsed}>
      <span class="chevron"><Icon name="chevron" size={14} /></span>
      <span class="name">{category.name}</span>
      <span class="count">{count}</span>
    </button>
    <span class="share-cell">
      {#if share !== null}
        <span
          class="share"
          title={category.shareOfIncome > 0
            ? t("category.share", { percent: formatPercent(share), income: formatPercent(category.shareOfIncome) })
            : t("category.shareNoIncome", { percent: formatPercent(share) })}
        >
          <span class="share-bar"><span class="share-fill" style:width="{Math.round(share * 100)}%"></span></span>
          <span class="share-text">{formatPercent(share)}</span>
        </span>
      {/if}
    </span>
    <span></span>
    <span class="money subtotal">{formatEuro(category.monthlyCents)}</span>
    <span class="money subtotal">{formatEuro(category.yearlyCents)}</span>
    <span class="actions">
      <button class="icon-btn" type="button" title={t("category.addEntry")} aria-label="{t('category.addEntry')}: {category.name}" onclick={() => openDialog({ type: "entry", kind: category.kind, categoryId: category.id })}><Icon name="plus" /></button>
      <button class="icon-btn" type="button" title={t("category.rename")} aria-label="{t('category.rename')}: {category.name}" onclick={() => openDialog({ type: "category", kind: category.kind, category })}><Icon name="edit" /></button>
      <button class="icon-btn danger" type="button" title={count > 0 ? t("category.inUse") : t("category.delete")} aria-label="{t('category.delete')}: {category.name}" onclick={() => confirmDeleteCategory(category, index)}><Icon name="trash" /></button>
    </span>
  </header>

  {#if !category.collapsed || app.printing}
    <div class="body" role="rowgroup">
      {#each entries as entry, i (entry.id)}
        <div class="drop-line" class:active={isEntryTarget(category.id, i)}></div>
        <EntryRow {entry} {category} index={i} {draggable} />
      {/each}
      <div class="drop-line" class:active={isEntryTarget(category.id, count)}></div>
      {#if count === 0}
        <div class="empty" class:active={isEntryTarget(category.id, 0)}>
          {isCategoryHighlighted(category.id) ? t("category.dropHere") : t("category.noEntries")}
        </div>
      {/if}
    </div>
  {:else if isCategoryHighlighted(category.id)}
    <div class="drop-collapsed">{t("category.dropInto", { name: category.name })}</div>
  {/if}
</section>

<style>
  .group {
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface);
    overflow: hidden;
    transition: box-shadow 0.12s, border-color 0.12s;
  }
  .group.dragging {
    opacity: 0.4;
  }
  .group.drop-target {
    border-color: var(--accent);
    box-shadow: 0 0 0 2px var(--accent-soft);
  }
  .head {
    display: grid;
    grid-template-columns: var(--cols);
    align-items: center;
    min-height: 40px;
    /* 3 px accent bar in the block color; left padding reduced by 3 px so
       the header columns stay aligned with the rows below. */
    padding: 0 8px 0 1px;
    border-left: 3px solid var(--block-color);
    background: var(--surface-3);
    cursor: grab;
  }
  .head:hover .actions {
    opacity: 1;
  }
  .handle {
    display: flex;
    justify-content: center;
    color: var(--text-3);
  }
  .title {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    padding: 6px 4px;
    border: 0;
    background: none;
    text-align: left;
    cursor: pointer;
    font-weight: 700;
    font-size: 15px;
    color: var(--text);
  }
  .chevron {
    display: flex;
    color: var(--text-3);
    transition: transform 0.15s;
    transform: rotate(90deg);
  }
  .collapsed .chevron {
    transform: rotate(0deg);
  }
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .count {
    padding: 0 7px;
    border-radius: 999px;
    background: var(--surface);
    border: 1px solid var(--border);
    font-size: 11px;
    font-weight: 600;
    color: var(--text-2);
  }
  .share-cell {
    display: flex;
    align-items: center;
  }
  .share {
    display: flex;
    align-items: center;
    gap: 6px;
    width: 100%;
    padding-right: 24px; /* clear gap to the due column */
  }
  .share-bar {
    flex: 1;
    height: 5px;
    border-radius: 3px;
    background: var(--border);
    overflow: hidden;
  }
  .share-fill {
    display: block;
    height: 100%;
    border-radius: 3px;
    background: var(--spending);
  }
  .share-text {
    font-size: 11px;
    color: var(--text-2);
    min-width: 32px;
    text-align: right;
  }
  .subtotal {
    text-align: right;
    padding-right: 12px;
    font-weight: 600;
    color: var(--text-2);
  }
  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 2px;
    opacity: 0;
    transition: opacity 0.1s;
  }
  .body {
    position: relative;
  }
  /* The insertion indicator between rows. It has no height of its own and
     the active line is an absolutely positioned overlay, so showing or
     moving it never changes the layout (no flicker while dragging). */
  .drop-line {
    position: relative;
    height: 0;
  }
  .drop-line.active::before {
    content: "";
    position: absolute;
    left: 6px;
    right: 6px;
    top: -1.5px;
    height: 3px;
    border-radius: 2px;
    background: var(--accent);
    z-index: 1;
    pointer-events: none;
  }
  .empty {
    padding: 10px 14px;
    border-top: 1px solid var(--border);
    color: var(--text-3);
    font-style: italic;
  }
  .empty.active {
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-style: normal;
    font-weight: 500;
  }
  .drop-collapsed {
    padding: 8px 14px;
    border-top: 1px solid var(--accent);
    background: var(--accent-soft);
    color: var(--accent-strong);
    font-weight: 500;
  }
</style>
