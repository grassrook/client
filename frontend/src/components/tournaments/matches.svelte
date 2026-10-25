<script lang="ts">
  /* eslint-disable svelte/require-each-key */
  import { createMutation, useQueryClient } from "@tanstack/svelte-query";
  import { capitalize } from "es-toolkit/string";

  import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "$src/components/ui/alert-dialog";
  import { Badge } from "$src/components/ui/badge";
  import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from "$src/components/ui/tabs";
  import type { Judgement } from "$src/lib/judging.svelte";
  import { Stage } from "$src/lib/stage.svelte";
  import {
    ConcludeBattle,
    CreateTournamentActivityLog,
    ResetBattle,
    SyncTournament,
    UpdateNote,
    UpdateTournamentState,
  } from "$wails/go/main/App";
  import { type repository } from "$wails/go/models";

  import Button from "../ui/button/button.svelte";
  import * as ButtonGroup from "../ui/button-group";
  import Empty from "../ui/empty/empty.svelte";
  import EmptyDescription from "../ui/empty/empty-description.svelte";
  import EmptyHeader from "../ui/empty/empty-header.svelte";
  import EmptyTitle from "../ui/empty/empty-title.svelte";

  type Battles = repository.ActiveBattle[];

  let {
    tournamentID,
    matches,
    judgement,
    stage,
    activeBattles,
    progress,
    finalized,
  }: {
    tournamentID: string;
    matches: Tournament.GroupedMatches;
    judgement: Judgement;
    stage: Stage;
    activeBattles: Battles;
    progress: number;
    finalized: boolean;
  } = $props();

  const queryClient = useQueryClient();

  let selectedMatchId = $state<string | undefined>(undefined);

  function getLead(
    p1Battle: Battles[number],
    p2Battle: Battles[number],
  ): "p1" | "p2" {
    if (p1Battle.leadStatus === -1) return "p2";
    if (p2Battle.leadStatus === -1) return "p1";

    return p1Battle.leadStatus === 1 ? "p1" : "p2";
  }

  function onStartMatch(match: repository.Match[]) {
    stage.clear();
    judgement.clear();

    selectedMatchId = match[0].matchId;
    stage.matchId = selectedMatchId;

    const p1 = match[0];
    const p2 = match[1];

    if (!p1.participantId || !p2.participantId) {
      console.warn("missing participant id");
      return;
    }
    const battles = activeBattles.filter(
      (battle) => battle.matchId === p1.matchId,
    );

    if (battles.length !== 2) {
      console.warn("active battle not found or incomplete");
      return;
    }

    const p1Battle = battles.find(
      (battle) => battle.participantId === p1.participantId,
    );
    const p2Battle = battles.find(
      (battle) => battle.participantId === p2.participantId,
    );

    if (!p1Battle || !p2Battle) {
      console.warn("missing active battle participant");
      return;
    }

    const lead = getLead(p1Battle, p2Battle);

    if (!lead) {
      console.warn("no lead participant");
      return;
    }

    stage.setDrivers(p1.participantId, p2.participantId, lead);
  }

  const sortedMatches = $derived(
    [...matches]
      .sort((a, b) => {
        const roundA = a[0]?.[0]?.round ?? 0;
        const roundB = b[0]?.[0]?.round ?? 0;

        if (roundA === 0) return 1;
        if (roundB === 0) return -1;

        return roundA - roundB;
      })
      .map((round) =>
        round.map((match) => {
          const battles = activeBattles.filter(
            (battle) => battle.matchId === match[0]?.matchId,
          );
          const leadingBattle = battles.find(
            (battle) => battle.leadStatus === 1,
          );
          return {
            match,
            leadingParticipantId: leadingBattle?.participantId ?? null,
          };
        }),
      ),
  );

  const updateNoteMutation = createMutation(() => ({
    mutationFn: (text: string) => UpdateNote(stage.matchId!, text),
    onError: (error) => {
      console.error("Failed to update note:", error);
    },
  }));

  async function onHandleOMT(match: repository.Match[]) {
    await updateNoteMutation.mutateAsync("");
    await ResetBattle(tournamentID, selectedMatchId!);

    await queryClient.invalidateQueries({
      queryKey: ["note", tournamentID, stage.matchId],
    });

    await queryClient.invalidateQueries({
      queryKey: ["activeBattles", tournamentID],
    });

    await queryClient.invalidateQueries({
      queryKey: ["matches", tournamentID],
    });
    await queryClient.invalidateQueries({
      queryKey: ["logs", tournamentID],
    });

    onStartMatch(match);
  }

  async function onHandleConclude() {
    await ConcludeBattle(tournamentID, selectedMatchId!);
    await SyncTournament(tournamentID);

    await queryClient.invalidateQueries({
      queryKey: ["tournament", tournamentID],
    });

    await queryClient.invalidateQueries({
      queryKey: ["activeBattles", tournamentID],
    });

    await queryClient.invalidateQueries({
      queryKey: ["matches", tournamentID],
    });

    await queryClient.invalidateQueries({
      queryKey: ["logs", tournamentID],
    });

    stage.clear();
    selectedMatchId = undefined;
  }

  let isProcessing = $state(false);
  let omtDialogOpen = $state(false);
  let concludeDialogOpen = $state(false);

  async function handleOMT(match: repository.Match[]) {
    isProcessing = true;

    try {
      await onHandleOMT(match);
      omtDialogOpen = false;
    } finally {
      isProcessing = false;
    }
  }

  async function handleConclude() {
    isProcessing = true;

    try {
      await onHandleConclude();
      concludeDialogOpen = false;
    } finally {
      isProcessing = false;
    }
  }

  function roundKeys() {
    let keys: string[] = [];
    for (let i = 0; i < sortedMatches.length; i++) {
      const matches = sortedMatches[i].length;
      if (matches > 2) {
        keys.push(`Round ${i + 1}`);
      } else {
        if (matches == 2) {
          keys.push("Semifinal");
        } else if (matches == 1) {
          if (sortedMatches[i][0].match[0].round === 0) {
            keys.push("Third Place");
            continue;
          }
          keys.push("Final");
        }
      }
    }
    return keys;
  }

  let finalizeDialogOpen = $state(false);

  async function handleFinalize() {
    isProcessing = true;

    try {
      await UpdateTournamentState(tournamentID, "finalize");
      await CreateTournamentActivityLog({
        tournament_id: tournamentID,
        content: `Tournament marked as finished`,
        logType: "TOURNAMENT_FINISHED",
      });
      await queryClient.invalidateQueries({
        queryKey: ["tournament", tournamentID],
      });
      finalizeDialogOpen = false;
    } finally {
      isProcessing = false;
    }
  }

  function clearMatchSelection() {
    stage.clear();
    judgement.clear();
    selectedMatchId = undefined;
  }
