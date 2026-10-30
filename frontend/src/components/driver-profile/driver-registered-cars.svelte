<script lang="ts">
  import { CarFront, ImageOff, Pencil, Plus, Trash2, X } from "@lucide/svelte";
  import {
    createMutation,
    createQuery,
    useQueryClient,
  } from "@tanstack/svelte-query";

  import * as Alert from "$src/components/ui/alert";
  import * as AlertDialog from "$src/components/ui/alert-dialog";
  import { Button } from "$src/components/ui/button";
  import * as Carousel from "$src/components/ui/carousel";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
  } from "$src/components/ui/dialog";
  import { Input } from "$src/components/ui/input";
  import { Skeleton } from "$src/components/ui/skeleton";
  import * as Table from "$src/components/ui/table";
  import { Textarea } from "$src/components/ui/textarea";
  import {
    CreateRegisteredCar,
    DeleteCar,
    GetCarPhoto,
    GetDriverCarList,
    ListCarMakes,
  } from "$wails/go/main/App";
  import type { car } from "$wails/go/models";

  import Dropzone from "../ui/dropzone/dropzone.svelte";
  import Label from "../ui/label/label.svelte";
  import * as Select from "../ui/select";

  type Props = {
    driverID: string;
    onEdit?: (car: car.RegisteredCar) => void | Promise<void>;
  };

  let { driverID, onEdit = () => {} }: Props = $props();

  const queryClient = useQueryClient();

  const carMakesQuery = createQuery(() => ({
    queryKey: ["car-makes"],
    queryFn: () => ListCarMakes(),
  }));

  let carMakes = $derived(carMakesQuery.data ?? []);

  const carsQuery = createQuery(() => ({
    queryKey: ["driver-cars", driverID],
    queryFn: () => GetDriverCarList(driverID),
    enabled: !!driverID,
  }));

  let cars = $derived(carsQuery.data ?? []);

  let selectedPhotos = $state<File[]>([]);

  const createCarMutation = createMutation(() => ({
    mutationFn: async (input: {
      name: string;
      description: string;
      carMakeID: string;
      driverID: string;
    }) => {
      const photos = await Promise.all(selectedPhotos.map(fileToBase64));

      return CreateRegisteredCar({
        ...input,
        photos,
      });
    },

    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["driver-cars", driverID],
      });

      selectedPhotos = [];
      closeCreateDialog();
    },
  }));

  const deleteCarMutation = createMutation(() => ({
    mutationFn: (id: string) => DeleteCar(id),

    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: ["driver-cars", driverID],
      });

      carToDelete = null;
    },
  }));

  let createDialogOpen = $state(false);

  let name = $state("");
  let description = $state("");
  let carMakeID = $state("");

  let previewCar = $state<car.RegisteredCar | null>(null);
  let carToDelete = $state<car.RegisteredCar | null>(null);

  let actionError = $state("");

  function resetForm() {
    name = "";
    description = "";
    carMakeID = "";
    selectedPhotos = [];
  }

  function closeCreateDialog() {
    createDialogOpen = false;
    resetForm();
    actionError = "";
  }

  function openCreateDialog() {
    resetForm();
    actionError = "";
    createDialogOpen = true;
  }

  function handlePhotosChange(files: File[]) {
    selectedPhotos = files;
  }

  function fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();

      reader.onload = () => {
        if (typeof reader.result !== "string") {
          reject(new Error(`Failed to read ${file.name}`));
          return;
        }

        resolve(reader.result);
      };

      reader.onerror = () => reject(reader.error);
      reader.readAsDataURL(file);
    });
  }

  function getCarMakeName(carMakeID: string) {
    return carMakes.find((make) => make.id === carMakeID)?.name ?? carMakeID;
  }

  async function handleCreate() {
    actionError = "";

    if (!name.trim()) {
      actionError = "Car name is required.";
      return;
    }

    if (!carMakeID.trim()) {
      actionError = "Car make is required.";
      return;
    }

    try {
      await createCarMutation.mutateAsync({
        name: name.trim(),
        description: description.trim(),
        carMakeID: carMakeID.trim(),
        driverID,
      });
    } catch (error) {
      actionError = String(error);
    }
  }

  function requestDelete(car: car.RegisteredCar) {
    carToDelete = car;
    actionError = "";
  }

  async function confirmDelete() {
    if (!carToDelete) return;

    try {
      await deleteCarMutation.mutateAsync(carToDelete.id);
    } catch (error) {
      actionError = String(error);
      carToDelete = null;
    }
  }

  async function handleEdit(car: car.RegisteredCar) {
    actionError = "";

    try {
      await onEdit(car);

      await queryClient.invalidateQueries({
        queryKey: ["driver-cars", driverID],
      });
    } catch (error) {
      actionError = String(error);
    }
  }

  let previewPhotos = $state<string[]>([]);
  let previewLoading = $state(false);
  let previewError = $state("");

  let previewRequest = 0;

  async function openPreview(car: car.RegisteredCar) {
    const request = ++previewRequest;

    previewCar = car;
    previewPhotos = [];
    previewError = "";
    previewLoading = true;

    try {
      const photos = await Promise.all(
        (car.photos ?? []).map((photo) => GetCarPhoto(photo)),
      );

      if (request !== previewRequest) return;

      previewPhotos = photos;
    } catch (error) {
      if (request !== previewRequest) return;

      previewError = String(error);
    } finally {
      if (request === previewRequest) {
        previewLoading = false;
      }
    }
  }

  function closePreview() {
    previewRequest++;
    previewCar = null;
    previewPhotos = [];
    previewError = "";
    previewLoading = false;
  }
