<script lang="ts">
  /** Asks whether an import should be added to or replace the current data. */
  import Modal from "./Modal.svelte";
  import { ImportMode, type ImportPreview, closeDialog, importData } from "../lib/store.svelte";
  import { t, plural } from "../lib/i18n.svelte";

  let { preview }: { preview: ImportPreview } = $props();
  let working = $state(false);

  async function run(mode: ImportMode) {
    // Read the prop before closing: props are getters into the dialog state.
    const path = preview.path;
    working = true;
    closeDialog();
    await importData(path, mode);
  }

  const fileName = $derived(preview.path.split(/[\\/]/).pop() ?? preview.path);
  const summary = $derived(
    t("importDialog.summary", {
      file: fileName,
      categories: plural("importDialog.categories", preview.categories),
      entries: plural("importDialog.entries", preview.entries),
      version: preview.version,
    }),
  );
</script>

<Modal title={t("importDialog.title")} width={480} onclose={closeDialog}>
  <p class="intro">{summary}</p>
  <p class="question">{t("importDialog.question")}</p>
  <div class="options">
    <button class="option" type="button" disabled={working} onclick={() => run(ImportMode.ImportMerge)}>
      <strong>{t("importDialog.merge.title")}</strong>
      <span>{t("importDialog.merge.description")}</span>
    </button>
    <button class="option danger" type="button" disabled={working} onclick={() => run(ImportMode.ImportReplace)}>
      <strong>{t("importDialog.replace.title")}</strong>
      <span>{t("importDialog.replace.description")}</span>
    </button>
  </div>
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>{t("dialog.cancel")}</button>
  {/snippet}
</Modal>

<style>
  .intro {
    margin: 4px 0 8px;
    color: var(--text-2);
  }
  .question {
    margin: 0 0 10px;
    font-weight: 500;
  }
  .options {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 12px;
  }
  .option {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 10px 12px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-sm);
    background: var(--surface);
    text-align: left;
    cursor: pointer;
  }
  .option:hover {
    background: var(--accent-soft);
    border-color: var(--accent);
  }
  .option.danger:hover {
    background: var(--danger-soft);
    border-color: var(--danger);
  }
  .option span {
    font-size: 12.5px;
    color: var(--text-2);
  }
</style>
