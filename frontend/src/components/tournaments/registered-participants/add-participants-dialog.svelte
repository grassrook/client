<script lang="ts">
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";

  import { Button } from "$src/components/ui/button/index.js";
  import * as Dialog from "$src/components/ui/dialog/index.js";
  import { GetDrivers, InsertBulkParticipant } from "$wails/go/main/App";
  import type { driver } from "$wails/go/models";

  type Props = {
    open: boolean;
    tournamentId: string;
    existingDriverIds: Set<string>;
  };

  let {
    open = $bindable(false),
    tournamentId,
    existingDriverIds,
  }: Props = $props();

  const queryClient = useQueryClient();

  const driversQuery = createQuery(() => ({
    queryKey: ["drivers"],
    queryFn: () => GetDrivers(),
  }));

  const insertParticipantsMutation = createMutation(() => ({
    mutationFn: async ({ driverIDs }: { driverIDs: string[] }) =>
      InsertBulkParticipant(tournamentId, driverIDs),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["entrants", tournamentId],
      });
    },
  }));

  let driverSearch = $state("");
  let selectedDrivers = $state<driver.Driver[]>([]);
  let driverSearchFocused = $state(false);
  let driverSearchInput = $state<HTMLInputElement | undefined>(undefined);

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

  function resetForm() {
    driverSearch = "";
    selectedDrivers = [];
    driverSearchFocused = false;
  }

  function submit() {
    if (selectedDrivers.length === 0) return;

    insertParticipantsMutation.mutate(
      { driverIDs: selectedDrivers.map((driver) => driver.id) },
      {
        onSuccess: () => {
          resetForm();
          open = false;
        },
      },
    );
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content
    onOpenAutoFocus={(event) => event.preventDefault()}
    class="sm:max-w-md"
  >
    <Dialog.Header>
      <Dialog.Title>Add participants</Dialog.Title>
      <Dialog.Description>
        Search and select the registered drivers to add.
      </Dialog.Description>
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
                    onclick={() => selectDriver(driver)}
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
        {selectedDrivers.length}
        {selectedDrivers.length === 1 ? "participant" : "participants"} will be added.
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
        onclick={() => (open = false)}
        disabled={insertParticipantsMutation.isPending}
      >
        Cancel
      </Button>
      <Button
        onclick={submit}
        disabled={insertParticipantsMutation.isPending ||
          selectedDrivers.length === 0}
      >
        {insertParticipantsMutation.isPending
          ? "Adding..."
          : "Add participants"}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
