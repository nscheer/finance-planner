<script lang="ts">
  /**
   * Application shell: top bar with global actions, the two blocks (income,
   * spending) on the left, the statistics box on the right, and the currently
   * open modal dialog.
   */
  import { onMount } from "svelte";
  import Icon from "./components/Icon.svelte";
  import Block from "./components/Block.svelte";
  import StatsPanel from "./components/StatsPanel.svelte";
  import Toasts from "./components/Toasts.svelte";
  import EntryDialog from "./components/EntryDialog.svelte";
  import CategoryDialog from "./components/CategoryDialog.svelte";
  import ConfirmDialog from "./components/ConfirmDialog.svelte";
  import AlertDialog from "./components/AlertDialog.svelte";
  import ImportDialog from "./components/ImportDialog.svelte";
  import { app, Kind, loadState, exportData, startImport } from "./lib/store.svelte";
  import { dnd, endDrag } from "./lib/dnd.svelte";

  /**
   * Runs last in the bubbling chain: if no row/category accepted the
   * dragover, the pointer is over a non-droppable area and the indicator
   * must disappear so it never suggests a drop that wouldn't happen.
   */
  function onWindowDragOver(event: DragEvent) {
    if (!event.defaultPrevented && dnd.target) dnd.target = null;
  }

  onMount(() => {
    loadState();
  });
</script>

<!-- A drag that ends outside any drop zone must clear the indicators. -->
<svelte:window ondragend={endDrag} ondrop={endDrag} ondragover={onWindowDragOver} />

<div class="app">
  <header class="topbar">
    <div class="brand">
      <h1>Finance Planner</h1>
      {#if app.state}
        <span class="path" title={app.state.dataPath}>{app.state.dataPath}</span>
      {/if}
    </div>
    <div class="actions">
      <button class="btn btn-sm" type="button" onclick={startImport}><Icon name="upload" size={14} /> Import</button>
      <button class="btn btn-sm" type="button" onclick={exportData}><Icon name="download" size={14} /> Export</button>
    </div>
  </header>

  <main class="content">
    {#if app.loadError}
      <div class="load-error">
        <strong>Could not load the data file.</strong>
        <p>{app.loadError}</p>
        <button class="btn" type="button" onclick={loadState}>Retry</button>
      </div>
    {:else if app.state}
      <div class="tables">
        <Block kind={Kind.KindIncome} categories={app.state.income ?? []} />
        <Block kind={Kind.KindSpending} categories={app.state.spending ?? []} />
      </div>
      <StatsPanel stats={app.state.stats} />
    {:else}
      <p class="muted">Loading…</p>
    {/if}
  </main>
</div>

{#if app.dialog}
  {#if app.dialog.type === "entry"}
    <EntryDialog kind={app.dialog.kind} entry={app.dialog.entry} categoryId={app.dialog.categoryId} />
  {:else if app.dialog.type === "category"}
    <CategoryDialog kind={app.dialog.kind} category={app.dialog.category} returnToEntry={app.dialog.returnToEntry} />
  {:else if app.dialog.type === "confirm"}
    <ConfirmDialog
      title={app.dialog.title}
      message={app.dialog.message}
      confirmLabel={app.dialog.confirmLabel}
      onConfirm={app.dialog.onConfirm}
    />
  {:else if app.dialog.type === "alert"}
    <AlertDialog title={app.dialog.title} message={app.dialog.message} />
  {:else if app.dialog.type === "import"}
    <ImportDialog preview={app.dialog.preview} />
  {/if}
{/if}

<Toasts />

<style>
  .app {
    height: 100vh;
    display: flex;
    flex-direction: column;
    min-width: var(--page-min);
  }
  .topbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 10px 24px;
    background: var(--surface);
    border-bottom: 1px solid var(--border);
    flex-shrink: 0;
  }
  .brand {
    display: flex;
    align-items: baseline;
    gap: 14px;
    min-width: 0;
  }
  h1 {
    font-size: 18px;
    letter-spacing: -0.01em;
  }
  .path {
    font-size: 12px;
    color: var(--text-3);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }
  .content {
    flex: 1;
    overflow: auto;
    display: grid;
    grid-template-columns: minmax(0, 1fr) var(--stats-width);
    align-items: start;
    gap: 20px;
    width: 100%;
    max-width: var(--page-max);
    margin: 0 auto;
    padding: 20px 24px 32px;
  }
  .tables {
    display: flex;
    flex-direction: column;
    gap: 20px;
    min-width: 0;
  }
  .load-error {
    grid-column: 1 / -1;
    padding: 20px;
    background: var(--danger-soft);
    border-radius: var(--radius);
    color: var(--danger);
  }
  .load-error p {
    user-select: text;
  }
</style>
