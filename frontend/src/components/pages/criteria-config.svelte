<script lang="ts">
  import Ellipsis from "@lucide/svelte/icons/ellipsis";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";

  import CriteriaConfigDialog from "$src/components/criteria-config/criteria-dialog.svelte";
  import * as AlertDialog from "$src/components/ui/alert-dialog";
  import { Badge } from "$src/components/ui/badge";
  import { Button } from "$src/components/ui/button";
  import * as DropdownMenu from "$src/components/ui/dropdown-menu";
  import { Separator } from "$src/components/ui/separator";
  import * as Table from "$src/components/ui/table";
  import {
    CopyCriteriaConfigById,
    DeleteCriteriaConfig,
    GetCriteriaConfigs,
  } from "$wails/go/main/App";
  import { repository } from "$wails/go/models";

  const queryClient = useQueryClient();

  const configsQuery = createQuery(() => ({
    queryKey: ["criteria-configs"],
    queryFn: () => GetCriteriaConfigs(),
  }));

  let configDialogOpen = $state(false);
  let dialogMode = $state<"create" | "update">("create");
  let selectedConfig = $state<repository.CriteriaConfig | undefined>(undefined);
  let configToDelete = $state<repository.CriteriaConfig | undefined>(undefined);
  let deleteDialogOpen = $state(false);
  let mounted = $state(false);

  function openUpdateDialog(config: repository.CriteriaConfig) {
    selectedConfig = config;
    dialogMode = "update";
    configDialogOpen = true;
  }

  const deleteMutation = createMutation(() => ({
    mutationFn: (id: number) => DeleteCriteriaConfig(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["criteria-configs"],
      });
    },
  }));

  const copyMutation = createMutation(() => ({
    mutationFn: (id: number) => CopyCriteriaConfigById(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["criteria-configs"],
      });
    },
  }));

  function deleteConfig(config: repository.CriteriaConfig) {
    configToDelete = config;
    deleteDialogOpen = true;
  }

  function confirmDelete() {
    if (!configToDelete) return;

    deleteMutation.mutate(configToDelete.id, {
      onSuccess: () => {
        deleteDialogOpen = false;
        configToDelete = undefined;
      },
    });
  }

  onMount(() => {
    mounted = true;
  });

  function exportConfig(config: repository.CriteriaConfig) {
    const json = JSON.stringify(JSON.parse(config.raw), null, 2);
    const blob = new Blob([json], { type: "application/json" });
    const url = URL.createObjectURL(blob);

    const link = document.createElement("a");
    link.href = url;
    link.download = `${config.name}.json`;
    link.click();

    URL.revokeObjectURL(url);
  }
</script>

