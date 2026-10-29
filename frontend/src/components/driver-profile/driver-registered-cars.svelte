<script lang="ts">
  import { CarFront, ImageOff, Pencil, Plus, Trash2 } from "@lucide/svelte";

  import { Button } from "$src/components/ui/button";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
  } from "$src/components/ui/dialog";

  type RegisteredCar = {
    id: string;
    name: string;
    brand: string;
    model: string;
    year: number;
    registration: string;
    imageUrl?: string;
  };

  type Props = {
    cars?: RegisteredCar[];
    onAdd?: () => void;
    onEdit?: (car: RegisteredCar) => void;
    onDelete?: (car: RegisteredCar) => void;
  };

  let {
    cars = [],
    onAdd = () => {},
    onEdit = () => {},
    onDelete = () => {},
  }: Props = $props();

  let previewCar = $state<RegisteredCar | null>(null);
  let imageFailed = $state(false);

  function openPreview(car: RegisteredCar) {
    previewCar = car;
    imageFailed = false;
  }

  function closePreview() {
    previewCar = null;
    imageFailed = false;
  }
</script>

<section class="overflow-hidden rounded-lg border">
  <header
    class="flex flex-wrap items-center justify-between gap-3 border-b px-5 py-4"
  >
    <div class="flex items-center gap-3">
      <div>
        <h2 class="text-sm font-semibold tracking-tight">Registered cars</h2>
        <p class="mt-0.5 text-xs text-muted-foreground">
          Cars associated with this driver
        </p>
      </div>
    </div>

    <Button
      size="sm"
      class="bg-blue-600 text-white hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600"
      onclick={onAdd}
    >
      <Plus class="mr-1.5 size-4" />
      Add car
    </Button>
  </header>

  {#if cars.length === 0}
    <div class="flex flex-col items-center px-6 py-12 text-center">
      <div class="flex size-14 items-center justify-center rounded-xl border">
        <CarFront class="size-7 text-muted-foreground" />
      </div>

      <h3 class="mt-4 text-sm font-semibold">No registered cars</h3>
      <p class="mt-1 max-w-xs text-sm text-muted-foreground">
        Add a car to associate it with this driver.
      </p>

      <Button
        class="mt-5 bg-blue-600 text-white hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600"
        size="sm"
        onclick={onAdd}
      >
        <Plus class="mr-1.5 size-4" />
        Add first car
      </Button>
    </div>
  {:else}
    <div class="divide-y">
      {#each cars as car (car.id)}
        <div
          class="group flex items-center gap-3 px-5 py-4 transition-colors sm:gap-4"
        >
          <button
            type="button"
            class="relative flex size-16 shrink-0 items-center justify-center overflow-hidden rounded-lg border transition-colors hover:border-blue-500/50 sm:size-20"
            aria-label={`Preview images for ${car.name}`}
            onclick={() => openPreview(car)}
          >
            {#if car.imageUrl && !imageFailed}
              <img
                src={car.imageUrl}
                alt={car.name}
                class="size-full object-cover"
                onerror={() => (imageFailed = true)}
              />
            {:else}
              <CarFront class="size-7 text-muted-foreground" />
            {/if}
          </button>

          <div class="min-w-0 flex-1">
            <h3 class="truncate text-sm font-medium">{car.name}</h3>

            <p class="mt-1 text-sm text-muted-foreground">
              {car.brand}
              {car.model}
            </p>
          </div>

          <div
            class="flex shrink-0 items-center gap-1 opacity-100 sm:opacity-70 sm:group-hover:opacity-100 sm:group-focus-within:opacity-100"
          >
            <Button
              variant="ghost"
              size="icon"
              class="size-8 text-muted-foreground hover:text-blue-600 dark:hover:text-blue-400"
              aria-label={`Edit ${car.name}`}
              onclick={() => onEdit(car)}
            >
              <Pencil class="size-4" />
            </Button>

            <Button
              variant="ghost"
              size="icon"
              class="size-8 text-muted-foreground hover:text-destructive"
              aria-label={`Delete ${car.name}`}
              onclick={() => onDelete(car)}
            >
              <Trash2 class="size-4" />
            </Button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  <footer class="border-t px-5 py-3">
    <p class="text-xs text-muted-foreground">
      {cars.length}
      {cars.length === 1 ? "car" : "cars"} registered
    </p>
  </footer>
</section>

<Dialog
  open={previewCar !== null}
  onOpenChange={(open) => {
    if (!open) closePreview();
  }}
>
  <DialogContent class="max-w-3xl overflow-hidden p-0">
    {#if previewCar}
      <div class="p-5 pb-0">
        <DialogHeader>
          <DialogTitle>{previewCar.name}</DialogTitle>
          <DialogDescription>
            {previewCar.brand}
            {previewCar.model} · {previewCar.year}
          </DialogDescription>
        </DialogHeader>
      </div>

      <div
        class="flex min-h-64 items-center justify-center border-y p-4 sm:min-h-96"
      >
        {#if previewCar.imageUrl && !imageFailed}
          <img
            src={previewCar.imageUrl}
            alt={previewCar.name}
            class="max-h-[65vh] w-full rounded-md object-contain"
            onerror={() => (imageFailed = true)}
          />
        {:else}
          <div class="flex flex-col items-center gap-2 text-muted-foreground">
            <ImageOff class="size-10" />
            <p class="text-sm">No image available</p>
          </div>
        {/if}
      </div>

      <div class="flex items-center justify-between gap-3 p-5 pt-3">
        <p class="text-xs text-muted-foreground">
          Registration: {previewCar.registration || "Not provided"}
        </p>

        <Button variant="outline" size="sm" onclick={closePreview}>
          Close
        </Button>
      </div>
    {/if}
  </DialogContent>
</Dialog>
