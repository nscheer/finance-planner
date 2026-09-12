<script lang="ts">
  /**
   * One of the two main blocks (income / spending): header with totals and
   * actions, a column header and the list of categories.
   */
  import Icon from "./Icon.svelte";
  import CategoryGroup from "./CategoryGroup.svelte";
  import { Kind, type CategoryView, kindLabel, openDialog } from "../lib/store.svelte";
  import { formatEuro } from "../lib/money";
  import { dnd, drop, hoverCategoryTarget, isCategoryTarget } from "../lib/dnd.svelte";

  let { kind, categories }: { kind: Kind; categories: CategoryView[] } = $props();

  const isIncome = $derived(kind === Kind.KindIncome);
  const monthly = $derived(categories.reduce((sum, c) => sum + c.monthlyCents, 0));
  const yearly = $derived(categories.reduce((sum, c) => sum + c.yearlyCents, 0));

  /** Space below the last category: dropping a category there appends it. */
  function onTailDragOver(event: DragEvent) {
    if (dnd.source?.type === "category") hoverCategoryTarget(event, kind, categories.length);
  }
</script>

<section class="block" class:income={isIncome} class:spending={!isIncome} ondrop={drop} role="table" aria-label={kindLabel(kind)}>
  <header class="block-head">
    <div class="heading">
      <span class="dot"></span>
      <h2>{kindLabel(kind)}</h2>
      <span class="totals">
        <span class="money">{formatEuro(monthly)}</span> / month
        <span class="sep">·</span>
        <span class="money">{formatEuro(yearly)}</span> / year
      </span>
    </div>
    <div class="buttons">
      <button class="btn btn-sm" type="button" onclick={() => openDialog({ type: "category", kind })}>
        <Icon name="plus" size={14} /> Category
      </button>
      <button class="btn btn-sm {isIncome ? 'btn-income' : 'btn-spending'}" type="button" onclick={() => openDialog({ type: "entry", kind })}>
        <Icon name="plus" size={14} /> {kindLabel(kind)}
      </button>
    </div>
  </header>

  <div class="columns" role="row">
    <span></span>
    <span>Name</span>
    <span>Entered</span>
    <span class="right">Per month</span>
    <span class="right">Per year</span>
    <span></span>
  </div>

  <div class="categories">
    {#if categories.length === 0}
      <div class="empty">
        No {kindLabel(kind).toLowerCase()} categories yet. Add a category first, then add entries to it.
      </div>
    {/if}
    {#each categories as category, i (category.id)}
      <div class="cat-drop" class:active={isCategoryTarget(kind, i)}></div>
      <CategoryGroup {category} index={i} />
    {/each}
    <div class="cat-drop" class:active={isCategoryTarget(kind, categories.length)}></div>
    <div class="tail" ondragover={onTailDragOver} role="presentation"></div>
  </div>
</section>

<style>
  .block {
    /* Shared column layout for the column header, category headers and rows. */
    --cols: 28px minmax(160px, 1fr) 110px 150px 150px 72px;
    --block-color: var(--income);
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    padding: 14px 16px 6px;
  }
  .block.spending {
    --block-color: var(--spending);
  }
  .block-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 10px;
  }
  .heading {
    display: flex;
    align-items: baseline;
    gap: 10px;
    min-width: 0;
  }
  .dot {
    align-self: center;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    background: var(--block-color);
  }
  h2 {
    font-size: 17px;
  }
  .totals {
    color: var(--text-2);
    font-size: 13px;
  }
  .totals .money {
    font-weight: 600;
    color: var(--text);
  }
  .sep {
    margin: 0 6px;
    color: var(--text-3);
  }
  .buttons {
    display: flex;
    gap: 8px;
    flex-shrink: 0;
  }
  .columns {
    display: grid;
    grid-template-columns: var(--cols);
    padding: 0 8px 6px 4px;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--text-3);
  }
  .columns .right {
    text-align: right;
    padding-right: 12px;
  }
  .categories {
    display: flex;
    flex-direction: column;
  }
  .empty {
    padding: 18px 14px;
    border: 1px dashed var(--border-strong);
    border-radius: var(--radius-sm);
    color: var(--text-3);
    text-align: center;
  }
  /* Indicator between categories for category reordering. */
  .cat-drop {
    height: 8px;
    position: relative;
  }
  .cat-drop.active::before {
    content: "";
    position: absolute;
    left: 0;
    right: 0;
    top: 2.5px;
    height: 3px;
    border-radius: 2px;
    background: var(--block-color);
  }
  .tail {
    min-height: 8px;
  }
</style>
