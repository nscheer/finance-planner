<script lang="ts">
  /**
   * Command palette (Ctrl+K): one search box over actions, categories and
   * entries. Arrow keys move the highlight, Enter runs the item, Esc closes
   * (handled by Modal).
   */
  import { tick } from "svelte";
  import Modal from "./Modal.svelte";
  import Icon, { type IconName } from "./Icon.svelte";
  import {
    app,
    Kind,
    type EntryView,
    categoriesOf,
    closeDialog,
    openDialog,
    exportData,
    exportCSV,
    startImport,
    setAllCollapsed,
    setCollapsed,
    printPlanner,
    confirmLoadSampleData,
    isEmpty,
  } from "../lib/store.svelte";
  import { t, locales, setLocale } from "../lib/i18n.svelte";
  import { themes, setTheme } from "../lib/theme.svelte";

  type Section = "actions" | "categories" | "entries";
  interface Item {
    id: string;
    section: Section;
    label: string;
    detail?: string;
    icon: IconName;
    run: () => void | Promise<void>;
  }

  let query = $state("");
  let active = $state(0);
  let listEl: HTMLElement | undefined = $state();

  /** Opens an entry's edit dialog (replaces the palette). */
  function editEntry(entry: EntryView, kind: Kind) {
    openDialog({ type: "entry", kind, entry });
  }

  /** Expands a category and scrolls it into view. */
  async function showCategory(id: string, collapsed: boolean) {
    closeDialog();
    if (collapsed) await setCollapsed(id, false);
    await tick();
    document.getElementById(`category-${id}`)?.scrollIntoView({ behavior: "smooth", block: "center" });
  }

  const actions = $derived.by((): Item[] => {
    const list: Item[] = [
      { id: "new-spending", section: "actions", label: t("entryDialog.titleNew.spending"), icon: "plus", run: () => openDialog({ type: "entry", kind: Kind.KindSpending }) },
      { id: "new-income", section: "actions", label: t("entryDialog.titleNew.income"), icon: "plus", run: () => openDialog({ type: "entry", kind: Kind.KindIncome }) },
      { id: "new-spending-category", section: "actions", label: t("palette.newSpendingCategory"), icon: "plus", run: () => openDialog({ type: "category", kind: Kind.KindSpending }) },
      { id: "new-income-category", section: "actions", label: t("palette.newIncomeCategory"), icon: "plus", run: () => openDialog({ type: "category", kind: Kind.KindIncome }) },
      { id: "import", section: "actions", label: t("app.import"), icon: "upload", run: () => { closeDialog(); startImport(); } },
      { id: "export", section: "actions", label: t("app.export"), icon: "download", run: () => { closeDialog(); exportData(); } },
      { id: "export-csv", section: "actions", label: t("app.exportCsv"), icon: "download", run: () => { closeDialog(); exportCSV(); } },
      { id: "print", section: "actions", label: t("app.print"), icon: "printer", run: () => { closeDialog(); printPlanner(); } },
      { id: "backups", section: "actions", label: t("app.backups"), icon: "history", run: () => openDialog({ type: "backups" }) },
      { id: "goal", section: "actions", label: t("palette.goal"), icon: "target", run: () => openDialog({ type: "goal" }) },
      { id: "expand-income", section: "actions", label: t("palette.expandIncome"), icon: "expand", run: () => { closeDialog(); setAllCollapsed(Kind.KindIncome, false); } },
      { id: "collapse-income", section: "actions", label: t("palette.collapseIncome"), icon: "collapse", run: () => { closeDialog(); setAllCollapsed(Kind.KindIncome, true); } },
      { id: "expand-spending", section: "actions", label: t("palette.expandSpending"), icon: "expand", run: () => { closeDialog(); setAllCollapsed(Kind.KindSpending, false); } },
      { id: "collapse-spending", section: "actions", label: t("palette.collapseSpending"), icon: "collapse", run: () => { closeDialog(); setAllCollapsed(Kind.KindSpending, true); } },
      { id: "shortcuts", section: "actions", label: t("shortcuts.title"), icon: "keyboard", run: () => openDialog({ type: "shortcuts" }) },
    ];
    for (const theme of themes) {
      list.push({ id: `theme-${theme}`, section: "actions", label: t("palette.theme", { name: t(`theme.${theme}`) }), icon: "sun", run: () => { closeDialog(); setTheme(theme); } });
    }
    for (const locale of locales) {
      list.push({ id: `lang-${locale.code}`, section: "actions", label: t("palette.language", { name: locale.label }), icon: "globe", run: () => { closeDialog(); setLocale(locale.code); } });
    }
    if (isEmpty()) {
      list.push({ id: "sample", section: "actions", label: t("app.getStarted.sample"), icon: "sparkles", run: () => confirmLoadSampleData() });
    }
    return list;
  });

  const dataItems = $derived.by((): Item[] => {
    const list: Item[] = [];
    for (const kind of [Kind.KindIncome, Kind.KindSpending]) {
      const kindLabel = t(kind === Kind.KindIncome ? "kind.income" : "kind.spending");
      for (const category of categoriesOf(kind)) {
        list.push({
          id: `cat-${category.id}`,
          section: "categories",
          label: category.name,
          detail: kindLabel,
          icon: "chevron",
          run: () => showCategory(category.id, category.collapsed),
        });
      }
      for (const category of categoriesOf(kind)) {
        for (const entry of category.entries ?? []) {
          list.push({
            id: `entry-${entry.id}`,
            section: "entries",
            label: entry.name,
            detail: category.name,
            icon: "edit",
            run: () => editEntry(entry, kind),
          });
        }
      }
    }
    return list;
  });

  const results = $derived.by((): Item[] => {
    const q = query.trim().toLowerCase();
    const all = [...actions, ...dataItems];
    const matches = q === "" ? all : all.filter((i) => i.label.toLowerCase().includes(q) || (i.detail ?? "").toLowerCase().includes(q));
    return matches.slice(0, 30);
  });

  // Keep the highlight valid and visible when the results change.
  $effect(() => {
    void results;
    if (active >= results.length) active = Math.max(0, results.length - 1);
  });
  $effect(() => {
    void active;
    listEl?.querySelector<HTMLElement>(`[data-index="${active}"]`)?.scrollIntoView({ block: "nearest" });
  });

  function onKeydown(event: KeyboardEvent) {
    if (event.key === "ArrowDown") {
      active = Math.min(active + 1, results.length - 1);
      event.preventDefault();
    } else if (event.key === "ArrowUp") {
      active = Math.max(active - 1, 0);
      event.preventDefault();
    } else if (event.key === "Enter") {
      event.preventDefault();
      results[active]?.run();
    }
  }

  const sectionLabel = (s: Section) => t(`palette.${s}`);