<!-- Delete confirmation -->
<AlertDialog.Root
  bind:open={deleteDialogOpen}
  onOpenChange={(open) => {
    if (!open && !deleteMutation.isPending) {
      configToDelete = undefined;
    }
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete criteria config?</AlertDialog.Title>
      <AlertDialog.Description>
        Are you sure you want to delete
        <strong>{configToDelete?.name}</strong>? This action cannot be undone.
      </AlertDialog.Description>
    </AlertDialog.Header>

    <AlertDialog.Footer>
      <AlertDialog.Cancel
        disabled={deleteMutation.isPending}
        onclick={() => {
          configToDelete = undefined;
        }}
      >
        Cancel
      </AlertDialog.Cancel>

      <AlertDialog.Action
        onclick={(event) => {
          event.preventDefault();
          confirmDelete();
        }}
        disabled={deleteMutation.isPending}
        class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
      >
        {deleteMutation.isPending ? "Deleting..." : "Delete"}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<CriteriaConfigDialog
  bind:open={configDialogOpen}
  mode={dialogMode}
  config={selectedConfig}
/>

{#if mounted}
  <div in:fade={{ duration: 300 }}>
    <div class="min-h-full">
      <!-- Page heading -->
      <header
        class="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-8 sm:px-6 lg:px-8"
      >
        <div>
          <h1 class="text-2xl font-semibold tracking-tight">
            Criteria Configs
          </h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Manage the criteria used for judging.
          </p>
        </div>

        <Button
          class="bg-blue-600 text-white hover:bg-blue-700"
          onclick={() => {
            dialogMode = "create";
            configDialogOpen = true;
            selectedConfig = undefined;
          }}
        >
          + New Config
        </Button>
      </header>

      <Separator />

      <main class="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <section>
          <!-- Directory heading -->
          <div>
            <div class="flex items-center gap-2">
              <h2 class="text-base font-semibold">Configuration directory</h2>

              {#if !configsQuery.isPending && !configsQuery.isError}
                <span
                  class="rounded-md border px-2 py-0.5 text-xs font-medium text-muted-foreground"
                >
                  {(configsQuery.data ?? []).length}
                </span>
              {/if}
            </div>

            <p class="mt-1 text-sm text-muted-foreground">
              View and manage your judging configurations.
            </p>
          </div>

          {#if deleteMutation.isError}
            <p class="mt-4 text-sm text-destructive">
              Failed to delete config: {String(deleteMutation.error)}
            </p>
          {/if}

          {#if copyMutation.isError}
            <p class="mt-4 text-sm text-destructive">
              Failed to copy config: {String(copyMutation.error)}
            </p>
          {/if}

          <!-- Configuration table -->
          <div class="mt-6">
            {#if configsQuery.isPending}
              <div class="space-y-5 py-5">
                {#each [1, 2, 3] as item (item)}
                  <div class="flex items-center gap-4">
                    <div class="h-4 w-36 animate-pulse rounded bg-muted"></div>
                    <div class="flex-1">
                      <div
                        class="h-4 w-56 max-w-full animate-pulse rounded bg-muted"
                      ></div>
                    </div>
                  </div>
                {/each}
              </div>
            {:else if configsQuery.isError}
              <div class="flex flex-col items-center gap-3 py-16 text-center">
                <p class="font-medium">Couldn't load configurations</p>
                <p class="text-sm text-muted-foreground">
                  Failed to load criteria configs:
                  {String(configsQuery.error)}
                </p>
                <Button
                  variant="outline"
                  onclick={() => configsQuery.refetch()}
                >
                  Try again
                </Button>
              </div>
            {:else if !(configsQuery.data ?? []).length}
              <div
                class="flex flex-col items-center border-y py-16 text-center"
              >
                <h3 class="font-medium">No configurations yet</h3>
                <p class="mt-1 max-w-sm text-sm text-muted-foreground">
                  Create a criteria configuration to start defining your judging
                  categories.
                </p>
                <Button
                  class="mt-5"
                  onclick={() => {
                    dialogMode = "create";
                    configDialogOpen = true;
                    selectedConfig = undefined;
                  }}
                >
                  + New Config
                </Button>
              </div>
            {:else}
              <div class="overflow-x-auto border-y">
                <Table.Root>
                  <Table.Header>
                    <Table.Row class="hover:bg-transparent">
                      <Table.Head class="h-11 min-w-48 pl-4">Name</Table.Head>
                      <Table.Head class="h-11 min-w-64">Categories</Table.Head>
                      <Table.Head class="h-11 w-20 pr-4 text-right">
                        Actions
                      </Table.Head>
                    </Table.Row>
                  </Table.Header>

                  <Table.Body>
                    {#each configsQuery.data ?? [] as config (config.id)}
                      <Table.Row class="transition-colors">
                        <Table.Cell class="py-4 pl-4 font-medium">
                          {config.name}
                        </Table.Cell>

                        <Table.Cell>
                          <div class="flex flex-wrap gap-2">
                            {#each [config.categories.common, config.categories.lead, config.categories.chase] as category, idx (idx)}
                              <Badge variant="secondary">
                                {category.label} · {Math.round(
                                  category.tags.reduce(
                                    (total, tag) =>
                                      total +
                                      (tag.type !== "deduction"
                                        ? (tag.max ?? 0)
                                        : 0),
                                    0,
                                  ),
                                )}
                              </Badge>
                            {/each}
                          </div>
                        </Table.Cell>

                        <Table.Cell class="pr-4">
                          <div class="flex justify-end">
                            <DropdownMenu.Root>
                              <DropdownMenu.Trigger>
                                {#snippet child({ props })}
                                  <Button
                                    {...props}
                                    variant="ghost"
                                    size="icon"
                                    aria-label="Open actions menu"
                                  >
                                    <Ellipsis class="size-4" />
                                  </Button>
                                {/snippet}
                              </DropdownMenu.Trigger>

                              <DropdownMenu.Content align="end">
                                <DropdownMenu.Item
                                  onclick={() => exportConfig(config)}
                                >
                                  Export to JSON
                                </DropdownMenu.Item>

                                <DropdownMenu.Item
                                  onclick={() => openUpdateDialog(config)}
                                >
                                  Edit
                                </DropdownMenu.Item>

                                <DropdownMenu.Item
                                  onclick={() => copyMutation.mutate(config.id)}
                                  disabled={copyMutation.isPending}
                                >
                                  Copy
                                </DropdownMenu.Item>

                                <DropdownMenu.Separator />

                                <DropdownMenu.Item
                                  variant="destructive"
                                  onclick={() => deleteConfig(config)}
                                  disabled={deleteMutation.isPending}
                                >
                                  Delete
                                </DropdownMenu.Item>
                              </DropdownMenu.Content>
                            </DropdownMenu.Root>
                          </div>
                        </Table.Cell>
                      </Table.Row>
                    {/each}
                  </Table.Body>
                </Table.Root>
              </div>

              <div class="py-4 text-xs text-muted-foreground">
                Showing {(configsQuery.data ?? []).length} configurations
              </div>
            {/if}
          </div>
        </section>
      </main>
    </div>
  </div>
{/if}
