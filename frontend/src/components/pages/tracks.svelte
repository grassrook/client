<script lang="ts">
  import { MapPin } from "@lucide/svelte";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";

  import {
    CreateTrack,
    DeleteTrack,
    GetTracks,
    UpdateTrack,
  } from "$wails/go/main/App";

  import * as AlertDialog from "../ui/alert-dialog";
  import Button from "../ui/button/button.svelte";
  import * as Dialog from "../ui/dialog";
  import { Input } from "../ui/input";
  import { Label } from "../ui/label";
  import Separator from "../ui/separator/separator.svelte";
  import * as Table from "../ui/table";

  type Track = {
    id: string;
    name: string;
    createdAt: number;
    updatedAt: number;
  };

  const queryClient = useQueryClient();

  const tracksQuery = createQuery(() => ({
    queryKey: ["tracks"],
    queryFn: () => GetTracks(),
  }));

  let dialogOpen = $state(false);
  let deleteDialogOpen = $state(false);
  let editingTrack = $state<Track | null>(null);
  let selectedTrack = $state<Track | null>(null);
  let name = $state("");
  let errorMessage = $state("");
  let mounted = $state(false);

  const refreshTracks = async () => {
    await queryClient.invalidateQueries({ queryKey: ["tracks"] });
  };

  const createTrackMutation = createMutation(() => ({
    mutationFn: (name: string) => CreateTrack({ name }),
    onSuccess: async () => {
      await refreshTracks();
      closeDialog();
    },
    onError: (error) => {
      errorMessage = String(error);
    },
  }));

  const updateMutation = createMutation(() => ({
    mutationFn: (input: { id: string; name: string }) => UpdateTrack(input),
    onSuccess: async () => {
      await refreshTracks();
      closeDialog();
    },
    onError: (error) => {
      errorMessage = String(error);
    },
  }));

  const deleteMutation = createMutation(() => ({
    mutationFn: (id: string) => DeleteTrack(id),
    onSuccess: async () => {
      await refreshTracks();
      deleteDialogOpen = false;
      selectedTrack = null;
    },
    onError: (error) => {
      errorMessage = String(error);
    },
  }));

  function openCreateDialog() {
    editingTrack = null;
    name = "";
    errorMessage = "";
    dialogOpen = true;
  }

  function openEditDialog(track: Track) {
    editingTrack = track;
    name = track.name;
    errorMessage = "";
    dialogOpen = true;
  }

  function closeDialog() {
    dialogOpen = false;
    editingTrack = null;
    name = "";
    errorMessage = "";
  }

  function openDeleteDialog(track: Track) {
    selectedTrack = track;
    errorMessage = "";
    deleteDialogOpen = true;
  }

  function submitTrack(event: SubmitEvent) {
    event.preventDefault();

    const trimmedName = name.trim();

    if (!trimmedName) {
      errorMessage = "Track name is required.";
      return;
    }

    errorMessage = "";

    if (editingTrack) {
      updateMutation.mutate({
        id: editingTrack.id,
        name: trimmedName,
      });
    } else {
      createTrackMutation.mutate(trimmedName);
    }
  }

  function confirmDelete() {
    if (!selectedTrack) return;

    errorMessage = "";
    deleteMutation.mutate(selectedTrack.id);
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
          <h1 class="text-2xl font-semibold tracking-tight">Tracks</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Manage the tracks used for judging.
          </p>
        </div>

        <Button
          class="bg-blue-600 text-white hover:bg-blue-700"
          onclick={openCreateDialog}
        >
          + New Track
        </Button>
      </header>

      <Separator />

      <main class="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <section>
          <!-- Directory heading -->
          <div class="flex items-end justify-between">
            <div>
              <div class="flex items-center gap-2">
                <h2 class="text-base font-semibold">Track directory</h2>

                {#if !tracksQuery.isPending && !tracksQuery.isError}
                  <span
                    class="rounded-md border px-2 py-0.5 text-xs font-medium text-muted-foreground"
                  >
                    {(tracksQuery.data ?? []).length}
                  </span>
                {/if}
              </div>

              <p class="mt-1 text-sm text-muted-foreground">
                View and manage your registered tracks.
              </p>
            </div>
          </div>

          <!-- Track table -->
          <div class="mt-6">
            {#if tracksQuery.isPending}
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
            {:else if tracksQuery.isError}
              <div class="flex flex-col items-center gap-3 py-16 text-center">
                <p class="font-medium">Couldn't load tracks</p>
                <p class="text-sm text-muted-foreground">
                  Something went wrong while fetching the track directory.
                </p>
                <Button variant="outline" onclick={() => tracksQuery.refetch()}>
                  Try again
                </Button>
              </div>
            {:else if !(tracksQuery.data ?? []).length}
              <div
                class="flex flex-col items-center border-y py-16 text-center"
              >
                <MapPin class="mb-4 size-9 text-muted-foreground" />
                <h3 class="font-medium">No tracks yet</h3>
                <p class="mt-1 max-w-sm text-sm text-muted-foreground">
                  Add a track to start managing the locations used for judging.
                </p>
                <Button class="mt-5" onclick={openCreateDialog}>
                  + New Track
                </Button>
              </div>
            {:else}
              <div class="overflow-x-auto border-y">
                <Table.Root>
                  <Table.Header>
                    <Table.Row class="hover:bg-transparent">
                      <Table.Head class="h-11 min-w-64 pl-4">Track</Table.Head>
                      <Table.Head class="h-11 min-w-32">Created</Table.Head>
                      <Table.Head class="h-11 min-w-32">
                        Last updated
                      </Table.Head>
                      <Table.Head class="h-11 min-w-40 pr-4 text-right">
                        Actions
                      </Table.Head>
                    </Table.Row>
                  </Table.Header>

                  <Table.Body>
                    {#each tracksQuery.data ?? [] as track (track.id)}
                      <Table.Row class="transition-colors">
                        <Table.Cell class="py-4 pl-4">
                          <span class="font-medium">{track.name}</span>
                        </Table.Cell>

                        <Table.Cell class="text-sm text-muted-foreground">
                          {formatDate(track.createdAt)}
                        </Table.Cell>

                        <Table.Cell class="text-sm text-muted-foreground">
                          {formatDate(track.updatedAt)}
                        </Table.Cell>

                        <Table.Cell class="pr-4">
                          <div class="flex justify-end gap-2">
                            <Button
                              variant="outline"
                              size="sm"
                              onclick={() => openEditDialog(track)}
                            >
                              Edit
                            </Button>

                            <Button
                              variant="destructive"
                              size="sm"
                              onclick={() => openDeleteDialog(track)}
                            >
                              Delete
                            </Button>
                          </div>
                        </Table.Cell>
                      </Table.Row>
                    {/each}
                  </Table.Body>
                </Table.Root>
              </div>

              <div class="py-4 text-xs text-muted-foreground">
                Showing {(tracksQuery.data ?? []).length} tracks
              </div>
            {/if}
          </div>
        </section>
      </main>
    </div>
  </div>

  <!-- Create / Edit dialog -->
  <Dialog.Root
    bind:open={dialogOpen}
    onOpenChange={(open) => {
      if (!open) closeDialog();
    }}
  >
    <Dialog.Content>
      <Dialog.Header>
        <Dialog.Title>
          {editingTrack ? "Edit track" : "Create track"}
        </Dialog.Title>

        <Dialog.Description>
          {editingTrack
            ? "Update the track name."
            : "Enter the name of the new track."}
        </Dialog.Description>
      </Dialog.Header>

      <form class="space-y-4" onsubmit={submitTrack}>
        <div class="space-y-2">
          <Label for="track-name">Track name</Label>
          <Input
            id="track-name"
            bind:value={name}
            required
            maxlength={255}
            placeholder="e.g. Ebisu"
          />
        </div>

        {#if errorMessage}
          <p class="text-sm text-destructive">{errorMessage}</p>
        {/if}

        <Dialog.Footer>
          <Button
            type="button"
            variant="outline"
            onclick={closeDialog}
            disabled={createTrackMutation.isPending || updateMutation.isPending}
          >
            Cancel
          </Button>

          <Button
            type="submit"
            disabled={createTrackMutation.isPending || updateMutation.isPending}
          >
            {createTrackMutation.isPending || updateMutation.isPending
              ? "Saving..."
              : editingTrack
                ? "Save changes"
                : "Create track"}
          </Button>
        </Dialog.Footer>
      </form>
    </Dialog.Content>
  </Dialog.Root>

  <!-- Delete confirmation -->
  <AlertDialog.Root
    bind:open={deleteDialogOpen}
    onOpenChange={(open) => {
      if (!open && !deleteMutation.isPending) {
        selectedTrack = null;
        errorMessage = "";
      }
    }}
  >
    <AlertDialog.Content>
      <AlertDialog.Header>
        <AlertDialog.Title>Delete track?</AlertDialog.Title>

        <AlertDialog.Description>
          Are you sure you want to delete
          <strong class="text-foreground">{selectedTrack?.name}</strong>? This
          action cannot be undone.
        </AlertDialog.Description>
      </AlertDialog.Header>

      {#if errorMessage}
        <p class="text-sm text-destructive">{errorMessage}</p>
      {/if}

      <AlertDialog.Footer>
        <AlertDialog.Cancel disabled={deleteMutation.isPending}>
          Cancel
        </AlertDialog.Cancel>

        <AlertDialog.Action
          class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
          disabled={deleteMutation.isPending}
          onclick={(event) => {
            event.preventDefault();
            confirmDelete();
          }}
        >
          {deleteMutation.isPending ? "Deleting..." : "Delete"}
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}
