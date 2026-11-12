<script lang="ts">
  import CopyIcon from "@lucide/svelte/icons/copy";
  import PlusIcon from "@lucide/svelte/icons/plus";
  import { createMutation, useQueryClient } from "@tanstack/svelte-query";

  import { Button } from "$src/components/ui/button/index.js";
  import { UpdateTournament } from "$wails/go/main/App";
  import type { entrant, tournament } from "$wails/go/models";

  import AddParticipantsDialog from "./registered-participants/add-participants-dialog.svelte";
  import DeleteParticipantDialog from "./registered-participants/delete-participant-dialog.svelte";
  import {
    initialOrderNames,
    orderEntrants,
  } from "./registered-participants/order";
  import ParticipantsEmptyState from "./registered-participants/participants-empty-state.svelte";
  import ParticipantsTable from "./registered-participants/participants-table.svelte";
  import ReplaceParticipantDialog from "./registered-participants/replace-participant-dialog.svelte";

  type Props = {
    tournament: tournament.Tournament;
    entrants: entrant.Entrant[];
    matches: Tournament.GroupedMatches;
    onDelete?: (entrant: entrant.Entrant) => void | Promise<void>;
  };

  let { tournament, entrants, matches }: Props = $props();

  const queryClient = useQueryClient();

  // participants can only be added, removed or reordered before the
  // tournament has started
  const editable = $derived(tournament.status === "pending");

  // the order is shuffled when the tournament starts, so there is no point in
  // letting the user arrange it by hand
  const randomized = $derived(Boolean(tournament.randomizeQualifyingOrder));
  const canReorder = $derived(editable && !randomized);

  const orderedEntrants = $derived(orderEntrants(entrants, tournament));

  // local copy so rows can be rearranged immediately; resyncs whenever the
  // entrants or the tournament's qualifying order change
  // eslint-disable-next-line svelte/prefer-writable-derived
  let items = $state<entrant.Entrant[]>([]);
  $effect(() => {
    items = [...orderedEntrants];
  });

  const existingDriverIds = $derived(
    new Set(
      orderedEntrants
        .map((entrant) => entrant.driverId)
        .filter((id) => id !== ""),
    ),
  );

  // ---------- bracket / replace ----------
  const bracketParticipantIds = $derived(
    new Set(
      matches
        .flat(2)
        .map((match) => match.participantId)
        .filter((id): id is string => !!id),
    ),
  );

  // only participants who are NOT in the bracket can be swapped in
  const replacementOptions = $derived(
    items.filter((e) => !bracketParticipantIds.has(e.participantId)),
  );

  const canReplace = $derived(
    tournament.status === "elimination" &&
      items.length > 1 &&
      items.length > tournament.size,
  );

  // ---------- reorder ----------
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

  function handleReorder(next: entrant.Entrant[]) {
    items = next;
    updateTournamentMutation.mutate({
      qualifyingOrder: next.map((entrant) => entrant.participantId),
    });
  }

  // ---------- dialogs ----------
  let createOpen = $state(false);
  let deleteTarget = $state<entrant.Entrant | null>(null);
  let replaceTarget = $state<entrant.Entrant | null>(null);

  async function copyInitialQualifyingOrder() {
    const names = initialOrderNames(tournament, entrants);
    if (names.length === 0) return;

    await navigator.clipboard.writeText(names.join("\n"));
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

        <div class="flex items-center gap-2">
          {#if tournament.status !== "elimination" && tournament.initialQualifyingOrder}
            <Button
              variant="outline"
              size="sm"
              onclick={copyInitialQualifyingOrder}
            >
              <CopyIcon class="mr-2 size-4" />
              Copy qualifying order
            </Button>
          {/if}

          {#if editable}
            <Button onclick={() => (createOpen = true)}>
              <PlusIcon class="mr-2 size-4" />
              Add participants
            </Button>
          {/if}
        </div>
      </div>

      {#if items.length === 0}
        <ParticipantsEmptyState {editable} onAdd={() => (createOpen = true)} />
      {:else}
        <ParticipantsTable
          {tournament}
          {items}
          {editable}
          {canReorder}
          {canReplace}
          {bracketParticipantIds}
          onReorder={handleReorder}
          onReplace={(entrant) => (replaceTarget = entrant)}
          onDelete={(entrant) => (deleteTarget = entrant)}
        />
      {/if}
    </div>
  </div>
</div>

{#if editable}
  <AddParticipantsDialog
    bind:open={createOpen}
    tournamentId={tournament.id}
    {existingDriverIds}
  />

  <DeleteParticipantDialog
    tournamentId={tournament.id}
    bind:target={deleteTarget}
  />
{/if}

{#if canReplace}
  <ReplaceParticipantDialog
    tournamentId={tournament.id}
    bind:target={replaceTarget}
    options={replacementOptions}
  />
{/if}
