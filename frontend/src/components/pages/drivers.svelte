<script lang="ts">
  import { Plus, UserRound, UsersRound } from "@lucide/svelte";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import { link } from "svelte-spa-router";

  import * as Avatar from "$src/components/ui/avatar";
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
  import * as Table from "$src/components/ui/table";
  import { Textarea } from "$src/components/ui/textarea";
  import { CreateDriver, GetDrivers } from "$wails/go/main/App";

  const queryClient = useQueryClient();

  const driversQuery = createQuery(() => ({
    queryKey: ["drivers"],
    queryFn: () => GetDrivers(),
  }));

  let dialogOpen = $state(false);
  let name = $state("");
  let description = $state("");
  let errorMessage = $state("");

  const createDriverMutation = createMutation(() => ({
    mutationFn: (input: { name: string; description: string }) =>
      CreateDriver(input),

    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["drivers"] });
      closeDialog();
    },

    onError: (error) => {
      errorMessage = String(error);
    },
  }));

  function openCreateDialog() {
    name = "";
    description = "";
    errorMessage = "";
    dialogOpen = true;
  }

  function closeDialog() {
    dialogOpen = false;
    name = "";
    description = "";
    errorMessage = "";
  }

  function submitDriver(event: SubmitEvent) {
    event.preventDefault();

    const trimmedName = name.trim();

    if (!trimmedName) {
      errorMessage = "Driver name is required.";
      return;
    }

    errorMessage = "";

    createDriverMutation.mutate({
      name: trimmedName,
      description: description.trim(),
    });
  }

  function formatDate(timestamp: number) {
    return new Date(timestamp * 1000).toLocaleDateString();
  }
  let mounted = $state(false);

  onMount(() => {
    mounted = true;
  });
</script>

