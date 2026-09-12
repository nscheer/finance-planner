<script lang="ts">
  /** Add or rename a category. */
  import Modal from "./Modal.svelte";
  import {
    Service,
    Kind,
    type CategoryView,
    apply,
    closeDialog,
    errorMessage,
    kindKey,
    notify,
    openDialog,
  } from "../lib/store.svelte";
  import { t } from "../lib/i18n.svelte";
  import { untrack } from "svelte";

  let { kind, category, returnToEntry = false }: { kind: Kind; category?: CategoryView; returnToEntry?: boolean } = $props();

  // Seeded once from the initial props; the dialog is recreated on each open.
  const initial = untrack(() => category);
  const isEdit = !!initial;

  let name = $state(initial?.name ?? "");
  let error = $state("");
  let working = $state(false);

  async function submit(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    working = true;
    try {
      if (initial) {
        await apply(Service.RenameCategory(initial.id, name));
        notify("success", t("toast.categoryRenamed", { name: name.trim() }));
      } else {
        await apply(Service.AddCategory(kind, name));
        notify("success", t(kindKey("toast.categoryAdded", kind), { name: name.trim() }));
      }
      if (returnToEntry) {
        openDialog({ type: "entry", kind });
        return;
      }
      closeDialog();
    } catch (err) {
      error = errorMessage(err);
    } finally {
      working = false;
    }
  }
</script>

<Modal title={isEdit ? t("categoryDialog.titleRename") : t(kindKey("categoryDialog.titleNew", kind))} width={420} onclose={closeDialog}>
  <form id="category-form" onsubmit={submit}>
    {#if error}<p class="form-error">{error}</p>{/if}
    <div class="field">
      <label for="category-name">{t("categoryDialog.name")}</label>
      <input id="category-name" class="input" type="text" bind:value={name} placeholder={t("categoryDialog.namePlaceholder")} autocomplete="off" />
    </div>
  </form>
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>{t("dialog.cancel")}</button>
    <button class="btn btn-primary" type="submit" form="category-form" disabled={working}>
      {isEdit ? t("dialog.save") : t("categoryDialog.submit")}
    </button>
  {/snippet}
</Modal>
