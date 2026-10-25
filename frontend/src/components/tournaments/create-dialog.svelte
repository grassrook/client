<script lang="ts">
  import { type CalendarDate, getLocalTimeZone } from "@internationalized/date";
  import { Plus } from "@lucide/svelte";
  import CalendarIcon from "@lucide/svelte/icons/calendar";
  import Clock2Icon from "@lucide/svelte/icons/clock-2";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";

  import { Button } from "$src/components/ui/button";
  import { Calendar } from "$src/components/ui/calendar";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "$src/components/ui/dialog";
  import * as Field from "$src/components/ui/field";
  import { Input } from "$src/components/ui/input";
  import * as InputGroup from "$src/components/ui/input-group";
  import { Label } from "$src/components/ui/label";
  import * as Popover from "$src/components/ui/popover";
  import * as Select from "$src/components/ui/select";
  import Switch from "$src/components/ui/switch/switch.svelte";
  import { CreateTournament, GetCriteriaConfigs } from "$wails/go/main/App";

  let { open = $bindable(false) } = $props();
  let error = $state("");
  let name = $state("");
  let size = $state(2);
  let trackName = $state("");

  let randomizeQualifyingOrder = $state(true);
  let criteriaConfigID = $state<number | undefined>(undefined);
  let datePopoverOpen = $state(false);
  let qualifyingRounds = $state<"1" | "2">("1"); // qualifying rounds in most drifting events is mostly up to 2 rounds

  let selectedDate = $state<CalendarDate | undefined>();
  let startingTime = $state("12:00");

  const queryClient = useQueryClient();

  const criteriaConfigsQuery = createQuery(() => ({
    queryKey: ["criteria-configs"],
    queryFn: () => GetCriteriaConfigs(),
  }));

  const createTournamentMutation = createMutation(() => ({
    mutationFn: async ({
      name,
      size,
      startingDate,
      track,
      randomizeQualifyingOrder,
      criteriaConfigID,
    }: {
      name: string;
      size: number;
      startingDate: string;
      track: string;
      randomizeQualifyingOrder: boolean;
      criteriaConfigID: number;
    }) => {
      await CreateTournament({
        name,
        size,
        startingDate,
        track,
        randomizeQualifyingOrder,
        criteriaConfigID,
        qualifyingRounds: Number(qualifyingRounds),
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["tournaments"],
      });
      open = false;
      name = "";
      selectedDate = undefined;
      startingTime = "12:00";
      trackName = "";
      error = "";
      criteriaConfigID = undefined;
      qualifyingRounds = "1";
    },
  }));

  async function createTournament() {
    error = "";
    if (!name.trim()) {
      error = "Tournament name is required.";
      return;
    }

    if (!selectedDate) {
      error = "Starting date is required.";
      return;
    }

    if (!startingTime) {
      error = "Starting time is required.";
      return;
    }

    const startingDate = `${selectedDate.toString()}T${startingTime}`;

    if (!trackName.trim()) {
      error = "Track is required.";
      return;
    }

    if (!criteriaConfigID) {
      error = "Criteria config is required";
      return;
    }

    await createTournamentMutation.mutateAsync({
      name: name.trim(),
      size,
      startingDate,
      track: trackName.trim(),
      randomizeQualifyingOrder,
      criteriaConfigID,
    });
  }
</script>

