<script lang="ts">
  import { Plus, Search, UserRound, UsersRound, X } from "@lucide/svelte";
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
  let searchQuery = $state("");
  let mounted = $state(false);

  const filteredDrivers = $derived.by(() => {
    const drivers = driversQuery.data ?? [];
    const query = searchQuery.trim().toLowerCase();

    if (!query) return drivers;

    return drivers.filter((driver) =>
      [driver.name, driver.description, driver.id].some((value) =>
        value.toLowerCase().includes(query),
      ),
    );
  });

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

  onMount(() => {
    mounted = true;
  });
</script>

{#if mounted}
  <div in:fade={{ duration: 300 }}>
    <div class="min-h-full">
      <!-- Page heading -->
      <header
        class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-8 sm:px-6 lg:px-8"
      >
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">Drivers</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Manage your drivers and their registered information.
          </p>
        </div>

        <Button
          class="bg-blue-600 text-white hover:bg-blue-700"
          onclick={openCreateDialog}
        >
          <Plus class="mr-2 size-4" />
          Add driver
        </Button>
      </header>

      <Separator />

      <main class="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <!-- Directory heading and search -->
        <section>
          <div
            class="flex flex-col justify-between gap-5 sm:flex-row sm:items-end"
          >
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-base font-semibold">Driver directory</h2>

                {#if !driversQuery.isPending && !driversQuery.isError}
                  <span
                    class="rounded-md border px-2 py-0.5 text-xs font-medium text-muted-foreground"
                  >
                    {(driversQuery.data ?? []).length}
                  </span>
                {/if}
              </div>

              <p class="mt-1 text-sm text-muted-foreground">
                Browse and search through your registered drivers.
              </p>
            </div>

            <div class="relative w-full sm:max-w-xs">
              <Search
                class="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
              />

              <Input
                bind:value={searchQuery}
                placeholder="Search by name or ID..."
                class="h-10 pl-9 pr-9"
              />

              {#if searchQuery}
                <button
                  type="button"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                  onclick={() => (searchQuery = "")}
                  aria-label="Clear search"
                >
                  <X class="size-4" />
                </button>
              {/if}
            </div>
          </div>

          <!-- Driver table -->
          <div class="mt-6">
            {#if driversQuery.isPending}
              <div class="space-y-5 py-5">
                {#each [1, 2, 3] as item (item)}
                  <div class="flex items-center gap-4">
                    <div
                      class="size-10 animate-pulse rounded-full bg-muted"
                    ></div>
                    <div class="flex-1 space-y-2">
                      <div
                        class="h-4 w-36 animate-pulse rounded bg-muted"
                      ></div>
                      <div
                        class="h-3 w-56 max-w-full animate-pulse rounded bg-muted"
                      ></div>
                    </div>
                  </div>
                {/each}
              </div>
            {:else if driversQuery.isError}
              <div class="flex flex-col items-center gap-3 py-16 text-center">
                <p class="font-medium">Couldn't load drivers</p>
                <p class="text-sm text-muted-foreground">
                  Something went wrong while fetching the driver directory.
                </p>
                <Button
                  variant="outline"
                  onclick={() => driversQuery.refetch()}
                >
                  Try again
                </Button>
              </div>
            {:else if !(driversQuery.data ?? []).length}
              <div
                class="flex flex-col items-center border-y py-16 text-center"
              >
                <UsersRound class="mb-4 size-9 text-muted-foreground" />
                <h3 class="font-medium">No drivers yet</h3>
                <p class="mt-1 max-w-sm text-sm text-muted-foreground">
                  Add your first driver to start building your directory.
                </p>
                <Button class="mt-5" onclick={openCreateDialog}>
                  <Plus class="mr-2 size-4" />
                  Add your first driver
                </Button>
              </div>
            {:else if !filteredDrivers.length}
              <div
                class="flex flex-col items-center border-y py-16 text-center"
              >
                <Search class="mb-4 size-8 text-muted-foreground" />
                <h3 class="font-medium">No matching drivers</h3>
                <p class="mt-1 text-sm text-muted-foreground">
                  No results found for "{searchQuery}".
                </p>
                <Button
                  class="mt-4"
                  variant="outline"
                  size="sm"
                  onclick={() => (searchQuery = "")}
                >
                  Clear search
                </Button>
              </div>
            {:else}
              <div class="overflow-x-auto border-y">
                <Table.Root>
                  <Table.Header>
                    <Table.Row class="hover:bg-transparent">
                      <Table.Head class="h-11 min-w-64 pl-4">Driver</Table.Head>
                      <Table.Head class="h-11 min-w-64">Description</Table.Head>
                      <Table.Head class="h-11 min-w-32 pr-4 text-right">
                        Date added
                      </Table.Head>
                    </Table.Row>
                  </Table.Header>

                  <Table.Body>
                    {#each filteredDrivers as driver (driver.id)}
                      <Table.Row class="transition-colors">
                        <Table.Cell class="py-4 pl-4">
                          <div class="flex items-center gap-3">
                            <Avatar.Root class="size-10 shrink-0 border">
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
                                class="font-medium hover:underline"
                              >
                                {driver.name}
                              </a>
                              <p
                                class="mt-0.5 truncate text-xs text-muted-foreground"
                              >
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

                        <Table.Cell
                          class="pr-4 text-right text-sm text-muted-foreground"
                        >
                          {formatDate(driver.createdAt)}
                        </Table.Cell>
                      </Table.Row>
                    {/each}
                  </Table.Body>
                </Table.Root>
              </div>

              <div
                class="flex items-center justify-between py-4 text-xs text-muted-foreground"
              >
                <p>
                  Showing {filteredDrivers.length} of {(driversQuery.data ?? [])
                    .length} drivers
                </p>

                {#if searchQuery}
                  <button
                    class="transition-colors hover:text-foreground"
                    onclick={() => (searchQuery = "")}
                  >
                    Clear filters
                  </button>
                {/if}
              </div>
            {/if}
          </div>
        </section>
      </main>
    </div>

    <!-- Create driver dialog -->
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
