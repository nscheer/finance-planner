<script lang="ts">
  /**
   * Application shell: top bar with global actions and the language
   * dropdown, the two blocks (income, spending) on the left, the statistics
   * box on the right, and the currently open modal dialog.
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
  import { app, Kind, loadState, exportData, startImport, alert, errorMessage } from "./lib/store.svelte";
  import { i18n, t, locales, setLocale, type LocaleCode } from "./lib/i18n.svelte";
  import { dnd, endDrag } from "./lib/dnd.svelte";

  onMount(() => {
    loadState();
  });

  /**
   * Last line of defence: an exception inside an event handler would
   * otherwise fail silently. Show it, so a broken action is never invisible.
   */
  function onUnhandledError(event: Event) {
    const err =
      event instanceof PromiseRejectionEvent
        ? event.reason
        : event instanceof ErrorEvent
          ? (event.error ?? event.message)
          : event;
    console.error(err);
    alert(t("alert.generic"), errorMessage(err));
  }

  /**
   * Runs last in the bubbling chain: if no row/category accepted the
   * dragover, the pointer is over a non-droppable area and the indicator
   * must disappear so it never suggests a drop that wouldn't happen.
   */
  function onWindowDragOver(event: DragEvent) {
    if (!event.defaultPrevented && dnd.target) dnd.target = null;
  }

  async function onLanguageChange(event: Event) {
    const code = (event.currentTarget as HTMLSelectElement).value as LocaleCode;
    const err = await setLocale(code);
    if (err) alert(t("alert.generic"), errorMessage(err));
  }
</script>

<!-- A drag that ends outside any drop zone must clear the indicators. -->
<svelte:window
  ondragend={endDrag}
  ondrop={endDrag}
  ondragover={onWindowDragOver}
  onerror={onUnhandledError}
  onunhandledrejection={onUnhandledError}
/>

<div class="app">
  <header class="topbar">
    <div class="brand">
      <h1>{t("app.title")}</h1>
      {#if app.state}
        <span class="path" title={app.state.dataPath}>{app.state.dataPath}</span>
      {/if}
    </div>
    <div class="actions">
      <button class="btn btn-sm" type="button" onclick={startImport}><Icon name="upload" size={14} /> {t("app.import")}</button>
      <button class="btn btn-sm" type="button" onclick={exportData}><Icon name="download" size={14} /> {t("app.export")}</button>
      <span class="divider"></span>
      <label class="language" title={t("app.language")}>
        <Icon name="globe" size={14} />
        <select class="select select-sm" value={i18n.locale.code} onchange={onLanguageChange} aria-label={t("app.language")}>
          {#each locales as locale (locale.code)}
            <option value={locale.code}>{locale.label}</option>
          {/each}
        </select>
      </label>
    </div>
  </header>

  <main class="content">
    {#if app.loadError}
      <div class="load-error">
        <strong>{t("app.loadError")}</strong>
        <p>{app.loadError}</p>
        <button class="btn" type="button" onclick={loadState}>{t("app.retry")}</button>
      </div>
    {:else if app.state}
      <div class="tables">
        <Block kind={Kind.KindIncome} categories={app.state.income ?? []} />
        <Block kind={Kind.KindSpending} categories={app.state.spending ?? []} />
      </div>
      <StatsPanel stats={app.state.stats} />
    {:else}
      <p class="muted">{t("app.loading")}</p>
    {/if}
  </main>
</div>

<!--
  The dialog is bound to a local constant on purpose: Svelte 5 passes props as
  getters, and a getter like `app.dialog.preview` would throw once the dialog
  closed itself (app.dialog = null) but still needs its props afterwards.
-->
{#if app.dialog}
  {@const dialog = app.dialog}
  {#if dialog.type === "entry"}
    <EntryDialog kind={dialog.kind} entry={dialog.entry} categoryId={dialog.categoryId} duplicateOf={dialog.duplicateOf} />
  {:else if dialog.type === "category"}
    <CategoryDialog kind={dialog.kind} category={dialog.category} returnToEntry={dialog.returnToEntry} />
  {:else if dialog.type === "confirm"}
    <ConfirmDialog title={dialog.title} message={dialog.message} confirmLabel={dialog.confirmLabel} onConfirm={dialog.onConfirm} />
  {:else if dialog.type === "alert"}
    <AlertDialog title={dialog.title} message={dialog.message} />
  {:else if dialog.type === "import"}
    <ImportDialog preview={dialog.preview} />
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
  .divider {
    width: 1px;
    height: 22px;
    margin: 0 4px;
    background: var(--border);
  }
  .language {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--text-2);
  }
  .select-sm {
    width: auto;
    padding: 4px 8px;
    font-size: 13px;
    color: var(--text);
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
