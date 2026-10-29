<script lang="ts">
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import { push } from "svelte-spa-router";

  import { Button } from "$src/components/ui/button";
  import {
    DeleteDriver,
    GetDriverByID,
    UpdateDriver,
  } from "$wails/go/main/App";
  import type { driver } from "$wails/go/models";

  import DeleteDriverDialog from "../driver-profile/delete-driver-dialog.svelte";
  import DriverProfileCard from "../driver-profile/driver-profile-card.svelte";
  import DriverProfileForm from "../driver-profile/driver-profile-form.svelte";
  import DriverProfileHeader from "../driver-profile/driver-profile-header.svelte";
  import DriverProfileSidebar from "../driver-profile/driver-profile-sidebar.svelte";

  type Props = {
    params: {
      id: string;
    };
  };

  let { params }: Props = $props();

  const queryClient = useQueryClient();

  const driverQuery = createQuery(() => ({
    queryKey: ["driver", params.id],
    queryFn: () => GetDriverByID(params.id),
    enabled: !!params.id,
  }));

  let mounted = $state(false);
  let name = $state("");
  let description = $state("");
  let initializedDriverID = $state("");
  let errorMessage = $state("");
  let successMessage = $state("");
  let deleteError = $state("");
  let deleteDialogOpen = $state(false);

  $effect(() => {
    const driver = driverQuery.data as driver.Driver | undefined;

    if (driver && initializedDriverID !== driver.id) {
      name = driver.name;
      description = driver.description;
      initializedDriverID = driver.id;
      errorMessage = "";
      successMessage = "";
    }
  });

  onMount(() => {
    mounted = true;
  });

  const updateDriverMutation = createMutation(() => ({
    mutationFn: (input: { id: string; name: string; description: string }) =>
      UpdateDriver(input),

    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({
          queryKey: ["driver", params.id],
        }),
        queryClient.invalidateQueries({
          queryKey: ["drivers"],
        }),
      ]);

      successMessage = "Driver details saved successfully.";
      errorMessage = "";
    },

    onError: (error) => {
      errorMessage = String(error);
      successMessage = "";
    },
  }));

  const deleteDriverMutation = createMutation(() => ({
    mutationFn: (id: string) => DeleteDriver(id),

    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["drivers"] }),
        queryClient.invalidateQueries({ queryKey: ["registered-cars"] }),
      ]);

      await queryClient.removeQueries({
        queryKey: ["driver", params.id],
      });

      push("/drivers");
    },

    onError: (error) => {
      deleteError = String(error);
    },
  }));

  function submitDriver(event: SubmitEvent) {
    event.preventDefault();

    const trimmedName = name.trim();

    errorMessage = "";
    successMessage = "";

    if (!trimmedName) {
      errorMessage = "Driver name is required.";
      return;
    }

    updateDriverMutation.mutate({
      id: params.id,
      name: trimmedName,
      description: description.trim(),
    });
  }

  function openDeleteDialog() {
    deleteError = "";
    deleteDialogOpen = true;
  }

  function handleDeleteDialogOpenChange(open: boolean) {
    if (!open && !deleteDriverMutation.isPending) {
      deleteDialogOpen = false;
      deleteError = "";
    }
  }

  function confirmDelete() {
    deleteError = "";
    deleteDriverMutation.mutate(params.id);
  }
</script>

{#if mounted}
  <div in:fade={{ duration: 300 }}>
    <div class="min-h-full">
      <DriverProfileHeader />

      <div class="mx-auto max-w-6xl space-y-6 px-4 py-6 sm:px-6 lg:px-8">
        {#if driverQuery.isPending}
          <div class="space-y-4 rounded-xl border p-6">
            <div class="h-16 w-16 animate-pulse rounded-full bg-muted"></div>
            <div class="h-5 w-48 animate-pulse rounded bg-muted"></div>
            <div
              class="h-4 w-72 max-w-full animate-pulse rounded bg-muted"
            ></div>
            <div class="h-32 animate-pulse rounded-lg bg-muted"></div>
          </div>
        {:else if driverQuery.isError}
          <div class="rounded-xl border px-6 py-12 text-center">
            <h2 class="font-semibold">Couldn't load this driver</h2>
            <p class="mt-2 text-sm text-muted-foreground">
              The driver may no longer exist, or something went wrong.
            </p>

            <div class="mt-5 flex justify-center gap-2">
              <Button variant="outline" onclick={() => driverQuery.refetch()}>
                Try again
              </Button>
              <Button onclick={() => push("/drivers")}>Back to drivers</Button>
            </div>
          </div>
        {:else if driverQuery.data}
          {@const currentDriver = driverQuery.data as driver.Driver}

          <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_280px]">
            <div class="min-w-0 space-y-6">
              <DriverProfileCard driver={currentDriver} />

              <DriverProfileForm
                bind:name
                bind:description
                {errorMessage}
                {successMessage}
                saving={updateDriverMutation.isPending}
                onSubmit={submitDriver}
              />
            </div>

            <DriverProfileSidebar
              driverId={currentDriver.id}
              onDelete={openDeleteDialog}
            />
          </div>
        {/if}
      </div>
    </div>

    <DeleteDriverDialog
      open={deleteDialogOpen}
      driverName={driverQuery.data?.name}
      error={deleteError}
      deleting={deleteDriverMutation.isPending}
      onOpenChange={handleDeleteDialogOpenChange}
      onConfirm={confirmDelete}
    />
  </div>
{/if}
