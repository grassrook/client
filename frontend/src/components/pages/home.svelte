<script lang="ts">
  import { createQuery } from "@tanstack/svelte-query";

  import { fade } from "svelte/transition";
  import CreateDialog from "../tournaments/create-dialog.svelte";
  import Grid from "../tournaments/grid.svelte";

  import { GetTournaments } from "$wails/go/main/App";
  import { onMount } from "svelte";
  import Separator from "../ui/separator/separator.svelte";

  let createTournamentOpen = $state(false);

  const tournamentsQuery = createQuery(() => ({
    queryKey: ["tournaments"],
    queryFn: GetTournaments,
  }));

  let mounted = $state(false);
  onMount(() => {
    mounted = true;
  });
</script>

{#if mounted}
  <div in:fade={{ duration: 300 }}>
    <header
      class="flex max-w-6xl mx-auto flex-col gap-5 pb-6 sm:flex-row sm:items-end sm:justify-between px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
    >
      <div class="space-y-2">
        <p class="text-sm font-medium text-muted-foreground">
          Tournament management
        </p>
        <h1 class="text-3xl font-semibold tracking-tight">Tournaments</h1>
        <p class="max-w-xl text-sm text-muted-foreground">
          Create a competition, manage qualifying participants, and open its
          bracket when you are ready.
        </p>
      </div>
      <CreateDialog bind:open={createTournamentOpen} />
    </header>

    <Separator />

    <Grid
      tournaments={tournamentsQuery.data ?? []}
      isLoading={tournamentsQuery.isPending}
      isError={tournamentsQuery.isError}
      error={tournamentsQuery.error}
      onRetry={() => tournamentsQuery.refetch()}
    />
  </div>
{/if}
