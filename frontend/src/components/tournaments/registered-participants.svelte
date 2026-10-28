<script lang="ts">
  /* eslint-disable svelte/no-useless-mustaches */
  import GripVerticalIcon from "@lucide/svelte/icons/grip-vertical";
  import PlusIcon from "@lucide/svelte/icons/plus";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";
  import UsersIcon from "@lucide/svelte/icons/users";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";

  import { Button } from "$src/components/ui/button/index.js";
  import * as Dialog from "$src/components/ui/dialog/index.js";
  import * as Table from "$src/components/ui/table/index.js";
  import {
    DeleteParticipant,
    GetDrivers,
    InsertBulkParticipant,
    UpdateTournament,
  } from "$wails/go/main/App";
  import type { driver, entrant, tournament } from "$wails/go/models";

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

  // the order is shuffled when the tournament starts, so there is no point in
  // letting the user arrange it by hand
  const randomized = $derived(Boolean(tournament.randomizeQualifyingOrder));
  const canReorder = $derived(editable && !randomized);

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

  const driversQuery = createQuery(() => ({
    queryKey: ["drivers"],
    queryFn: () => GetDrivers(),
  }));

  // ---------- mutations ----------
  const insertParticipantsMutation = createMutation(() => ({
    mutationFn: async ({ driverIDs }: { driverIDs: string[] }) =>
      InsertBulkParticipant(tournament.id, driverIDs),
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
    if (!canReorder) return;
    dragIndex = index;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = "move";
      e.dataTransfer.setData("text/plain", String(index));
    }
  }

  function handleDragOver(e: DragEvent, index: number) {
    if (!canReorder || dragIndex === null) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = "move";
    overIndex = index;
  }

  function handleDrop(e: DragEvent, index: number) {
    e.preventDefault();
    const from = dragIndex;
    resetDrag();
    if (!canReorder || from === null || from === index) return;

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

  // ---------- create participants ----------
  let createOpen = $state(false);
  let rawNames = $state("");
  let driverSearch = $state("");
  let selectedDrivers = $state<driver.Driver[]>([]);
  let driverSearchFocused = $state(false);

  const parsedNames = $derived(
    rawNames
      .split("\n")
      .map((name) => name.trim())
      .filter((name) => name !== ""),
  );

  const existingDriverIds = $derived(
    new Set(
      orderedEntrants
        .map((entrant) => entrant.driverId)
        .filter((id) => id !== ""),
    ),
  );

  const matchingDrivers = $derived.by(() => {
    const query = driverSearch.trim().toLowerCase();

    return (driversQuery.data ?? [])
      .filter(
        (driver) =>
          driver.name.toLowerCase().includes(query) &&
          !selectedDrivers.some((selected) => selected.id === driver.id) &&
          !existingDriverIds.has(driver.id),
      )
      .slice(0, 8);
  });

  // Combine selected drivers and manually entered names.
  // Prevent duplicate names, ignoring capitalization.
  const participantNames = $derived(
    [...selectedDrivers.map((driver) => driver.name), ...parsedNames].filter(
      (name, index, names) =>
        names.findIndex(
          (other) => other.toLowerCase() === name.toLowerCase(),
        ) === index,
    ),
  );

  let driverSearchInput = $state<HTMLInputElement | undefined>(undefined);
  function selectDriver(driver: driver.Driver) {
    selectedDrivers = [...selectedDrivers, driver];
    driverSearch = "";
    driverSearchFocused = false;
    driverSearchInput?.blur();
  }

  function removeDriver(id: string) {
    selectedDrivers = selectedDrivers.filter((driver) => driver.id !== id);
    driverSearchFocused = false;
    driverSearchInput?.blur();
  }

  function resetCreateForm() {
    rawNames = "";
    driverSearch = "";
    selectedDrivers = [];
    driverSearchFocused = false;
  }

  function submitCreate() {
    if (!editable || participantNames.length === 0) return;

    insertParticipantsMutation.mutate(
      {
        driverIDs: selectedDrivers.map((driver) => driver.id),
      },
      {
        onSuccess: () => {
          resetCreateForm();
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
  <div class="mx-auto max-w-6xl space-y-1 px-4 pb-2 sm:px-6 lg:pb-2">
    <div class="border border-border/70 rounded-xl p-4">
      <div class="flex items-end justify-between mb-4">
        <div>
          <h3 class="font-semibold">Participants</h3>
          <p class="text-sm text-muted-foreground">
            {items.length}
            {items.length === 1 ? "participant" : "participants"} registered.
            {#if editable}
              {#if randomized}
                The qualifying order will be randomized once the tournament
                starts.
              {:else}
                Drag rows to change the qualifying order.
              {/if}
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
              Add the people taking part in this tournament. You can paste
              several names at once, one per line.
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
                {#if canReorder}
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
                  draggable={canReorder}
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
                  {#if canReorder}
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
</div>

{#if editable}
  <!-- create participants -->
  <Dialog.Root bind:open={createOpen}>
    <Dialog.Content
      onOpenAutoFocus={(event) => event.preventDefault()}
      class="sm:max-w-md"
    >
      <Dialog.Header>
        <Dialog.Title>Add participants</Dialog.Title>
        <Dialog.Description
          >Enter one participant name per line.</Dialog.Description
        >
      </Dialog.Header>

      <div class="space-y-4">
        <div class="space-y-2">
          <label class="text-sm font-medium mb-2" for="driver-search">
            Registered drivers
          </label>

          <div class="relative mt-2">
            <input
              id="driver-search"
              bind:this={driverSearchInput}
              bind:value={driverSearch}
              onfocus={() => (driverSearchFocused = true)}
              onblur={() => (driverSearchFocused = false)}
              placeholder="Search drivers..."
              disabled={insertParticipantsMutation.isPending}
              class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring"
            />

            {#if driverSearchFocused}
              <div
                class="absolute inset-x-0 top-full z-50 mt-1 max-h-56 overflow-y-auto rounded-md border bg-popover p-1 text-popover-foreground shadow-md"
              >
                {#if driversQuery.isPending}
                  <p class="px-3 py-2 text-sm text-muted-foreground">
                    Loading drivers...
                  </p>
                {:else if matchingDrivers.length > 0}
                  {#each matchingDrivers as driver (driver.id)}
                    <button
                      type="button"
                      class="flex w-full rounded-sm px-3 py-2 text-left text-sm hover:bg-accent hover:text-accent-foreground"
                      onclick={() => {
                        selectDriver(driver);
                      }}
                      onmousedown={(event) => event.preventDefault()}
                    >
                      {driver.name}
                    </button>
                  {/each}
                {:else}
                  <p class="px-3 py-2 text-sm text-muted-foreground">
                    No matching drivers.
                  </p>
                {/if}
              </div>
            {/if}
          </div>

          {#if selectedDrivers.length > 0}
            <div class="flex flex-wrap gap-2 pt-1">
              {#each selectedDrivers as driver (driver.id)}
                <span
                  class="inline-flex items-center gap-2 rounded-md border bg-secondary px-2 py-1 text-sm"
                >
                  {driver.name}
                  <button
                    type="button"
                    aria-label={`Remove ${driver.name}`}
                    onmousedown={(event) => event.preventDefault()}
                    onclick={() => removeDriver(driver.id)}
                  >
                    ×
                  </button>
                </span>
              {/each}
            </div>
          {/if}
        </div>

        <p class="text-xs text-muted-foreground">
          {participantNames.length}
          {participantNames.length === 1 ? "participant" : "participants"} will be
          added.
        </p>
      </div>

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
            participantNames.length === 0}
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
