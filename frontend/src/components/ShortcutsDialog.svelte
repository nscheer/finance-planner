<script lang="ts">
  /** Lists the keyboard shortcuts (see App.svelte for the handling). */
  import Modal from "./Modal.svelte";
  import { closeDialog } from "../lib/store.svelte";
  import { t } from "../lib/i18n.svelte";

  const rows: { keys: string[]; label: "shortcuts.newSpending" | "shortcuts.newIncome" | "shortcuts.newCategory" | "shortcuts.search" | "shortcuts.print" | "shortcuts.palette" | "shortcuts.clear" | "shortcuts.help" }[] = [
    { keys: ["N"], label: "shortcuts.newSpending" },
    { keys: ["I"], label: "shortcuts.newIncome" },
    { keys: ["C"], label: "shortcuts.newCategory" },
    { keys: ["/", "Ctrl + F"], label: "shortcuts.search" },
    { keys: ["Ctrl + K"], label: "shortcuts.palette" },
    { keys: ["Ctrl + P"], label: "shortcuts.print" },
    { keys: ["Esc"], label: "shortcuts.clear" },
    { keys: ["?"], label: "shortcuts.help" },
  ];
</script>

<Modal title={t("shortcuts.title")} width={420} onclose={closeDialog}>
  <table>
    <tbody>
      {#each rows as row (row.label)}
        <tr>
          <td class="keys">
            {#each row.keys as k, i (k)}
              {#if i > 0}<span class="or">/</span>{/if}<kbd>{k}</kbd>
            {/each}
          </td>
          <td>{t(row.label)}</td>
        </tr>
      {/each}
    </tbody>
  </table>
  <p class="hint">{t("shortcuts.hint")}</p>
  {#snippet footer()}
    <button class="btn btn-primary" type="button" onclick={closeDialog}>{t("dialog.ok")}</button>
  {/snippet}
</Modal>

<style>
  table {
    width: 100%;
    border-collapse: collapse;
    margin-bottom: 10px;
  }
  td {
    padding: 6px 4px;
    border-top: 1px solid var(--border);
  }
  tr:first-child td {
    border-top: 0;
  }
  .keys {
    width: 130px;
    white-space: nowrap;
  }
  .or {
    margin: 0 4px;
    color: var(--text-3);
  }
  kbd {
    display: inline-block;
    padding: 1px 7px;
    border: 1px solid var(--border-strong);
    border-bottom-width: 2px;
    border-radius: 4px;
    background: var(--surface-2);
    font-family: var(--mono);
    font-size: 12px;
  }
  .hint {
    margin: 0 0 12px;
  }
</style>
