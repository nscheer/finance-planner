<script lang="ts">
  /** Asks whether an import should be added to or replace the current data. */
  import Modal from "./Modal.svelte";
  import { ImportMode, type ImportPreview, closeDialog, importData } from "../lib/store.svelte";

  let { preview }: { preview: ImportPreview } = $props();
  let working = $state(false);

  async function run(mode: ImportMode) {
    working = true;
    closeDialog();
    await importData(preview.path, mode);
    working = false;
  }

  const fileName = $derived(preview.path.split(/[\\/]/).pop() ?? preview.path);
</script>

<Modal title="Import data" width={480} onclose={closeDialog}>
  <p class="intro">
    <strong>{fileName}</strong> contains {preview.categories}
    {preview.categories === 1 ? "category" : "categories"} and {preview.entries}
    {preview.entries === 1 ? "entry" : "entries"} (file version {preview.version}).
  </p>
  <p class="question">How should the data be imported?</p>
  <div class="options">
    <button class="option" type="button" disabled={working} onclick={() => run(ImportMode.ImportMerge)}>
      <strong>Add to current data</strong>
      <span>Keeps everything you have. Categories with the same name are reused.</span>
    </button>
    <button class="option danger" type="button" disabled={working} onclick={() => run(ImportMode.ImportReplace)}>
      <strong>Replace current data</strong>
      <span>Deletes all current categories and entries and uses the imported file instead.</span>
    </button>
  </div>
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>Cancel</button>
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
