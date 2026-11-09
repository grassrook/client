<script lang="ts">
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";

  import * as Select from "$src/components/ui/select/index.js";
  import { GetDriverCarList, UpdateParticipantCar } from "$wails/go/main/App";

  type Props = {
    tournamentId: string;
    participantId: string;
    driverId: string;
    registeredCarId?: string;
    disabled?: boolean;
  };

  let {
    tournamentId,
    participantId,
    driverId,
    registeredCarId,
    disabled = false,
  }: Props = $props();

  const queryClient = useQueryClient();

  // cached per driver, so several rows never fetch the same list twice
  const carsQuery = createQuery(() => ({
    queryKey: ["driverCars", driverId],
    queryFn: () => GetDriverCarList(driverId),
  }));

  const cars = $derived(carsQuery.data ?? []);

  const selectedCarName = $derived(
    cars.find((car) => car.id === registeredCarId)?.name ?? "Unknown car",
  );

  const updateCarMutation = createMutation(() => ({
    mutationFn: async (carId: string) =>
      UpdateParticipantCar(participantId, carId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["entrants", tournamentId],
      });
    },
  }));
</script>

{#if carsQuery.isPending}
  <span class="text-sm text-muted-foreground">Loading cars...</span>
{:else if carsQuery.isError}
  <span class="text-sm text-destructive">
    {carsQuery.error instanceof Error
      ? carsQuery.error.message
      : "Failed to load cars."}
  </span>
{:else}
  <Select.Root
    type="single"
    value={registeredCarId || "__none__"}
    onValueChange={(value) =>
      updateCarMutation.mutate(value === "__none__" ? "" : value)}
    disabled={updateCarMutation.isPending || disabled}
  >
    <Select.Trigger
      class="bg-transparent! hover:bg-transparent! data-[state=open]:bg-transparent! h-9 w-full min-w-40"
    >
      <Select.Value>
        {#if registeredCarId}
          {selectedCarName}
        {:else}
          No car assigned
        {/if}
      </Select.Value>
    </Select.Trigger>

    <Select.Content>
      <Select.Item value="__none__">No car assigned</Select.Item>

      {#each cars as car (car.id)}
        <Select.Item value={car.id}>{car.name}</Select.Item>
      {/each}
    </Select.Content>
  </Select.Root>
{/if}
