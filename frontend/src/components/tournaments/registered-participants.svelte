<script lang="ts">
  /* eslint-disable svelte/no-useless-mustaches */
  import GripVerticalIcon from "@lucide/svelte/icons/grip-vertical";
  import PlusIcon from "@lucide/svelte/icons/plus";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";
  import UsersIcon from "@lucide/svelte/icons/users";
  import { createMutation, useQueryClient } from "@tanstack/svelte-query";

  import { Button } from "$src/components/ui/button/index.js";
  import * as Dialog from "$src/components/ui/dialog/index.js";
  import * as Table from "$src/components/ui/table/index.js";
  import { Textarea } from "$src/components/ui/textarea/index.js";
  import {
    DeleteParticipant,
    InsertBulkParticipant,
    UpdateTournament,
  } from "$wails/go/main/App";
  import type { entrant, tournament } from "$wails/go/models";

  type Props = {
    tournament: tournament.Tournament;
    entrants: entrant.Entrant[];
    onDelete?: (entrant: entrant.Entrant) => void | Promise<void>;
  };

  let { tournament, entrants }: Props = $props();

  const queryClient = useQueryClient();

  // participants can only be added, removed or reordered before the
  // tournament has started
  const editable = $derived(tournament.status === "pending");

  const orderIds = $derived.by(() => {
    const order = tournament.qualifying_order as unknown as string | string[];
    if (Array.isArray(order)) return order;
    return (order ?? "")
      .split(",")
      .map((id) => id.trim())
      .filter((id) => id !== "");
  });

  const orderedEntrants = $derived.by(() => {
    const rank = new Map(orderIds.map((id, i) => [id, i]));
    return entrants
      .map((e, i) => ({ e, i }))
      .sort((a, b) => {
        const ra = rank.get(a.e.participantId) ?? Number.MAX_SAFE_INTEGER;
        const rb = rank.get(b.e.participantId) ?? Number.MAX_SAFE_INTEGER;
        return ra - rb || a.i - b.i;
      })
      .map(({ e }) => e);
  });

  // local copy so rows can be rearranged immediately; resyncs whenever the
  // entrants or the tournament's qualifying order change
  // eslint-disable-next-line svelte/prefer-writable-derived
  let items = $state<entrant.Entrant[]>([]);
  $effect(() => {
    items = [...orderedEntrants];
  });

  // ---------- mutations ----------
  const insertParticipantsMutation = createMutation(() => ({
    mutationFn: async ({ names }: { names: string[] }) =>
      InsertBulkParticipant(tournament.id, names),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["entrants", tournament.id],
      });
    },
  }));

  const updateTournamentMutation = createMutation(() => ({
    mutationFn: async ({ qualifyingOrder }: { qualifyingOrder: string[] }) =>
      UpdateTournament(tournament.id, { qualifyingOrder }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["tournament", tournament.id],
      });

      await queryClient.invalidateQueries({
        queryKey: ["logs", tournament.id],
      });
    },
    onError: () => {
      // put the rows back to the last known server order
      items = [...orderedEntrants];
    },
  }));

  const deleteParticipantMutation = createMutation(() => ({
    mutationFn: async ({ participantId }: { participantId: string }) =>
      DeleteParticipant(tournament.id, participantId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["entrants", tournament.id],
      });

      // the participant is also removed from the qualifying order
      await queryClient.invalidateQueries({
        queryKey: ["tournament", tournament.id],
      });
    },
  }));

  // ---------- drag and drop ----------
  let dragIndex = $state<number | null>(null);
  let overIndex = $state<number | null>(null);

  function handleDragStart(e: DragEvent, index: number) {
    if (!editable) return;
    dragIndex = index;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = "move";
      e.dataTransfer.setData("text/plain", String(index));
    }
  }

  function handleDragOver(e: DragEvent, index: number) {
    if (!editable || dragIndex === null) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = "move";
    overIndex = index;
  }

  function handleDrop(e: DragEvent, index: number) {
    e.preventDefault();
    const from = dragIndex;
    resetDrag();
    if (!editable || from === null || from === index) return;

    const next = [...items];
    const [moved] = next.splice(from, 1);
    next.splice(index, 0, moved);
    items = next;

    updateTournamentMutation.mutate({
      qualifyingOrder: next.map((entrant) => entrant.participantId),
    });
  }

  function resetDrag() {
    dragIndex = null;
    overIndex = null;
  }

  // ---------- create dialog ----------
  let createOpen = $state(false);
  let rawNames = $state("");

  const parsedNames = $derived(
    rawNames
      .split("\n")
      .map((n) => n.trim())
      .filter((n) => n !== ""),
  );

  function submitCreate() {
    if (!editable || parsedNames.length === 0) return;

    insertParticipantsMutation.mutate(
      { names: parsedNames },
      {
        onSuccess: () => {
          rawNames = "";
          createOpen = false;
        },
      },
    );
  }

  // ---------- delete dialog ----------
  let deleting = $state(false);
  let deleteTarget = $state<entrant.Entrant | null>(null);

  const deleteOpen = $derived(deleteTarget !== null);

  async function confirmDelete() {
    if (!editable || !deleteTarget) return;
    deleting = true;

    await deleteParticipantMutation
      .mutateAsync(
        { participantId: deleteTarget.participantId },
        {
          onSuccess: () => {
            deleteTarget = null;
          },
        },
      )
      .finally(() => (deleting = false));
  }
</script>