</script>

<section class="overflow-hidden rounded-lg border">
  <header
    class="flex flex-wrap items-center justify-between gap-3 border-b px-5 py-4"
  >
    <div>
      <h2 class="text-sm font-semibold tracking-tight">Registered cars</h2>
      <p class="mt-0.5 text-xs text-muted-foreground">
        Cars associated with this driver
      </p>
    </div>

    <Button
      size="sm"
      class="bg-blue-600 text-white hover:bg-blue-700"
      onclick={openCreateDialog}
    >
      <Plus class="mr-1.5 size-4" />
      Add car
    </Button>
  </header>

  {#if actionError}
    <div class="border-b px-5 py-3">
      <Alert.Root variant="destructive">
        <Alert.Title>Something went wrong</Alert.Title>
        <Alert.Description class="flex items-center justify-between gap-3">
          <span>{actionError}</span>
          <Button
            variant="ghost"
            size="icon"
            class="size-7"
            aria-label="Dismiss error"
            onclick={() => (actionError = "")}
          >
            <X class="size-4" />
          </Button>
        </Alert.Description>
      </Alert.Root>
    </div>
  {/if}

  {#if carsQuery.isPending}
    <div class="space-y-4 px-5 py-6">
      {#each Array.from({ length: 3 }, (_, i) => i) as i (i)}
        <div class="flex items-center gap-3">
          <Skeleton class="size-11 rounded-lg" />
          <div class="flex-1 space-y-2">
            <Skeleton class="h-4 w-32" />
            <Skeleton class="h-3 w-48" />
          </div>
          <Skeleton class="h-8 w-16" />
        </div>
      {/each}
    </div>
  {:else if carsQuery.isError}
    <div class="px-5 py-12 text-center">
      <p class="text-sm text-destructive">Failed to load registered cars.</p>
      <Button
        variant="outline"
        size="sm"
        class="mt-3"
        onclick={() => carsQuery.refetch()}
      >
        Retry
      </Button>
    </div>
  {:else if cars.length === 0}
    <div class="flex flex-col items-center px-6 py-12 text-center">
      <div class="flex size-14 items-center justify-center rounded-xl border">
        <CarFront class="size-7 text-muted-foreground" />
      </div>

      <h3 class="mt-4 text-sm font-semibold">No registered cars</h3>
      <p class="mt-1 max-w-xs text-sm text-muted-foreground">
        Add a car to associate it with this driver.
      </p>

      <Button
        class="mt-5 bg-blue-600 text-white hover:bg-blue-700"
        size="sm"
        onclick={openCreateDialog}
      >
        <Plus class="mr-1.5 size-4" />
        Add first car
      </Button>
    </div>
  {:else}
    <div class="overflow-x-auto">
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head class="pl-5">Car</Table.Head>
            <Table.Head>Car make</Table.Head>
            <Table.Head>Photos</Table.Head>
            <Table.Head class="text-right pr-5">Actions</Table.Head>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {#each cars as car (car.id)}
            <Table.Row>
              <Table.Cell>
                <div class="flex items-center gap-3 pl-2">
                  <button
                    type="button"
                    class="flex size-11 shrink-0 items-center justify-center rounded-lg border"
                    onclick={() => openPreview(car)}
                    aria-label={`Preview ${car.name}`}
                  >
                    <CarFront class="size-5 text-muted-foreground" />
                  </button>

                  <div class="min-w-0">
                    <p class="truncate font-medium">{car.name}</p>
                    <p
                      class="mt-0.5 max-w-xs truncate text-xs text-muted-foreground"
                    >
                      {car.description || "No description"}
                    </p>
                  </div>
                </div>
              </Table.Cell>

              <Table.Cell class="text-muted-foreground">
                {getCarMakeName(car.carMakeID)}
              </Table.Cell>

              <Table.Cell class="text-muted-foreground">
                {car.photos?.length ?? 0}
              </Table.Cell>

              <Table.Cell>
                <div class="flex justify-end gap-1 pr-2">
                  <Button
                    variant="ghost"
                    size="icon"
                    class="size-8 text-muted-foreground hover:text-blue-600"
                    aria-label={`Edit ${car.name}`}
                    onclick={() => handleEdit(car)}
                  >
                    <Pencil class="size-4" />
                  </Button>

                  <Button
                    variant="ghost"
                    size="icon"
                    class="size-8 text-muted-foreground hover:text-destructive"
                    aria-label={`Delete ${car.name}`}
                    disabled={deleteCarMutation.isPending}
                    onclick={() => requestDelete(car)}
                  >
                    <Trash2 class="size-4" />
                  </Button>
                </div>
              </Table.Cell>
            </Table.Row>
          {/each}
        </Table.Body>
      </Table.Root>
    </div>
  {/if}

  <footer class="border-t px-5 py-3">
    <p class="text-xs text-muted-foreground">
      {cars.length}
      {cars.length === 1 ? "car" : "cars"} registered
    </p>
  </footer>
</section>

<!-- Create car dialog -->
<Dialog
  open={createDialogOpen}
  onOpenChange={(open) => {
    if (!open) {
      closeCreateDialog();
    } else {
      createDialogOpen = true;
    }
  }}
>
  <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-xl">
    <DialogHeader>
      <DialogTitle>Register a car</DialogTitle>
      <DialogDescription>
        Add a car to this driver's registered vehicles.
      </DialogDescription>
    </DialogHeader>

    <form
      class="space-y-5"
      onsubmit={(event) => {
        event.preventDefault();
        handleCreate();
      }}
    >
      <div class="space-y-2">
        <Label for="car-name">Car name</Label>
        <Input
          id="car-name"
          bind:value={name}
          placeholder="e.g. Toyota Supra MK4"
          required
        />
      </div>

      <div class="space-y-2">
        <Label>Car make</Label>

        {#if carMakesQuery.isPending}
          <Skeleton class="h-10 w-full" />
        {:else if carMakesQuery.isError}
          <div class="space-y-2">
            <Alert.Root variant="destructive">
              <Alert.Title>Failed to load car makes</Alert.Title>
              <Alert.Description>
                Please try loading the car makes again.
              </Alert.Description>
            </Alert.Root>

            <Button
              type="button"
              variant="outline"
              size="sm"
              onclick={() => carMakesQuery.refetch()}
            >
              Retry
            </Button>
          </div>
        {:else}
          <Select.Root
            type="single"
            bind:value={carMakeID}
            disabled={carMakes.length === 0}
          >
            <Select.Trigger class="w-full">
              <Select.Value placeholder="Select a car make" />
            </Select.Trigger>

            <Select.Content>
              {#each carMakes as make (make.id)}
                <Select.Item value={make.id} label={make.name}>
                  {make.name}
                </Select.Item>
              {/each}
            </Select.Content>
          </Select.Root>
          {#if carMakes.length === 0}
            <p class="text-xs text-muted-foreground">No car makes available.</p>
          {/if}
        {/if}
      </div>

      <div class="space-y-2">
        <Label for="car-description">
          Description
          <span class="font-normal text-muted-foreground">(optional)</span>
        </Label>

        <Textarea
          id="car-description"
          bind:value={description}
          placeholder="Additional information about this car..."
          rows={3}
          class="resize-none"
        />
      </div>

      <div class="space-y-2">
        <Label>Car photos</Label>

        <Dropzone
          accept="image/jpeg,image/png,image/webp"
          maxSizeMB={10}
          multiple={true}
          autoUpload={false}
          onfileschange={handlePhotosChange}
        />
      </div>

      {#if actionError}
        <Alert.Root variant="destructive">
          <Alert.Title>Failed to create car</Alert.Title>
          <Alert.Description>{actionError}</Alert.Description>
        </Alert.Root>
      {/if}

      <div class="flex justify-end gap-2 pt-4">
        <Button
          type="button"
          variant="outline"
          onclick={closeCreateDialog}
          disabled={createCarMutation.isPending}
        >
          Cancel
        </Button>

        <Button
          type="submit"
          class="bg-blue-600 text-white hover:bg-blue-700"
          disabled={createCarMutation.isPending}
        >
          {#if createCarMutation.isPending}
            Creating...
          {:else}
            <Plus class="mr-1.5 size-4" />
            Create car
          {/if}
        </Button>
      </div>
    </form>
  </DialogContent>
</Dialog>

<!-- Delete confirmation -->
<AlertDialog.Root
  open={carToDelete !== null}
  onOpenChange={(open) => {
    if (!open) carToDelete = null;
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Delete registered car?</AlertDialog.Title>
      <AlertDialog.Description>
        {#if carToDelete}
          Are you sure you want to delete {carToDelete.name}? This action cannot
          be undone.
        {/if}
      </AlertDialog.Description>
    </AlertDialog.Header>

    <AlertDialog.Footer>
      <AlertDialog.Cancel disabled={deleteCarMutation.isPending}>
        Cancel
      </AlertDialog.Cancel>

      <AlertDialog.Action
        class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
        disabled={deleteCarMutation.isPending}
        onclick={confirmDelete}
      >
        {deleteCarMutation.isPending ? "Deleting..." : "Delete"}
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<!-- Car preview -->
<Dialog
  open={previewCar !== null}
  onOpenChange={(open) => {
    if (!open) closePreview();
  }}
>
  <DialogContent class="max-h-[90vh] overflow-y-auto sm:max-w-lg">
    {#if previewCar}
      <DialogHeader>
        <DialogTitle>{previewCar.name}</DialogTitle>
      </DialogHeader>

      {#if previewLoading}
        <div class="flex min-h-48 items-center justify-center">
          <p class="text-sm text-muted-foreground">Loading photos...</p>
        </div>
      {:else if previewError}
        <Alert.Root variant="destructive">
          <Alert.Title>Failed to load photos</Alert.Title>
          <Alert.Description>{previewError}</Alert.Description>
        </Alert.Root>
      {:else if previewPhotos.length > 0}
        <Carousel.Root class="mx-auto w-full max-w-full">
          <Carousel.Content>
            {#each previewPhotos as photo, i (`${previewCar.id}-${i}`)}
              <Carousel.Item>
                <div
                  class="flex aspect-video items-center justify-center overflow-hidden rounded-lg"
                >
                  <img
                    src={photo}
                    alt={`${previewCar.name} photo ${i + 1}`}
                    class="h-full w-full object-contain"
                  />
                </div>
                <p class="mt-2 text-center text-xs text-muted-foreground">
                  {i + 1} / {previewPhotos.length}
                </p>
              </Carousel.Item>
            {/each}
          </Carousel.Content>

          {#if previewPhotos.length > 1}
            <Carousel.Previous />
            <Carousel.Next />
          {/if}
        </Carousel.Root>
      {:else}
        <div
          class="flex min-h-48 flex-col items-center justify-center gap-2 rounded-lg border text-muted-foreground"
        >
          <ImageOff class="size-9" />
          <p class="text-sm">No photos available</p>
        </div>
      {/if}

      <div class="flex justify-end">
        <Button variant="outline" size="sm" onclick={closePreview}>
          Close
        </Button>
      </div>
    {/if}
  </DialogContent>
</Dialog>