</script>

<div>
  <h3 class="font-semibold">Bracket</h3>
  <p class="text-sm text-muted-foreground">Drift battle matches by round</p>
</div>

<Tabs value="round-1">
  {#if sortedMatches.length === 0}
    <Empty class="min-h-48 border-dashed py-10">
      <EmptyHeader>
        <EmptyTitle>No bracket matches yet</EmptyTitle>
        <EmptyDescription>
          Matches will appear here after qualifying is concluded and the bracket
          is synced.
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  {:else}
    <TabsList class="overflow-x justify-start gap-1 overscroll-x-contain">
      {#each sortedMatches as round, i}
        {@const roundNumber = round[0]?.match[0]?.round ?? 0}

        {#if roundNumber === 0}
          <TabsTrigger
            onclick={clearMatchSelection}
            class="min-w-fit flex-none"
            value="third-place">3rd Place</TabsTrigger
          >
        {:else}
          <TabsTrigger
            onclick={clearMatchSelection}
            class="min-w-fit flex-none"
            value={`round-${roundNumber}`}
          >
            {roundKeys()[i]}
          </TabsTrigger>
        {/if}
      {/each}
    </TabsList>
  {/if}

  {#each sortedMatches as round, i}
    {@const roundNumber = round[0]?.match[0]?.round ?? 0}
    {@const tabValue =
      roundNumber === 0 ? "third-place" : `round-${roundNumber}`}

    <TabsContent value={tabValue} class="mt-4 space-y-3">
      <div>
        <h4 class="font-medium">
          {roundKeys()[i]}
        </h4>

        <p class="text-sm text-muted-foreground">
          {round.length}
          {round.length === 1 ? "battle" : "battles"}
        </p>
      </div>

      <div class="grid gap-3">
        {#each round as matchData}
          {@const match = matchData.match}
          {@const currentMatch = match[0]}
          {@const hasBothParticipants =
            match[0].participantId && match[1].participantId}

          <div
            class={`border border-border/80 rounded-lg transition-colors hover:border-primary/60 hover:bg-muted/50 ${
              selectedMatchId === currentMatch.matchId
                ? "border-primary/60 bg-muted/50"
                : ""
            }`}
          >
            <div class="border-b px-4 py-3">
              <div class="flex items-center justify-between">
                <div class="text-sm font-semibold">
                  Match {round.indexOf(matchData) + 1}
                </div>

                <div class="flex items-center gap-2">
                  {#if currentMatch.status !== "open"}
                    <Badge variant="secondary">
                      {capitalize(currentMatch.status)}
                    </Badge>
                  {/if}
                </div>
              </div>
            </div>

            <div class="px-0">
              <button
                type="button"
                class="block w-full cursor-pointer text-left"
                disabled={currentMatch.status !== "open" ||
                  !hasBothParticipants}
                onclick={() => {
                  if (currentMatch.status === "open" && hasBothParticipants) {
                    if (selectedMatchId !== currentMatch.matchId) {
                      onStartMatch(match);
                    } else {
                      clearMatchSelection();
                    }
                  }
                }}
              >
                <div class="mx-4 flex items-center gap-3 pt-3 pb-3">
                  {#each match as entrant, idx}
                    {#if idx === 1}
                      <span
                        class="shrink-0 text-xs uppercase text-muted-foreground"
                      >
                        VS
                      </span>
                    {/if}

                    <div
                      class={`flex min-w-0 flex-1 justify-between items-center gap-2 ${
                        idx === 1 ? "flex-row-reverse" : ""
                      }`}
                    >
                      <div
                        class={`flex items-center ${
                          idx === 1 ? "flex-row-reverse" : ""
                        }`}
                      >
                        <span class="min-w-0 truncate font-medium">
                          {entrant.participant_name ?? "TBD"}
                        </span>

                        {#if hasBothParticipants && matchData.leadingParticipantId === entrant.participantId && currentMatch.status !== "complete"}
                          <Badge variant="outline">Lead</Badge>
                        {/if}
                      </div>

                      <span
                        class={`shrink-0 text-lg font-semibold tabular-nums ${
                          entrant.participant_name && entrant.winnerId
                            ? entrant.winnerId ===
                              entrant.challongeParticipantId
                              ? "text-green-500"
                              : "text-muted-foreground/30"
                            : ""
                        }`}
                      >
                        {entrant.score}
                      </span>
                    </div>
                  {/each}
                </div>
              </button>

              {#if currentMatch.bothRunFinished && selectedMatchId === currentMatch.matchId}
                <ButtonGroup.Root
                  orientation="horizontal"
                  class="flex justify-end w-full pr-2 py-2"
                  onclick={(event) => event.stopPropagation()}
                >
                  <AlertDialog bind:open={omtDialogOpen}>
                    <AlertDialogTrigger>
                      <Button
                        variant="outline"
                        class="rounded-r-none"
                        disabled={isProcessing}>OMT</Button
                      >
                    </AlertDialogTrigger>

                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Start OMT?</AlertDialogTitle>
                        <AlertDialogDescription>
                          This will reset the battle and start an OMT. Are you
                          sure you want to continue?
                        </AlertDialogDescription>
                      </AlertDialogHeader>

                      <AlertDialogFooter>
                        <AlertDialogCancel disabled={isProcessing}>
                          Cancel
                        </AlertDialogCancel>

                        <AlertDialogAction
                          disabled={isProcessing}
                          onclick={() => handleOMT(match)}
                        >
                          {#if isProcessing}
                            Processing...
                          {:else}
                            Continue
                          {/if}
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>

                  <AlertDialog bind:open={concludeDialogOpen}>
                    <AlertDialogTrigger>
                      <Button
                        variant="outline"
                        class="-ml-px rounded-l-none"
                        disabled={isProcessing}>Conclude</Button
                      >
                    </AlertDialogTrigger>

                    <AlertDialogContent>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Conclude match?</AlertDialogTitle>
                        <AlertDialogDescription>
                          This will finalize the result of this match. Are you
                          sure?
                        </AlertDialogDescription>
                      </AlertDialogHeader>

                      <AlertDialogFooter>
                        <AlertDialogCancel disabled={isProcessing}>
                          Cancel
                        </AlertDialogCancel>

                        <AlertDialogAction
                          disabled={isProcessing}
                          onclick={() => handleConclude()}
                        >
                          {#if isProcessing}
                            Processing...
                          {:else}
                            Conclude
                          {/if}
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </ButtonGroup.Root>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    </TabsContent>
  {/each}
</Tabs>
{#if progress >= 100 && !finalized}
  <div class="w-full pt-4">
    <AlertDialog bind:open={finalizeDialogOpen}>
      <AlertDialogTrigger class="w-full">
        <Button class="w-full" variant="default" disabled={isProcessing}
          >Mark as Done</Button
        >
      </AlertDialogTrigger>

      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Mark tournament as done?</AlertDialogTitle>
          <AlertDialogDescription>
            This will finalize the tournament. Make sure all matches have been
            concluded before continuing. This action may not be reversible.
          </AlertDialogDescription>
        </AlertDialogHeader>

        <AlertDialogFooter>
          <AlertDialogCancel disabled={isProcessing}>Cancel</AlertDialogCancel>

          <AlertDialogAction disabled={isProcessing} onclick={handleFinalize}>
            {#if isProcessing}
              Finalizing...
            {:else}
              Confirm
            {/if}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  </div>
{/if}