{#if mounted}
  <div in:fade={{ duration: 300 }}>
    <div class="min-h-full">
      <div
        class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
      >
        <div class="flex items-start gap-3">
          <div
            class="flex size-11 items-center justify-center rounded-xl border bg-muted/50"
          >
            <UsersRound class="size-5 text-muted-foreground" />
          </div>

          <div>
            <h1 class="text-2xl font-semibold tracking-tight">Drivers</h1>
            <p class="mt-1 text-sm text-muted-foreground">
              Manage your drivers and their profiles.
            </p>
          </div>
        </div>

        <Button
          class="bg-blue-600 text-white hover:bg-blue-700"
          onclick={openCreateDialog}
        >
          <Plus class="mr-2 size-4" />
          New driver
        </Button>
      </div>

      <Separator />

      <div class="mx-auto max-w-6xl px-4 py-6 sm:px-6 lg:px-8">
        <div class="overflow-hidden rounded-xl border">
          <div class="flex items-center justify-between border-b px-5 py-4">
            <div>
              <h2 class="font-medium">All drivers</h2>
              <p class="mt-1 text-sm text-muted-foreground">
                {#if !driversQuery.isPending && !driversQuery.isError}
                  {(driversQuery.data ?? []).length}
                  {(driversQuery.data ?? []).length === 1
                    ? " driver"
                    : " drivers"} registered
                {:else}
                  Driver directory
                {/if}
              </p>
            </div>
          </div>

          {#if driversQuery.isPending}
            <div class="space-y-4 p-5">
              {#each [1, 2, 3] as item (item)}
                <div class="flex items-center gap-3">
                  <div
                    class="size-10 animate-pulse rounded-full bg-muted"
                  ></div>
                  <div class="flex-1 space-y-2">
                    <div class="h-4 w-36 animate-pulse rounded bg-muted"></div>
                    <div
                      class="h-3 w-56 max-w-full animate-pulse rounded bg-muted"
                    ></div>
                  </div>
                </div>
              {/each}
            </div>
          {:else if driversQuery.isError}
            <div
              class="flex flex-col items-center gap-3 px-6 py-12 text-center"
            >
              <p class="font-medium">Couldn't load drivers</p>
              <p class="text-sm text-muted-foreground">
                Something went wrong while fetching the driver directory.
              </p>
              <Button variant="outline" onclick={() => driversQuery.refetch()}>
                Try again
              </Button>
            </div>
          {:else if !(driversQuery.data ?? []).length}
            <div class="flex flex-col items-center px-6 py-14 text-center">
              <div
                class="mb-4 flex size-14 items-center justify-center rounded-full bg-muted"
              >
                <UsersRound class="size-6 text-muted-foreground" />
              </div>
              <h3 class="font-semibold">No drivers yet</h3>
              <p class="mt-1 max-w-sm text-sm text-muted-foreground">
                Get started by adding your first driver to the directory.
              </p>
              <Button class="mt-5" onclick={openCreateDialog}>
                <Plus class="mr-2 size-4" />
                Create your first driver
              </Button>
            </div>
          {:else}
            <div class="overflow-x-auto">
              <Table.Root>
                <Table.Header>
                  <Table.Row class="bg-muted/30 hover:bg-muted/30">
                    <Table.Head class="min-w-64">Driver</Table.Head>
                    <Table.Head class="min-w-64">Description</Table.Head>
                    <Table.Head class="min-w-32">Date added</Table.Head>
                  </Table.Row>
                </Table.Header>

                <Table.Body>
                  {#each driversQuery.data ?? [] as driver (driver.id)}
                    <Table.Row class="cursor-pointer transition-colors">
                      <Table.Cell>
                        <div class="flex items-center gap-3">
                          <Avatar.Root class="size-10 border">
                            <Avatar.Fallback
                              class="bg-muted text-muted-foreground"
                            >
                              <UserRound class="size-5" />
                            </Avatar.Fallback>
                          </Avatar.Root>

                          <div class="min-w-0">
                            <a
                              href={`/drivers/${driver.id}`}
                              use:link
                              class="truncate font-medium hover:underline"
                            >
                              {driver.name}
                            </a>
                            <p class="truncate text-xs text-muted-foreground">
                              ID: {driver.id}
                            </p>
                          </div>
                        </div>
                      </Table.Cell>

                      <Table.Cell>
                        <p
                          class="max-w-md truncate text-sm text-muted-foreground"
                        >
                          {driver.description || "No description provided"}
                        </p>
                      </Table.Cell>

                      <Table.Cell class="text-sm text-muted-foreground">
                        {formatDate(driver.createdAt)}
                      </Table.Cell>
                    </Table.Row>
                  {/each}
                </Table.Body>
              </Table.Root>
            </div>
          {/if}
        </div>
      </div>
    </div>

    <Dialog
      bind:open={dialogOpen}
      onOpenChange={(open) => {
        if (!open && !createDriverMutation.isPending) {
          closeDialog();
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create driver</DialogTitle>
          <DialogDescription>
            Add a driver to your directory. You can register their cars
            separately.
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-4" onsubmit={submitDriver}>
          <div class="space-y-2">
            <Label for="driver-name">Driver name</Label>
            <Input
              id="driver-name"
              bind:value={name}
              placeholder="e.g. Alex Tanaka"
              maxlength={255}
              required
            />
          </div>

          <div class="space-y-2">
            <Label for="driver-description">Description</Label>
            <Textarea
              id="driver-description"
              bind:value={description}
              placeholder="Add any relevant information about this driver..."
              rows={3}
            />
            <p class="text-xs text-muted-foreground">
              Optional. You can leave this blank.
            </p>
          </div>

          {#if errorMessage}
            <p class="text-sm text-destructive">{errorMessage}</p>
          {/if}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onclick={closeDialog}
              disabled={createDriverMutation.isPending}
            >
              Cancel
            </Button>

            <Button type="submit" disabled={createDriverMutation.isPending}>
              {#if createDriverMutation.isPending}
                Creating...
              {:else}
                <Plus class="mr-2 size-4" />
                Create driver
              {/if}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
{/if}
