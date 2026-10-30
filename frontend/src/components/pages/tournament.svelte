<script lang="ts">
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
    GetTournamentActivityLogs,
  } from "$wails/go/main/App";

  import ActivityLogs from "../activity-logs.svelte";
  import BattleHandler from "../judging/battle-handler.svelte";
  import EntrantList from "../judging/entrant-list.svelte";
  import Panel from "../judging/panel.svelte";
  import QualifyingHandler from "../judging/qualifying-handler.svelte";
  import StageCard from "../judging/stage-card.svelte";
  import Matches from "../tournaments/matches.svelte";
  import RegisteredParticipants from "../tournaments/registered-participants.svelte";
  import TournamentReadyAlert from "../tournaments/tournament-ready-alert.svelte";
  import Progress from "../ui/progress/progress.svelte";

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

  const logsQuery = createQuery(() => ({
    queryKey: ["logs", tournamentID],
    enabled: !!tournamentID,
    queryFn: () => GetTournamentActivityLogs(tournamentID),
  }));

  let mounted = $state(false);

  onMount(() => {
    mounted = true;
  });

  let hidePanel = $state(false);
</script>

{#if mounted}
  <div class="space-y-6" in:fade={{ duration: 300 }}>
    <div
      class="sticky top-0 z-50 mx-auto w-full bg-background backdrop-blur supports-backdrop-filter:bg-background/80"
    >
      <header
        class="flex gap-5 px-4 pb-6 sm:flex-row sm:items-end sm:justify-between sm:px-6 lg:px-8 py-4"
      >
        <div class="flex flex-col gap-5">
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
        </div>
      </header>

      <Progress class="rounded-none" value={tournamentQuery.data?.progress} />
    </div>

    {#if tournamentQuery.data}
      {@const tournament = tournamentQuery.data}
      {@const isQualifying = tournament.status === "qualifying"}

      <div class="space-y-2 pb-4 relative">
        <aside
          class="fixed -inset-y-4 right-0 z-0 hidden w-80 overflow-y-auto lg:block border-l bg-sidebar/10"
        >
          <div class="mt-30 min-h-[calc(100vh-9rem)]">
            <ActivityLogs logs={logsQuery.data ?? []} />
          </div>
        </aside>

        <TournamentReadyAlert
          entrants={entrantsQuery.data ?? []}
          {tournament}
        />
        <RegisteredParticipants
          {tournament}
          entrants={entrantsQuery.data ?? []}
        />

        {#if tournament.status === "qualifying" || tournament.status === "elimination"}
          <div class="mr-0 lg:mr-80">
            <div class="mx-auto max-w-6xl space-y-1 px-4 sm:px-6">
              <div class="min-w-0">
                <StageCard>
                  {#if isQualifying}
                    <EntrantList
                      {tournament}
                      entrants={entrantsQuery.data ?? []}
                      bind:hidePanel
                    />
                  {:else}
                    <Matches
                      {tournamentID}
                      {stage}
                      matches={matchesQuery.data ?? []}
                      activeBattles={activeBattles.data ?? []}
                      {judgement}
                      progress={tournament.progress ?? 0}
                      finalized={tournament.finalized ?? false}
                    />
                  {/if}
                </StageCard>

                {#if !hidePanel}
                  <Panel {stage} {judgement} {tournament} {isQualifying}>
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
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
{/if}
