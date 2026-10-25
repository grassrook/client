<script lang="ts">
  import CircleCheckIcon from "@lucide/svelte/icons/circle-check";
  import PlayIcon from "@lucide/svelte/icons/play";
  import { createMutation, useQueryClient } from "@tanstack/svelte-query";

  import * as Alert from "$src/components/ui/alert/index.js";
  import * as AlertDialog from "$src/components/ui/alert-dialog/index.js";
  import { Button } from "$src/components/ui/button/index.js";
  import { StartTournament } from "$wails/go/main/App";
  import type { entrant, tournament } from "$wails/go/models";

  type Props = {
    tournament: tournament.Tournament;
    entrants: entrant.Entrant[];
  };

  let { tournament, entrants }: Props = $props();

  const queryClient = useQueryClient();

  // ready once every bracket slot has an entrant, and only while the
  // tournament hasn't been started yet
  const ready = $derived(
    tournament.status === "pending" && tournament.size <= entrants.length,
  );

  const randomized = $derived(Boolean(tournament.randomizeQualifyingOrder));

  let confirmOpen = $state(false);

  const startTournamentMutation = createMutation(() => ({
    mutationFn: async () => StartTournament(tournament.id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["tournament", tournament.id],
      });

      await queryClient.invalidateQueries({
        queryKey: ["logs", tournament.id],
      });
    },
  }));

  // called when the user confirms in the dialog
  function start() {
    if (!ready) return;
    startTournamentMutation.mutate();
  }
</script>

{#if ready}
  <div class="mr-0 lg:mr-80">
    <div class="mx-auto max-w-6xl space-y-1 px-4 pb-2 sm:px-6 lg:pb-2">
      <Alert.Root>
        <CircleCheckIcon />
        <Alert.Title>This tournament is ready to start</Alert.Title>
        <Alert.Description>
          <p>
            All {tournament.size} spots are filled with {entrants.length}
            {entrants.length === 1 ? "participant" : "participants"}.
            {#if randomized}
              <!-- // -->
            {:else}
              Qualifying will begin in the current order.
            {/if}
          </p>
          {#if startTournamentMutation.isError}
            <p class="mt-2 text-destructive">
              {startTournamentMutation.error?.message ??
                "Failed to start the tournament."}
            </p>
          {/if}
        </Alert.Description>
        <Alert.Action>
          <Button
            size="xs"
            onclick={() => (confirmOpen = true)}
            disabled={startTournamentMutation.isPending}
          >
            <PlayIcon class="mr-2 size-4" />
            {startTournamentMutation.isPending ? "Starting..." : "Start"}
          </Button>
        </Alert.Action>
      </Alert.Root>
    </div>
  </div>

  <!-- confirm start -->
  <AlertDialog.Root bind:open={confirmOpen}>
    <AlertDialog.Content>
      <AlertDialog.Header>
        <AlertDialog.Title>Start the tournament?</AlertDialog.Title>
        <AlertDialog.Description>
          Qualifying will begin right away.
          {#if randomized}
            The qualifying order will be randomized.
          {/if}
          Once started, participants can no longer be added, removed or reordered.
          This can't be undone.
        </AlertDialog.Description>
      </AlertDialog.Header>
      <AlertDialog.Footer>
        <AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
        <AlertDialog.Action onclick={start}>
          Start the tournament
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}
