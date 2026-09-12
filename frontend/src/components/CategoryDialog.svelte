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
    kindLabel,
    notify,
    openDialog,
  } from "../lib/store.svelte";

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
        notify("success", `Renamed category to "${name.trim()}".`);
      } else {
        await apply(Service.AddCategory(kind, name));
        notify("success", `Added ${kindLabel(kind).toLowerCase()} category "${name.trim()}".`);
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

<Modal title={isEdit ? "Rename category" : `New ${kindLabel(kind).toLowerCase()} category`} width={420} onclose={closeDialog}>
  <form id="category-form" onsubmit={submit}>
    {#if error}<p class="form-error">{error}</p>{/if}
    <div class="field">
      <label for="category-name">Name</label>
      <input id="category-name" class="input" type="text" bind:value={name} placeholder="e.g. Housing" autocomplete="off" />
    </div>
  </form>
  {#snippet footer()}
    <button class="btn" type="button" onclick={closeDialog}>Cancel</button>
    <button class="btn btn-primary" type="submit" form="category-form" disabled={working}>
      {isEdit ? "Save" : "Add category"}
    </button>
  {/snippet}
</Modal>