</script>

<Modal title={t("shortcuts.palette")} width={560} onclose={closeDialog}>
  <input
    class="input"
    type="text"
    placeholder={t("palette.placeholder")}
    aria-label={t("palette.placeholder")}
    autocomplete="off"
    bind:value={query}
    onkeydown={onKeydown}
  />
  <ul class="results" bind:this={listEl} role="listbox">
    {#each results as item, i (item.id)}
      {#if i === 0 || results[i - 1].section !== item.section}
        <li class="section" role="presentation">{sectionLabel(item.section)}</li>
      {/if}
      <li role="option" aria-selected={i === active} data-index={i}>
        <button type="button" class="item" class:active={i === active} onmouseenter={() => (active = i)} onclick={() => item.run()}>
          <span class="icon"><Icon name={item.icon} size={14} /></span>
          <span class="label">{item.label}</span>
          {#if item.detail}<span class="detail">{item.detail}</span>{/if}
        </button>
      </li>
    {:else}
      <li class="empty">{t("palette.empty")}</li>
    {/each}
  </ul>
  <p class="hint">{t("palette.hint")}</p>
</Modal>

<style>
  .results {
    margin: 10px 0 6px;
    padding: 0;
    list-style: none;
    max-height: 360px;
    overflow: auto;
  }
  .section {
    padding: 8px 6px 4px;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--text-3);
  }
  .item {
    display: grid;
    grid-template-columns: 20px 1fr auto;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 7px 8px;
    border: 0;
    border-radius: var(--radius-sm);
    background: transparent;
    text-align: left;
    cursor: pointer;
  }
  .item.active {
    background: var(--accent-soft);
  }
  .icon {
    display: flex;
    color: var(--text-3);
  }
  .label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .detail {
    font-size: 12px;
    color: var(--text-3);
    white-space: nowrap;
  }
  .empty {
    padding: 12px 8px;
    color: var(--text-3);
  }
  .hint {
    margin: 0 0 8px;
  }
</style>
