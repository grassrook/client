<script lang="ts">
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";
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
</script>

<div in:fade={{ duration: 300 }}>
  <div
    class="mx-auto flex max-w-6xl items-end justify-between px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
  >
    <div>
      <h1 class="text-2xl font-semibold">Tracks</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Manage the tracks used for judging.
      </p>
    </div>

    <Button onclick={openCreateDialog}>+ New Track</Button>
  </div>

  <Separator />

  <div class="mx-auto max-w-6xl px-4 py-6 sm:px-6 lg:px-8">
    {#if tracksQuery.isPending}
      <p class="text-sm text-muted-foreground">Loading tracks...</p>
    {:else if tracksQuery.isError}
      <div class="text-sm text-destructive">
        Failed to load tracks.
        <button class="ml-2 underline" onclick={() => tracksQuery.refetch()}>
          Retry
        </button>
      </div>
    {:else if (tracksQuery.data ?? []).length === 0}
      <div class="rounded-lg border border-dashed p-10 text-center">
        <h2 class="font-medium">No tracks yet</h2>
        <p class="mt-1 text-sm text-muted-foreground">
          Create a track to get started.
        </p>
        <div class="mt-4">
          <Button onclick={openCreateDialog}>Create track</Button>
        </div>
      </div>
    {:else}
      <div class="overflow-x-auto rounded-lg border">
        <table class="w-full text-left text-sm">
          <thead class="bg-muted/50 text-muted-foreground">
            <tr>
              <th class="px-4 py-3 font-medium">Name</th>
              <th class="px-4 py-3 font-medium">Created</th>
              <th class="px-4 py-3 font-medium">Last updated</th>
              <th class="px-4 py-3 text-right font-medium">Actions</th>
            </tr>
          </thead>

          <tbody>
            {#each tracksQuery.data ?? [] as track (track.id)}
              <tr class="border-t">
                <td class="px-4 py-3 font-medium">{track.name}</td>
                <td class="px-4 py-3 text-muted-foreground">
                  {formatDate(track.createdAt)}
                </td>
                <td class="px-4 py-3 text-muted-foreground">
                  {formatDate(track.updatedAt)}
                </td>
                <td class="px-4 py-3">
                  <div class="flex justify-end gap-2">
                    <Button
                      variant="outline"
                      onclick={() => openEditDialog(track)}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="destructive"
                      onclick={() => openDeleteDialog(track)}
                    >
                      Delete
                    </Button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</div>

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
