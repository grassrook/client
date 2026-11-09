<script lang="ts">
  import ArrowLeftRightIcon from "@lucide/svelte/icons/arrow-left-right";
  import GripVerticalIcon from "@lucide/svelte/icons/grip-vertical";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";
  import UserIcon from "@lucide/svelte/icons/user";

  import { Button } from "$src/components/ui/button/index.js";
  import * as Table from "$src/components/ui/table/index.js";
  import type { entrant, tournament } from "$wails/go/models";

  import ParticipantCarSelect from "./participant-car-select.svelte";

  type Props = {
    tournament: tournament.Tournament;
    items: entrant.Entrant[];
    editable: boolean;
    canReorder: boolean;
    canReplace: boolean;
    bracketParticipantIds: Set<string>;
    onReorder: (next: entrant.Entrant[]) => void;
    onReplace: (entrant: entrant.Entrant) => void;
    onDelete: (entrant: entrant.Entrant) => void;
  };

  let {
    tournament,
    items,
    editable,
    canReorder,
    canReplace,
    bracketParticipantIds,
    onReorder,
    onReplace,
    onDelete,
  }: Props = $props();

  const showActions = $derived(editable || canReplace);
  const isElimination = $derived(tournament.status === "elimination");

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
    onReorder(next);
  }

  function resetDrag() {
    dragIndex = null;
    overIndex = null;
  }
</script>

<div class="rounded-lg border">
  <Table.Root>
    <Table.Header>
      <Table.Row>
        {#if canReorder}
          <Table.Head class="w-10"></Table.Head>
        {/if}
        <Table.Head class="w-16">#</Table.Head>
        <Table.Head>Name</Table.Head>
        {#if isElimination}
          <Table.Head class="w-24 text-left">Score</Table.Head>
        {/if}

        <Table.Head class={!editable ? "text-left" : ""}>
          Registered car
        </Table.Head>
        {#if showActions}
          <Table.Head class="w-24 text-right">Actions</Table.Head>
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

          <Table.Cell class="text-muted-foreground">{index + 1}</Table.Cell>
          <Table.Cell class="font-medium">
            <div class="flex items-center gap-3">
              <!-- placeholder profile picture -->
              <div
                class="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground"
              >
                <UserIcon class="size-4" />
              </div>
              <span>{entrant.name}</span>
            </div>
          </Table.Cell>

          {#if isElimination}
            <Table.Cell class="text-left tabular-nums font-medium">
              {entrant.score ?? 0}
            </Table.Cell>
          {/if}

          <Table.Cell>
            {#if entrant.driverId}
              <ParticipantCarSelect
                tournamentId={tournament.id}
                participantId={entrant.participantId}
                driverId={entrant.driverId}
                registeredCarId={entrant.registeredCarID}
                disabled={!editable}
              />
            {:else}
              <span class="text-sm text-muted-foreground">
                No registered driver
              </span>
            {/if}
          </Table.Cell>

          {#if showActions}
            <Table.Cell class="text-right">
              {#if canReplace && bracketParticipantIds.has(entrant.participantId)}
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Replace ${entrant.name}`}
                  onclick={() => onReplace(entrant)}
                >
                  <ArrowLeftRightIcon class="size-4" />
                </Button>
              {/if}
              {#if editable}
                <Button
                  variant="ghost"
                  size="icon"
                  aria-label={`Delete ${entrant.name}`}
                  onclick={() => onDelete(entrant)}
                >
                  <Trash2Icon class="size-4 text-destructive" />
                </Button>
              {/if}
            </Table.Cell>
          {/if}
        </Table.Row>
      {/each}
    </Table.Body>
  </Table.Root>
</div>
