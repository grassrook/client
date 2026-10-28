<script lang="ts">
  import {
    ArrowLeft,
    CalendarDays,
    Clock3,
    Save,
    Trash2,
    UserRound,
  } from "@lucide/svelte";
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
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from "$src/components/ui/dialog";
  import { Input } from "$src/components/ui/input";
  import { Label } from "$src/components/ui/label";
  import Separator from "$src/components/ui/separator/separator.svelte";
  import { Textarea } from "$src/components/ui/textarea";
  import {
    DeleteDriver,
    GetDriverByID,
    UpdateDriver,
  } from "$wails/go/main/App";
  import type { driver } from "$wails/go/models";

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

  function confirmDelete() {
    deleteError = "";
    deleteDriverMutation.mutate(params.id);
  }

  function formatDate(timestamp: number) {
    if (!timestamp) return "Not available";
    return new Date(timestamp * 1000).toLocaleDateString(undefined, {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  }
</script>

{#if mounted}
  <div in:fade={{ duration: 300 }}>
    <div class="min-h-full">
      <div
        class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
      >
        <div class="flex items-center gap-3">
          <Button
            variant="ghost"
            size="icon"
            onclick={() => push("/drivers")}
            aria-label="Back to drivers"
          >
            <ArrowLeft class="size-5" />
          </Button>

          <div>
            <p class="text-sm text-muted-foreground">Driver directory</p>
            <h1 class="text-2xl font-semibold tracking-tight">
              Driver profile
            </h1>
          </div>
        </div>
      </div>

      <Separator />

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
          {@const driver = driverQuery.data as driver.Driver}

          <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_280px]">
            <div class="min-w-0 space-y-6">
              <section class="overflow-hidden rounded-xl border">
                <div
                  class="h-24 bg-gradient-to-r from-muted via-muted/60 to-background"
                ></div>

                <div class="px-6 pb-6">
                  <div
                    class="-mt-10 flex flex-wrap items-end justify-between gap-4"
                  >
                    <div class="flex items-end gap-4">
                      <div
                        class="flex size-20 items-center justify-center rounded-2xl border-4 border-card bg-muted shadow-sm"
                      >
                        <UserRound class="size-9 text-muted-foreground" />
                      </div>

                      <div class="min-w-0 pb-1">
                        <h2 class="text-xl font-semibold tracking-tight">
                          {driver.name}
                        </h2>
                        <p class="mt-1 text-sm text-muted-foreground">
                          Driver profile
                        </p>
                      </div>
                    </div>
                  </div>

                  <Separator class="my-5" />

                  <div class="grid gap-4 sm:grid-cols-2">
                    <div class="flex items-start gap-3">
                      <div class="rounded-lg border bg-muted/40 p-2">
                        <CalendarDays class="size-4 text-muted-foreground" />
                      </div>
                      <div>
                        <p class="text-xs text-muted-foreground">Date added</p>
                        <p class="mt-1 text-sm font-medium">
                          {formatDate(driver.createdAt)}
                        </p>
                      </div>
                    </div>

                    <div class="flex items-start gap-3">
                      <div class="rounded-lg border bg-muted/40 p-2">
                        <Clock3 class="size-4 text-muted-foreground" />
                      </div>
                      <div>
                        <p class="text-xs text-muted-foreground">
                          Last updated
                        </p>
                        <p class="mt-1 text-sm font-medium">
                          {formatDate(driver.updatedAt)}
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </section>

              <section class="rounded-xl border">
                <div class="border-b px-6 py-4">
                  <h2 class="font-semibold">Personal information</h2>
                  <p class="mt-1 text-sm text-muted-foreground">
                    Update this driver's profile information.
                  </p>
                </div>

                <form class="space-y-5 p-6" onsubmit={submitDriver}>
                  <div class="space-y-2">
                    <Label for="driver-name">Driver name</Label>
                    <Input
                      id="driver-name"
                      bind:value={name}
                      placeholder="Enter driver name"
                      maxlength={255}
                      required
                    />
                  </div>

                  <div class="space-y-2">
                    <Label for="driver-description">Description</Label>
                    <Textarea
                      id="driver-description"
                      bind:value={description}
                      placeholder="Add information about this driver..."
                      rows={5}
                    />
                    <p class="text-xs text-muted-foreground">
                      Optional. Use this field for notes or other relevant
                      details.
                    </p>
                  </div>

                  {#if errorMessage}
                    <p class="text-sm text-destructive">{errorMessage}</p>
                  {/if}

                  {#if successMessage}
                    <p class="text-sm text-green-600 dark:text-green-400">
                      {successMessage}
                    </p>
                  {/if}

                  <div class="flex justify-end border-t pt-5">
                    <Button
                      type="submit"
                      disabled={updateDriverMutation.isPending}
                    >
                      <Save class="mr-2 size-4" />
                      {updateDriverMutation.isPending
                        ? "Saving changes..."
                        : "Save changes"}
                    </Button>
                  </div>
                </form>
              </section>
            </div>

            <aside class="space-y-6">
              <section class="rounded-xl border p-5">
                <h2 class="font-semibold">Driver information</h2>
                <p class="mt-1 text-sm text-muted-foreground">
                  System details for this record.
                </p>

                <Separator class="my-4" />

                <div class="space-y-2">
                  <Label for="driver-id">Driver ID</Label>
                  <Input id="driver-id" value={driver.id} readonly />
                  <p class="text-xs text-muted-foreground">
                    Unique identifier assigned to this driver.
                  </p>
                </div>
              </section>

              <section class="rounded-xl border border-destructive/30 p-5">
                <div class="flex items-center gap-2">
                  <Trash2 class="size-4 text-destructive" />
                  <h2 class="font-semibold">Danger zone</h2>
                </div>

                <p class="mt-2 text-sm text-muted-foreground">
                  Deleting this driver also deletes their registered cars and
                  associated car photos.
                </p>

                <Button
                  class="mt-4 w-full"
                  variant="destructive"
                  onclick={openDeleteDialog}
                >
                  Delete driver
                </Button>
              </section>
            </aside>
          </div>
        {/if}
      </div>
    </div>

    <Dialog
      bind:open={deleteDialogOpen}
      onOpenChange={(open) => {
        if (!open && !deleteDriverMutation.isPending) {
          deleteDialogOpen = false;
          deleteError = "";
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete driver?</DialogTitle>
          <DialogDescription>
            You're about to delete
            <strong class="text-foreground">{driverQuery.data?.name}</strong>.
            Their registered cars and associated photos will also be deleted.
            This action cannot be undone.
          </DialogDescription>
        </DialogHeader>

        {#if deleteError}
          <p class="text-sm text-destructive">{deleteError}</p>
        {/if}

        <DialogFooter>
          <Button
            variant="outline"
            onclick={() => (deleteDialogOpen = false)}
            disabled={deleteDriverMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onclick={confirmDelete}
            disabled={deleteDriverMutation.isPending}
          >
            <Trash2 class="mr-2 size-4" />
            {deleteDriverMutation.isPending
              ? "Deleting..."
              : "Delete permanently"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
{/if}
