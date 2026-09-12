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
  import SavingsGoalDialog from "./components/SavingsGoalDialog.svelte";
  import BackupsDialog from "./components/BackupsDialog.svelte";
  import ShortcutsDialog from "./components/ShortcutsDialog.svelte";
  import SelectionBar from "./components/SelectionBar.svelte";
  import CommandPalette from "./components/CommandPalette.svelte";
  import {
    app,
    Kind,
    Period,
    loadState,
    exportData,
    exportCSV,
    startImport,
    alert,
    errorMessage,
    openDialog,
    closeDialog,
    visibleCategories,
    filterActive,
    filterCounts,
    clearFilter,
    isEmpty,
    confirmLoadSampleData,
    selectionCount,
    clearSelection,
    printPlanner,
    type PeriodFilter,
  } from "./lib/store.svelte";
  import { i18n, t, locales, setLocale, formatEuro, formatDateTime, type LocaleCode } from "./lib/i18n.svelte";
  import { theme, themes, setTheme, type Theme } from "./lib/theme.svelte";
  import { dnd, endDrag } from "./lib/dnd.svelte";

  onMount(() => {
    loadState();
  });

  // A changed filter changes the visible rows: drop the selection.
  $effect(() => {
    void app.filter.query;
    void app.filter.period;
    clearSelection();
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

  let searchEl: HTMLInputElement | undefined = $state();

  const periodFilters: { value: PeriodFilter; label: string }[] = $derived([
    { value: "all", label: t("app.filter.all") },
    { value: Period.PeriodMonthly, label: t("entry.monthly") },
    { value: Period.PeriodQuarterly, label: t("entry.quarterly") },
    { value: Period.PeriodHalfYearly, label: t("entry.halfyearly") },
    { value: Period.PeriodYearly, label: t("entry.yearly") },
    { value: "paused", label: t("app.filter.paused") },
  ]);

  /**
   * Keyboard shortcuts. They are ignored while a dialog is open or an input
   * has the focus, so typing never triggers them.
   */
  function onKeydown(event: KeyboardEvent) {
    const target = event.target as HTMLElement | null;
    // Only text entry counts as "in a field"; a focused checkbox or button
    // must not swallow the shortcuts (e.g. Esc after clicking a checkbox).
    const textInput =
      !!target &&
      target.tagName === "INPUT" &&
      !["checkbox", "radio", "button", "submit"].includes((target as HTMLInputElement).type);
    const inField = !!target && (textInput || ["TEXTAREA", "SELECT"].includes(target.tagName) || target.isContentEditable);
    const ctrlF = (event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "f";
    const ctrlP = (event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "p";
    const ctrlK = (event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k";

    if (ctrlK) {
      event.preventDefault();
      if (app.state) openDialog({ type: "palette" });
      return;
    }
    if (ctrlF) {
      event.preventDefault();
      if (!app.dialog) searchEl?.focus();
      return;
    }
    if (ctrlP) {
      event.preventDefault();
      if (!app.dialog && app.state) printPlanner();
      return;
    }
    if (event.key === "Escape") {
      if (app.dialog) return; // Modal.svelte handles it
      if (inField && target === searchEl) {
        clearFilter();
        searchEl?.blur();
        event.preventDefault();
      } else if (!inField && (selectionCount() > 0 || app.selectMode)) {
        clearSelection();
        event.preventDefault();
      }
      return;
    }
    if (app.dialog || inField || event.ctrlKey || event.metaKey || event.altKey) return;
    if (!app.state) return;

    switch (event.key) {
      case "n":
        openDialog({ type: "entry", kind: Kind.KindSpending });
        break;
      case "i":
        openDialog({ type: "entry", kind: Kind.KindIncome });
        break;
      case "c":
        openDialog({ type: "category", kind: Kind.KindSpending });
        break;
      case "/":
        searchEl?.focus();
        break;
      case "?":
        openDialog({ type: "shortcuts" });
        break;
      default:
        return;
    }
    event.preventDefault();
  }

  async function onThemeChange(event: Event) {
    const value = (event.currentTarget as HTMLSelectElement).value as Theme;
    const err = await setTheme(value);
    if (err) alert(t("alert.generic"), errorMessage(err));
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
  onkeydown={onKeydown}
/>

<div class="app">
  <header class="topbar">
    <div class="brand">
      <h1>{t("app.title")}</h1>
      {#if app.state}
        <span class="path" title={app.state.dataPath}>{app.state.dataPath}</span>
      {/if}
    </div>
    <div class="filters">
      <div class="search" class:active={app.filter.query.trim() !== ""}>
        <span class="search-icon"><Icon name="search" size={14} /></span>
        <input
          class="input search-input"
          type="text"
          placeholder={t("app.search")}
          aria-label={t("shortcuts.search")}
          autocomplete="off"
          bind:value={app.filter.query}
          bind:this={searchEl}
        />
        {#if app.filter.query !== ""}
          <button class="icon-btn small" type="button" title={t("app.searchClear")} aria-label={t("app.searchClear")} onclick={() => { app.filter.query = ""; searchEl?.focus(); }}><Icon name="close" size={13} /></button>
        {/if}
      </div>
      <select class="select select-sm period-filter" class:active={app.filter.period !== "all"} bind:value={app.filter.period} aria-label={t("app.filter.all")}>
        {#each periodFilters as f (f.value)}
          <option value={f.value}>{f.label}</option>
        {/each}
      </select>
      {#if filterActive()}
        {@const counts = filterCounts()}
        <span class="filter-result">{t("app.filterResult", counts)}</span>
        <button class="icon-btn" type="button" title={t("app.filterReset")} aria-label={t("app.filterReset")} onclick={clearFilter}><Icon name="close" size={14} /></button>
      {/if}
    </div>
    <div class="actions">
      <button class="btn btn-sm" type="button" onclick={startImport}><Icon name="upload" size={14} /> {t("app.import")}</button>
      <button class="btn btn-sm" type="button" onclick={exportData}><Icon name="download" size={14} /> {t("app.export")}</button>
      <button class="btn btn-sm" type="button" onclick={exportCSV}><Icon name="download" size={14} /> {t("app.exportCsv")}</button>
      <button class="btn btn-sm" type="button" onclick={() => openDialog({ type: "backups" })}><Icon name="history" size={14} /> {t("app.backups")}</button>
      <button class="btn btn-sm" type="button" onclick={printPlanner}><Icon name="printer" size={14} /> {t("app.print")}</button>
      <button class="icon-btn" type="button" title="{t('shortcuts.palette')} (Ctrl+K)" aria-label={t("shortcuts.palette")} onclick={() => openDialog({ type: "palette" })}><Icon name="command" size={16} /></button>
      <button class="icon-btn" type="button" title={t("app.shortcuts")} aria-label={t("app.shortcuts")} onclick={() => openDialog({ type: "shortcuts" })}><Icon name="keyboard" size={16} /></button>
      <span class="divider"></span>
      <label class="language" title={t("app.theme")}>
        <Icon name="sun" size={14} />
        <select class="select select-sm" value={theme.current} onchange={onThemeChange} aria-label={t("app.theme")}>
          {#each themes as value (value)}
            <option {value}>{t(`theme.${value}`)}</option>
          {/each}
        </select>
      </label>
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

  <main class="content" class:has-selection={app.selectMode}>
    {#if app.state}
      <!-- Shown only on paper (see the print stylesheet). -->
      <header class="print-header">
        <h1>{t("app.title")}</h1>
        <span>{t("print.generated", { date: formatDateTime(new Date()) })} · {app.state.dataPath}</span>
      </header>
    {/if}
    {#if app.loadError}
      <div class="load-error">
        <strong>{t("app.loadError")}</strong>
        <p>{app.loadError}</p>
        <button class="btn" type="button" onclick={loadState}>{t("app.retry")}</button>
      </div>
    {:else if app.state}
      <div class="tables">
        {#if app.state.stats.saldoMonthlyCents < 0}
          <div class="banner danger" role="alert">
            <Icon name="alert" size={16} />
            <span>{t("warning.negativeSaldo", { amount: formatEuro(-app.state.stats.saldoMonthlyCents) })}</span>
          </div>
        {:else if !app.state.stats.goalReachable}
          <div class="banner warn" role="status">
            <Icon name="target" size={16} />
            <span>{t("warning.goal", { amount: formatEuro(-app.state.stats.remainingAfterGoalCents) })}</span>
          </div>
        {/if}
        {#if isEmpty()}
          <div class="get-started">
            <span class="icon"><Icon name="sparkles" size={22} /></span>
            <div>
              <strong>{t("app.getStarted.title")}</strong>
              <p>{t("app.getStarted.text")}</p>
              <button class="btn btn-primary btn-sm" type="button" onclick={confirmLoadSampleData}>{t("app.getStarted.sample")}</button>
            </div>
          </div>
        {/if}
        <Block kind={Kind.KindIncome} categories={visibleCategories(Kind.KindIncome)} />
        <Block kind={Kind.KindSpending} categories={visibleCategories(Kind.KindSpending)} />
      </div>
      <StatsPanel stats={app.state.stats} spending={app.state.spending ?? []} />
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
  {:else if dialog.type === "goal"}
    <SavingsGoalDialog />
  {:else if dialog.type === "backups"}
    <BackupsDialog />
  {:else if dialog.type === "shortcuts"}
    <ShortcutsDialog />
  {:else if dialog.type === "palette"}
    <CommandPalette />
  {/if}
{/if}

{#if app.selectMode}
  <SelectionBar />
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
  .filters {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1;
    min-width: 0;
  }
  .search {
    display: flex;
    align-items: center;
    gap: 6px;
    flex: 1;
    max-width: 420px;
    min-width: 200px;
    padding: 2px 4px 2px 8px;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--surface-2);
  }
  .search.active,
  .period-filter.active {
    border-color: var(--accent);
  }
  .period-filter {
    flex-shrink: 0;
  }
  .search-icon {
    display: flex;
    color: var(--text-3);
  }
  .search-input {
    flex: 1;
    min-width: 80px;
    padding: 4px 6px;
    border: 0;
    background: transparent;
  }
  .search-input:focus {
    outline: none;
  }
  .search .icon-btn.small {
    width: 22px;
    height: 22px;
  }
  .filter-result {
    font-size: 12px;
    color: var(--text-2);
    white-space: nowrap;
  }
  .get-started {
    display: flex;
    gap: 14px;
    padding: 16px 18px;
    border: 1px dashed var(--border-strong);
    border-radius: var(--radius);
    background: var(--surface);
  }
  .get-started .icon {
    color: var(--accent);
    margin-top: 2px;
  }
  .get-started p {
    margin: 4px 0 10px;
    color: var(--text-2);
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
    /* No top padding here: the top gap is a margin on both columns, so the
       sticky statistics box starts exactly level with the income block. */
    padding: 0 24px 32px;
  }
  .tables {
    display: flex;
    flex-direction: column;
    gap: 20px;
    min-width: 0;
    padding-top: 20px;
  }
  /* Room to scroll the last rows above the floating selection toolbar. */
  .content.has-selection {
    padding-bottom: 96px;
  }
  .banner {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    border-radius: var(--radius-sm);
    font-weight: 500;
  }
  .banner.danger {
    background: var(--danger-soft);
    color: var(--danger);
  }
  .banner.warn {
    background: var(--warn-soft);
    color: var(--warn);
  }
  .print-header {
    display: none;
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