<div class="mr-0 lg:mr-80">
  <div
    class="mx-auto max-w-6xl space-y-1 px-4 py-6 pb-2 sm:px-6 lg:py-8 lg:pb-2"
  >
    <div class="flex items-end justify-between mb-4">
      <div>
        <h2 class="text-lg font-semibold">Participants</h2>
        <p class="text-sm text-muted-foreground">
          {items.length}
          {items.length === 1 ? "participant" : "participants"} registered.
          {#if editable}
            Drag rows to change the qualifying order.
          {/if}
        </p>
      </div>
      {#if editable}
        <Button onclick={() => (createOpen = true)}>
          <PlusIcon class="mr-2 size-4" />
          Add participants
        </Button>
      {/if}
    </div>

    {#if items.length === 0}
      <div
        class="flex flex-col items-center justify-center rounded-lg border border-dashed px-6 py-16 text-center"
      >
        <div
          class="mb-4 flex size-12 items-center justify-center rounded-full bg-muted"
        >
          <UsersIcon class="size-6 text-muted-foreground" />
        </div>
        <h3 class="text-base font-semibold">No participants yet</h3>
        {#if editable}
          <p class="mt-1 max-w-sm text-sm text-muted-foreground">
            Add the people taking part in this tournament. You can paste several
            names at once, one per line.
          </p>
          <Button class="mt-6" onclick={() => (createOpen = true)}>
            <PlusIcon class="mr-2 size-4" />
            Add participants
          </Button>
        {:else}
          <p class="mt-1 max-w-sm text-sm text-muted-foreground">
            This tournament has already started, so participants can no longer
            be added.
          </p>
        {/if}
      </div>
    {:else}
      <div class="rounded-lg border">
        <Table.Root>
          <Table.Header>
            <Table.Row>
              {#if editable}
                <Table.Head class="w-10"></Table.Head>
              {/if}
              <Table.Head class="w-16">#</Table.Head>
              <Table.Head>Name</Table.Head>
              {#if editable}
                <Table.Head class="w-16 text-right">Actions</Table.Head>
              {/if}
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each items as entrant, index (entrant.id)}
              <Table.Row
                draggable={editable}
                ondragstart={(e) => handleDragStart(e, index)}
                ondragover={(e) => handleDragOver(e, index)}
                ondrop={(e) => handleDrop(e, index)}
                ondragend={resetDrag}
                class={[
                  "transition-colors",
                  dragIndex === index && "opacity-40",
                  overIndex === index &&
                    dragIndex !== null &&
                    dragIndex !== index &&
                    (dragIndex < index
                      ? "border-b-2 border-b-primary"
                      : "border-t-2 border-t-primary"),
                ]}
              >
                {#if editable}
                  <Table.Cell
                    class="cursor-grab text-muted-foreground active:cursor-grabbing"
                  >
                    <GripVerticalIcon class="size-4" />
                  </Table.Cell>
                {/if}
                <Table.Cell class="text-muted-foreground"
                  >{index + 1}</Table.Cell
                >
                <Table.Cell class="font-medium">{entrant.name}</Table.Cell>
                {#if editable}
                  <Table.Cell class="text-right">
                    <Button
                      variant="ghost"
                      size="icon"
                      aria-label={`Delete ${entrant.name}`}
                      onclick={() => (deleteTarget = entrant)}
                    >
                      <Trash2Icon class="size-4 text-destructive" />
                    </Button>
                  </Table.Cell>
                {/if}
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
      </div>
    {/if}
  </div>
</div>

{#if editable}
  <!-- create participants -->
  <Dialog.Root bind:open={createOpen}>
    <Dialog.Content class="sm:max-w-md">
      <Dialog.Header>
        <Dialog.Title>Add participants</Dialog.Title>
        <Dialog.Description
          >Enter one participant name per line.</Dialog.Description
        >
      </Dialog.Header>

      <Textarea
        bind:value={rawNames}
        rows={8}
        placeholder={"Alice\nBob\nCharlie"}
        disabled={insertParticipantsMutation.isPending}
      />
      <p class="text-xs text-muted-foreground">
        {parsedNames.length}
        {parsedNames.length === 1 ? "participant" : "participants"} will be added.
      </p>
      {#if insertParticipantsMutation.isError}
        <p class="text-sm text-destructive">
          {insertParticipantsMutation.error?.message ??
            "Failed to add participants."}
        </p>
      {/if}

      <Dialog.Footer>
        <Button
          variant="outline"
          onclick={() => (createOpen = false)}
          disabled={insertParticipantsMutation.isPending}
        >
          Cancel
        </Button>
        <Button
          onclick={submitCreate}
          disabled={insertParticipantsMutation.isPending ||
            parsedNames.length === 0}
        >
          {insertParticipantsMutation.isPending
            ? "Adding..."
            : "Add participants"}
        </Button>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Root>

  <!-- delete confirmation -->
  <Dialog.Root
    open={deleteOpen}
    onOpenChange={(open) => {
      if (!open) deleteTarget = null;
    }}
  >
    <Dialog.Content class="sm:max-w-sm">
      <Dialog.Header>
        <Dialog.Title>Delete participant</Dialog.Title>
        <Dialog.Description>
          Remove <span class="font-medium text-foreground"
            >{deleteTarget?.name}</span
          > from this tournament? This can't be undone.
        </Dialog.Description>
      </Dialog.Header>
      <Dialog.Footer>
        <Button
          variant="outline"
          onclick={() => (deleteTarget = null)}
          disabled={deleting}
        >
          Cancel
        </Button>
        <Button
          variant="destructive"
          onclick={confirmDelete}
          disabled={deleting}
        >
          {deleting ? "Deleting..." : "Delete"}
        </Button>
      </Dialog.Footer>
    </Dialog.Content>
  </Dialog.Root>
{/if}
