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

  async function onHandleOMT(
    match: repository.Match[],
    key: string,
    matchNumber: number,
  ) {
    await updateNoteMutation.mutateAsync("");
    await ResetBattle(tournamentID, selectedMatchId!);

    await CreateTournamentActivityLog({
      tournament_id: tournamentID,
      content: `${key} Match #${matchNumber} OMT`,
      logType: "OMT",
    });

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

  async function handleOMT(
    match: repository.Match[],
    key: string,
    matchNumber: number,
  ) {
    isProcessing = true;

    try {
      await onHandleOMT(match, key, matchNumber);
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
    <TabsList
      variant="line"
      class="overflow-x justify-start gap-1 overscroll-x-contain"
    >
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
          {@const roundKey = roundKeys()[i]}
          {@const matchNumber = round.indexOf(matchData) + 1}
          {@const isSelected = selectedMatchId === currentMatch.matchId}
          {@const canLoad =
            currentMatch.status === "open" && !!hasBothParticipants}
          {@const showActions = !!currentMatch.bothRunFinished && isSelected}
          {@const score1 = Number(match[0].score ?? 0)}
          {@const score2 = Number(match[1].score ?? 0)}
          {@const totalScore = score1 + score2}
          {@const isOngoing = currentMatch.status !== "complete"}
          {@const isTied =
            !!currentMatch.bothRunFinished && isOngoing && score1 === score2}
          {@const winner1 = !!(
            match[0].participant_name &&
            match[0].winnerId &&
            match[0].winnerId === match[0].challongeParticipantId
          )}
          {@const winner2 = !!(
            match[1].participant_name &&
            match[1].winnerId &&
            match[1].winnerId === match[1].challongeParticipantId
          )}
          {@const bar1 =
            totalScore === 0
              ? "bg-muted"
              : winner1
                ? "bg-green-500"
                : score1 >= score2
                  ? "bg-foreground/80"
                  : "bg-muted-foreground/40"}
          {@const bar2 =
            totalScore === 0
              ? "bg-muted"
              : winner2
                ? "bg-green-500"
                : score2 >= score1
                  ? "bg-foreground/80"
                  : "bg-muted-foreground/40"}
          {@const marginText =
            totalScore === 0
              ? "No scores yet"
              : score1 === score2
                ? "Level"
                : `${match[score1 > score2 ? 0 : 1].participant_name ?? "TBD"} by ${Math.abs(score1 - score2)}`}

          <div
            class={`rounded-lg border border-border/80 transition-colors ${
              isSelected ? "bg-muted/40" : ""
            }`}
          >
            <div class="border-b px-4 py-3">
              <div class="flex items-center justify-between">
                <div class="text-sm font-semibold">
                  Match {matchNumber}
                </div>

                <div class="flex items-center gap-2">
                  <!-- PLACEHOLDER: run progress pips. Per-run status is not wired yet; both fill in once bothRunFinished is true. -->
                  {#each ["R1", "R2"] as pip}
                    <span
                      class={`inline-flex h-5 items-center rounded-full border px-2 text-[11px] font-semibold ${
                        currentMatch.bothRunFinished
                          ? "border-foreground/80 bg-foreground text-background"
                          : "border-border text-muted-foreground"
                      }`}
                    >
                      {pip}
                    </span>
                  {/each}
                  {#if isSelected}
                    <Badge variant="secondary" class="gap-1.5">
                      <span class="size-1.5 rounded-full bg-foreground"></span>
                      Loaded
                    </Badge>
                  {/if}
                  {#if currentMatch.status !== "open"}
                    <Badge variant="secondary">
                      {capitalize(currentMatch.status)}
                    </Badge>
                  {/if}
                </div>
              </div>
            </div>

            <div class="px-0">
              <!-- Entrants are display only now. Loading a match is done with the explicit button in the footer. -->
              <div>
                <div class="px-4 pt-4 pb-4">
                  <div class="relative grid grid-cols-2 gap-2.5">
                    {#each match as entrant, idx}
                      {@const isRight = idx === 1}
                      {@const myScore = idx === 0 ? score1 : score2}
                      {@const otherScore = idx === 0 ? score2 : score1}
                      {@const hasResult = !!(
                        entrant.participant_name && entrant.winnerId
                      )}
                      {@const isWinner =
                        hasResult &&
                        entrant.winnerId === entrant.challongeParticipantId}
                      {@const isLead =
                        hasBothParticipants &&
                        matchData.leadingParticipantId ===
                          entrant.participantId &&
                        currentMatch.status !== "complete"}
                      {@const isChase =
                        hasBothParticipants &&
                        matchData.leadingParticipantId !== null &&
                        matchData.leadingParticipantId !==
                          entrant.participantId &&
                        currentMatch.status !== "complete"}
                      {@const isLeading =
                        hasBothParticipants &&
                        isOngoing &&
                        myScore > otherScore}

                      <div
                        class={`flex min-w-0 flex-col gap-2.5 rounded-xl border p-3.5 ${
                          isRight
                            ? "items-end pl-7 text-right"
                            : "items-start pr-7"
                        } ${
                          isWinner
                            ? "border-green-500/60 bg-green-500/5"
                            : isLeading
                              ? "border-foreground/25 bg-muted/60"
                              : "border-border/60 bg-muted/20"
                        }`}
                      >
                        <div
                          class={`flex w-full min-w-0 items-center gap-2 ${
                            isRight ? "flex-row-reverse" : ""
                          }`}
                        >
                          <span
                            class={`min-w-0 truncate font-medium ${
                              hasResult && !isWinner
                                ? "text-muted-foreground"
                                : ""
                            }`}
                          >
                            {entrant.participant_name ?? "TBD"}
                          </span>

                          {#if isLead}
                            <Badge
                              variant="outline"
                              class="border-sky-300/60 text-[11px] font-bold tracking-wider text-sky-300 uppercase"
                              >Lead</Badge
                            >
                          {:else if isChase}
                            <Badge
                              variant="outline"
                              class="border-violet-300/60 text-[11px] font-bold tracking-wider text-violet-300 uppercase"
                              >Chase</Badge
                            >
                          {/if}
                        </div>

                        <span
                          class={`text-5xl leading-none font-semibold tracking-tight tabular-nums ${
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

                        <div
                          class={`flex min-h-5 w-full ${
                            isRight ? "justify-end" : "justify-start"
                          }`}
                        >
                          {#if isWinner}
                            <Badge
                              variant="outline"
                              class="border-green-500/40 bg-green-500/10 text-green-500"
                              >Winner</Badge
                            >
                          {:else if isTied && hasBothParticipants}
                            <Badge
                              variant="outline"
                              class="border-amber-400/50 bg-amber-400/10 text-amber-400"
                              >Tied</Badge
                            >
                          {:else if isLeading}
                            <Badge variant="secondary"
                              >+{myScore - otherScore} ahead</Badge
                            >
                          {/if}
                        </div>

                        <!-- PLACEHOLDER: per-run breakdown. Run scores and roles are not wired yet. -->
                        <div class="flex w-full flex-col gap-0.5 border-t pt-2">
                          {#each [1, 2] as runNumber}
                            <div
                              class={`flex items-center justify-between gap-2 rounded-md px-1.5 py-1 text-xs ${
                                isRight ? "flex-row-reverse" : ""
                              }`}
                            >
                              <span class="text-muted-foreground">
                                Run {runNumber}
                                <span class="font-semibold">—</span>
                              </span>
                              <span
                                class="font-semibold text-muted-foreground tabular-nums"
                                >—</span
                              >
                            </div>
                          {/each}
                        </div>
                      </div>
                    {/each}

                    <div
                      class="pointer-events-none absolute top-1/2 left-1/2 flex size-11 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full border bg-background text-xs font-bold tracking-widest ring-4 ring-background {isTied
                        ? 'border-amber-400 text-amber-400'
                        : 'text-muted-foreground'}"
                    >
                      {isTied ? "TIE" : "VS"}
                    </div>
                  </div>

                  <div class="mt-4 flex flex-col gap-2">
                    <div class="relative flex h-1.5 gap-0.5">
                      <div
                        class={`min-w-0 rounded-l-full ${bar1}`}
                        style={`flex: ${totalScore > 0 ? score1 : 1} 1 0%`}
                      ></div>
                      <div
                        class={`min-w-0 rounded-r-full ${bar2}`}
                        style={`flex: ${totalScore > 0 ? score2 : 1} 1 0%`}
                      ></div>
                      <div
                        class="absolute -top-1 left-1/2 h-3.5 w-0.5 -translate-x-1/2 rounded bg-muted-foreground/50"
                      ></div>
                    </div>
                    <div
                      class="flex justify-between gap-2 text-xs text-muted-foreground tabular-nums"
                    >
                      <span>
                        {totalScore > 0
                          ? `${Math.round((score1 / totalScore) * 100)}%`
                          : "—"}
                      </span>
                      <span class="min-w-0 truncate">{marginText}</span>
                      <span>
                        {totalScore > 0
                          ? `${Math.round((score2 / totalScore) * 100)}%`
                          : "—"}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {#if canLoad || showActions}
                <div
                  class="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3"
                >
                  <!-- PLACEHOLDER: helper copy is static text for now. -->
                  <p
                    class={`min-w-0 flex-1 basis-56 text-xs ${
                      showActions && isTied
                        ? "text-amber-400"
                        : "text-muted-foreground"
                    }`}
                  >
                    {showActions
                      ? isTied
                        ? "Scores are level. Call OMT to run it again."
                        : "Both runs judged. Conclude to finalize the result."
                      : isSelected
                        ? "Loaded for judging."
                        : "Load this match to start judging."}
                  </p>

                  <div class="flex flex-wrap items-center justify-end gap-2">
                    {#if canLoad}
                      <Button
                        type="button"
                        variant={isSelected ? "outline" : "secondary"}
                        aria-pressed={isSelected}
                        class="focus-visible:border-border focus-visible:ring-2 focus-visible:ring-muted-foreground/30"
                        onclick={() => {
                          if (
                            currentMatch.status === "open" &&
                            hasBothParticipants
                          ) {
                            if (selectedMatchId !== currentMatch.matchId) {
                              onStartMatch(match);
                            } else {
                              clearMatchSelection();
                            }
                          }
                        }}
                      >
                        {isSelected ? "Unload" : "Load match"}
                      </Button>
                    {/if}

                    {#if showActions}
                      <ButtonGroup.Root
                        orientation="horizontal"
                        class="flex justify-end"
                        onclick={(event) => event.stopPropagation()}
                      >
                        <AlertDialog bind:open={omtDialogOpen}>
                          <AlertDialogTrigger>
                            <Button
                              variant="outline"
                              class={`rounded-r-none ${
                                isTied ? "border-amber-400 text-amber-400" : ""
                              }`}
                              disabled={isProcessing}>OMT</Button
                            >
                          </AlertDialogTrigger>

                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Start OMT?</AlertDialogTitle>
                              <AlertDialogDescription>
                                This will reset the battle and start an OMT. Are
                                you sure you want to continue?
                              </AlertDialogDescription>
                            </AlertDialogHeader>

                            <AlertDialogFooter>
                              <AlertDialogCancel disabled={isProcessing}>
                                Cancel
                              </AlertDialogCancel>

                              <AlertDialogAction
                                disabled={isProcessing}
                                onclick={() =>
                                  handleOMT(match, roundKey, matchNumber)}
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
                              <AlertDialogTitle
                                >Conclude match?</AlertDialogTitle
                              >
                              <AlertDialogDescription>
                                This will finalize the result of this match. Are
                                you sure?
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
              {:else if !hasBothParticipants}
                <div class="border-t px-4 py-3 text-xs text-muted-foreground">
                  Waiting for participants
                </div>
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
