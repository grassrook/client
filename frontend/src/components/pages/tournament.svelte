<script lang="ts">
  // TOOD (important): LOADING STATE FOR tournament, matches, activeBattles, entrants
  import { LucideArrowLeft } from "@lucide/svelte";
  import { createQuery } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import { link } from "svelte-spa-router";

  import { Judgement } from "$src/lib/judging.svelte";
  import { Stage } from "$src/lib/stage.svelte";
  import {
    GetActiveMatches,
    GetEntrants,
    GetMatches,
    GetTournament,
  } from "$wails/go/main/App";

  import BattleHandler from "../judging/battle-handler.svelte";
  import EntrantList from "../judging/entrant-list.svelte";
  import Panel from "../judging/panel.svelte";
  import QualifyingHandler from "../judging/qualifying-handler.svelte";
  import StageCard from "../judging/stage-card.svelte";
  import Matches from "../tournaments/matches.svelte";
  import Progress from "../ui/progress/progress.svelte";
  import Separator from "../ui/separator/separator.svelte";

  let { params = {} } = $props();
  const tournamentID = $derived(params.id);

  let stage = new Stage();
  let judgement = new Judgement();

  const tournamentQuery = createQuery(() => ({
    queryKey: ["tournament", tournamentID],
    enabled: !!tournamentID,
    queryFn: async () => GetTournament(tournamentID),
  }));

  const matchesQuery = createQuery(() => ({
    queryKey: ["matches", tournamentID],
    enabled: !!tournamentID,
    queryFn: async (): Promise<Tournament.GroupedMatches> =>
      GetMatches(tournamentID),
  }));

  const activeBattles = createQuery(() => ({
    queryKey: ["activeBattles", tournamentID],
    enabled: !!tournamentID,
    queryFn: async () => GetActiveMatches(tournamentID),
  }));

  const entrantsQuery = createQuery(() => ({
    queryKey: ["entrants", tournamentID],
    enabled: !!tournamentID,
    queryFn: () => GetEntrants(tournamentID),
  }));

  let mounted = $state(false);
  onMount(() => {
    mounted = true;
  });

  let hidePanel = $state(false);
</script>

{#if mounted}
  <div in:fade={{ duration: 300 }}>
    <header
      class="flex max-w-6xl mx-auto flex-col gap-5 pb-6 sm:flex-row sm:items-end sm:justify-between px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
    >
      <div class="space-y-2">
        <a
          use:link
          href="/"
          class="inline-flex items-center gap-1 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <LucideArrowLeft class="size-4" />
          Back to tournament list
        </a>
        <h1 class="text-3xl font-semibold tracking-tight">
          {tournamentQuery.data?.name ?? "Loading..."}
        </h1>
      </div>
    </header>

    <Progress class="rounded-none" value={tournamentQuery.data?.progress} />

    <Separator />
    {#if tournamentQuery.data}
      {@const tournament = tournamentQuery.data}
      {@const isQualifying = tournament.status === "qualifying"}
      <div
        class="mx-auto max-w-6xl space-y-1 px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
      >
        <StageCard {isQualifying}>
          {#snippet stageContent({ setHasRoundedBottom })}
            {#if isQualifying}
              <EntrantList
                tournament={tournamentQuery.data ?? {}}
                entrants={entrantsQuery.data ?? []}
                {setHasRoundedBottom}
                bind:hidePanel
              />
            {:else}
              <Matches
                {tournamentID}
                {stage}
                matches={matchesQuery.data ?? []}
                activeBattles={activeBattles.data ?? []}
                {judgement}
                progress={tournamentQuery.data.progress ?? 0}
                finalized={tournamentQuery.data.finalized ?? false}
                {setHasRoundedBottom}
              />
            {/if}
          {/snippet}
        </StageCard>

        {#if !hidePanel}
          <Panel
            {stage}
            {judgement}
            tournament={tournamentQuery.data ?? {}}
            {isQualifying}
          >
            {#if isQualifying}
              <QualifyingHandler
                {tournament}
                {judgement}
                entrants={entrantsQuery.data ?? []}
              />
            {:else}
              <BattleHandler {tournamentID} {judgement} {stage} />
            {/if}
          </Panel>
        {/if}
      </div>
    {/if}
  </div>
{/if}
