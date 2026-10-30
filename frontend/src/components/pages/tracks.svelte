<script lang="ts">
  import { MapPin, MapPinned } from "@lucide/svelte";
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
  import * as Empty from "../ui/empty";
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
  let mounted = $state(false);

  onMount(() => {
    mounted = true;
  });
</script>

{#if mounted}
  <div in:fade={{ duration: 300 }}>
    <div
      class="mx-auto flex max-w-6xl items-center justify-between px-4 py-6 sm:px-6 lg:px-8 lg:py-8"
    >
      <div class="flex items-start gap-3">
        <div
          class="flex size-11 items-center justify-center rounded-xl border bg-muted/50"
        >
          <MapPinned class="size-5 text-muted-foreground" />
        </div>
        <div>
          <h1 class="text-2xl font-semibold">Tracks</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Manage the tracks used for judging.
          </p>
        </div>
      </div>

      <Button
        class="bg-blue-600 text-white hover:bg-blue-700"
        onclick={openCreateDialog}>+ New Track</Button
      >
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
      {:else}
        <div class="overflow-x-auto rounded-lg border">
          <Table.Root>
            <Table.Header class="bg-muted/50">
              <Table.Row class="hover:bg-transparent">
                <Table.Head>Name</Table.Head>
                <Table.Head>Created</Table.Head>
                <Table.Head>Last updated</Table.Head>
                <Table.Head class="text-right">Actions</Table.Head>
              </Table.Row>
            </Table.Header>

            <Table.Body>
              {#if (tracksQuery.data ?? []).length === 0}
                <Table.Row>
                  <Table.Cell colspan={4} class="p-0">
                    <Empty.Root class="py-12">
                      <Empty.Header>
                        <Empty.Media variant="icon">
                          <MapPin />
                        </Empty.Media>
                        <Empty.Title>No tracks yet</Empty.Title>
                        <Empty.Description>
                          Add a track to start managing the locations used for
                          judging.
                        </Empty.Description>
                      </Empty.Header>
                      <Empty.Content>
                        <Button onclick={openCreateDialog}>+ New Track</Button>
                      </Empty.Content>
                    </Empty.Root>
                  </Table.Cell>
                </Table.Row>
              {:else}
                {#each tracksQuery.data ?? [] as track (track.id)}
                  <Table.Row>
                    <Table.Cell class="font-medium">
                      {track.name}
                    </Table.Cell>
                    <Table.Cell class="text-muted-foreground">
                      {formatDate(track.createdAt)}
                    </Table.Cell>
                    <Table.Cell class="text-muted-foreground">
                      {formatDate(track.updatedAt)}
                    </Table.Cell>
                    <Table.Cell class="text-right">
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
                    </Table.Cell>
                  </Table.Row>
                {/each}
              {/if}
            </Table.Body>
          </Table.Root>
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
{/if}