<Dialog bind:open>
  <DialogTrigger>
    {#snippet child({ props })}
      <Button {...props}>
        <Plus />
        New tournament
      </Button>
    {/snippet}
  </DialogTrigger>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>Create Tournament</DialogTitle>
      <DialogDescription>
        Create a tournament and add its qualifying participants.
      </DialogDescription>
    </DialogHeader>
    <div class="space-y-4">
      <div class="space-y-2">
        <Label for="name">Tournament name</Label>
        <Input
          id="name"
          placeholder="Summer Drift Championship"
          bind:value={name}
        />
      </div>
      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="size">Size</Label>
          <Input
            type="number"
            min="2"
            id="size"
            placeholder="Size of the bracket"
            bind:value={size}
          />
        </div>
        <div class="space-y-2">
          <Label>Starting date</Label>
          <Popover.Root bind:open={datePopoverOpen}>
            <Popover.Trigger>
              {#snippet child({ props })}
                <Button
                  {...props}
                  variant="outline"
                  class="w-full justify-start font-normal"
                >
                  <CalendarIcon class="mr-2 size-4" />
                  {#if selectedDate}
                    {selectedDate
                      .toDate(getLocalTimeZone())
                      .toLocaleDateString()}
                    at {startingTime}
                  {:else}
                    Select date time
                  {/if}
                </Button>
              {/snippet}
            </Popover.Trigger>

            <Popover.Content class="w-auto p-0" align="start">
              <div class="p-3">
                <Calendar
                  type="single"
                  bind:value={selectedDate}
                  captionLayout="dropdown"
                />
              </div>

              <div class="border-t bg-card p-3">
                <Field.Group>
                  <Field.Field>
                    <Field.Label for="starting-time">Start time</Field.Label>
                    <InputGroup.Root>
                      <InputGroup.Input
                        id="starting-time"
                        type="time"
                        step="60"
                        bind:value={startingTime}
                        class="appearance-none [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
                      />
                      <InputGroup.Addon>
                        <Clock2Icon class="text-muted-foreground" />
                      </InputGroup.Addon>
                    </InputGroup.Root>
                  </Field.Field>
                </Field.Group>

                <Button
                  class="mt-3 w-full"
                  onclick={() => (datePopoverOpen = false)}
                  disabled={!selectedDate || !startingTime}
                >
                  Done
                </Button>
              </div>
            </Popover.Content>
          </Popover.Root>
        </div>
      </div>
      <div class="space-y-2">
        <Label for="track">Track</Label>
        <Input
          id="track"
          placeholder="Enter or select a track"
          bind:value={trackName}
        />
      </div>
      <div class="space-y-2 w-full">
        <Label for="qualifyingRounds">Qualifying Rounds</Label>
        <Select.Root type="single" bind:value={qualifyingRounds}>
          <Select.Trigger class="w-full">
            <Select.Value placeholder="Select qualifying rounds" />
          </Select.Trigger>
          <Select.Content>
            <Select.Item value="1">1 Round</Select.Item>
            <Select.Item value="2">2 Rounds</Select.Item>
          </Select.Content>
        </Select.Root>
      </div>
      <div class="space-y-2">
        <Label for="criteria-config">Criteria configuration</Label>

        <Select.Root
          type="single"
          value={criteriaConfigID?.toString() ?? ""}
          onValueChange={(value) => {
            criteriaConfigID = value ? Number(value) : undefined;
          }}
        >
          <Select.Trigger id="criteria-config" class="w-full">
            {#if criteriaConfigsQuery.isPending}
              Loading criteria configurations...
            {:else}
              {criteriaConfigsQuery.data?.find(
                (config) => config.id === criteriaConfigID,
              )?.name ?? "Select a criteria configuration"}
            {/if}
          </Select.Trigger>

          <Select.Content>
            {#each criteriaConfigsQuery.data ?? [] as config (config.id)}
              <Select.Item value={config.id.toString()}>
                {config.name}
              </Select.Item>
            {/each}
          </Select.Content>
        </Select.Root>
        {#if criteriaConfigsQuery.isError}
          <p class="text-sm text-destructive">
            Failed to load criteria configurations.
          </p>
        {:else if !criteriaConfigsQuery.isPending && !criteriaConfigsQuery.data?.length}
          <p class="text-sm text-muted-foreground">
            No criteria configurations found. Create one before creating a
            tournament.
          </p>
        {/if}
      </div>

      <div class="flex items-center justify-between gap-4 py-2.5">
        <Label>Randomize Qualifying Order</Label>
        <Switch bind:checked={randomizeQualifyingOrder} />
      </div>
      {#if error}
        <p class="text-sm text-destructive">
          {error}
        </p>
      {/if}
      {#if createTournamentMutation.isError}
        <p class="text-sm text-destructive">
          {createTournamentMutation.error.message}
        </p>
      {/if}
    </div>
    <DialogFooter>
      <Button
        disabled={createTournamentMutation.isPending ||
          criteriaConfigsQuery.isPending ||
          !criteriaConfigsQuery.data?.length}
        onclick={createTournament}
      >
        {createTournamentMutation.isPending
          ? "Creating..."
          : "Create tournament"}
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
