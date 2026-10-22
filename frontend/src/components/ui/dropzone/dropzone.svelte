<script lang="ts">
  import AlertCircle from "@lucide/svelte/icons/alert-circle";
  import Check from "@lucide/svelte/icons/check";
  import CloudUpload from "@lucide/svelte/icons/cloud-upload";
  import FileIcon from "@lucide/svelte/icons/file";
  import FileText from "@lucide/svelte/icons/file-text";
  import ImageIcon from "@lucide/svelte/icons/image";
  import Trash2 from "@lucide/svelte/icons/trash-2";
  import X from "@lucide/svelte/icons/x";

  interface FileItem {
    id: string;
    file: File;
    name: string;
    size: string;
    type: string;
    preview: string | null;
    progress: number;
    status: "uploading" | "complete" | "error";
  }

  interface Props {
    accept?: string;
    maxSizeMB?: number;
    multiple?: boolean;
    disabled?: boolean;
    autoUpload?: boolean;
    showFileList?: boolean;
    onfileschange?: (files: File[]) => void;
  }

  let {
    accept = "*",
    maxSizeMB = 10,
    multiple = true,
    disabled = false,
    autoUpload = true,
    showFileList = true,
    onfileschange,
  }: Props = $props();

  let isDragging = $state(false);
  let fileList = $state<FileItem[]>([]);
  let errorMessage = $state<string | null>(null);
  let fileInputRef: HTMLInputElement | null = $state(null);

  let totalFiles = $derived(fileList.length);

  function formatBytes(bytes: number, decimals = 1): string {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + " " + sizes[i];
  }

  function validateFile(file: File): string | null {
    if (file.size > maxSizeMB * 1024 * 1024) {
      return `File "${file.name}" exceeds the maximum allowed size of ${maxSizeMB}MB.`;
    }
    return null;
  }

  function processFiles(incomingFiles: FileList | File[]) {
    if (disabled) return;
    errorMessage = null;

    const validFilesToProcess: File[] = [];

    for (const file of Array.from(incomingFiles)) {
      const error = validateFile(file);
      if (error) {
        errorMessage = error;
        continue;
      }
      validFilesToProcess.push(file);
      if (!multiple) break;
    }

    if (validFilesToProcess.length === 0) return;

    const newItems: FileItem[] = validFilesToProcess.map((file) => {
      const id = Math.random().toString(36).substring(2, 9);
      const isImg = file.type.startsWith("image/");
      const preview = isImg ? URL.createObjectURL(file) : null;

      return {
        id,
        file,
        name: file.name,
        size: formatBytes(file.size),
        type: file.type || "Unknown format",
        preview,
        progress: autoUpload ? 0 : 100,
        status: autoUpload ? "uploading" : "complete",
      };
    });

    if (multiple) {
      fileList = [...fileList, ...newItems];
    } else {
      fileList.forEach((item) => {
        if (item.preview) URL.revokeObjectURL(item.preview);
      });
      fileList = newItems;
    }

    onfileschange?.(fileList.map((item) => item.file));

    if (autoUpload) {
      newItems.forEach((item) => simulateUpload(item.id));
    }
  }

  function simulateUpload(id: string) {
    const interval = setInterval(() => {
      const index = fileList.findIndex((f) => f.id === id);
      if (index === -1) {
        clearInterval(interval);
        return;
      }

      if (fileList[index].progress < 100) {
        fileList[index].progress += 15;
        if (fileList[index].progress >= 100) {
          fileList[index].progress = 100;
          fileList[index].status = "complete";
          clearInterval(interval);
        }
      }
    }, 200);
  }

  function removeFile(id: string) {
    const itemToRemove = fileList.find((f) => f.id === id);
    if (itemToRemove?.preview) {
      URL.revokeObjectURL(itemToRemove.preview);
    }
    fileList = fileList.filter((f) => f.id !== id);
    onfileschange?.(fileList.map((item) => item.file));
  }

  function clearAll() {
    fileList.forEach((item) => {
      if (item.preview) URL.revokeObjectURL(item.preview);
    });
    fileList = [];
    errorMessage = null;
    onfileschange?.([]);
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    if (!disabled) isDragging = true;
  }

  function handleDragLeave(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;
    if (disabled || !e.dataTransfer?.files) return;
    processFiles(e.dataTransfer.files);
  }

  function handleInputChange(e: Event) {
    const target = e.target as HTMLInputElement;
    if (target.files) {
      processFiles(target.files);
      target.value = "";
    }
  }
</script>

<div class="w-full space-y-4">
  <!-- Dropzone Box -->
  <div
    aria-label="Upload zone"
    role="button"
    tabindex="0"
    class="relative flex flex-col items-center justify-center w-full min-h-[180px] p-6 border-2 border-dashed rounded-xl transition-all duration-200 cursor-pointer text-center outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2
		{isDragging
      ? 'border-primary bg-primary/10 scale-[0.99]'
      : 'border-zinc-300 dark:border-zinc-800 hover:border-zinc-400 dark:hover:border-zinc-700 bg-zinc-50/50 dark:bg-zinc-900/50'}
		{disabled ? 'opacity-50 cursor-not-allowed' : ''}"
    onpointerdown={() => !disabled && fileInputRef?.click()}
    onkeydown={(e) =>
      (e.key === "Enter" || e.key === " ") &&
      !disabled &&
      fileInputRef?.click()}
    ondragover={handleDragOver}
    ondragleave={handleDragLeave}
    ondrop={handleDrop}
  >
    <input
      bind:this={fileInputRef}
      type="file"
      {accept}
      {multiple}
      {disabled}
      class="hidden"
      onchange={handleInputChange}
    />

    <div class="flex flex-col items-center gap-2 pointer-events-none">
      <div
        class="p-3 rounded-full bg-zinc-100 dark:bg-zinc-800 text-zinc-600 dark:text-zinc-400"
      >
        <CloudUpload class="w-6 h-6" />
      </div>
      <div class="space-y-1">
        <p class="text-sm font-medium text-zinc-900 dark:text-zinc-100">
          <span class="text-primary hover:underline">Click to upload</span> or drag
          and drop
        </p>
        <p class="text-xs text-zinc-500 dark:text-zinc-400">
          {accept === "*" ? "Any file type" : accept} up to {maxSizeMB}MB
        </p>
      </div>
    </div>
  </div>

  <!-- Error Alert -->
  {#if errorMessage}
    <div
      class="flex items-center gap-2 p-3 text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/40 border border-red-200 dark:border-red-900 rounded-lg"
    >
      <AlertCircle class="w-4 h-4 shrink-0" />
      <span class="flex-1">{errorMessage}</span>
      <button
        type="button"
        class="p-1 hover:bg-red-100 dark:hover:bg-red-900/50 rounded"
        onclick={() => (errorMessage = null)}
      >
        <X class="w-3.5 h-3.5" />
      </button>
    </div>
  {/if}

  <!-- File List Header -->
  {#if showFileList && totalFiles > 0}
    <div class="space-y-3">
      <div
        class="flex items-center justify-between text-xs text-zinc-500 dark:text-zinc-400 px-1"
      >
        <span>Attached files ({totalFiles})</span>
        <button
          type="button"
          class="text-xs text-zinc-500 hover:text-red-500 dark:hover:text-red-400 transition-colors flex items-center gap-1"
          onclick={clearAll}
        >
          <Trash2 class="w-3.5 h-3.5" />
          Clear all
        </button>
      </div>

      <!-- Files List -->
      <div class="space-y-2 max-h-[320px] overflow-y-auto pr-1">
        {#each fileList as item (item.id)}
          <div
            class="group relative flex items-center gap-3 p-3 bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-lg text-left shadow-sm transition-all hover:border-zinc-300 dark:hover:border-zinc-700"
          >
            <!-- File Icon / Thumbnail -->
            <div
              class="w-10 h-10 rounded-md overflow-hidden bg-zinc-100 dark:bg-zinc-800 shrink-0 flex items-center justify-center border border-zinc-200/50 dark:border-zinc-700/50"
            >
              {#if item.preview}
                <img
                  src={item.preview}
                  alt={item.name}
                  class="w-full h-full object-cover"
                />
              {:else if item.type.includes("pdf")}
                <FileText class="w-5 h-5 text-red-500" />
              {:else if item.type.includes("image")}
                <ImageIcon class="w-5 h-5 text-blue-500" />
              {:else}
                <FileIcon class="w-5 h-5 text-zinc-500" />
              {/if}
            </div>

            <!-- Details & Progress -->
            <div class="flex-1 min-w-0 space-y-1">
              <div class="flex items-center justify-between gap-2">
                <p
                  class="text-xs font-medium text-zinc-900 dark:text-zinc-100 truncate"
                >
                  {item.name}
                </p>
                <span
                  class="text-[10px] text-zinc-400 dark:text-zinc-500 shrink-0"
                >
                  {item.size}
                </span>
              </div>

              {#if item.status === "uploading"}
                <div
                  class="w-full h-1 bg-zinc-100 dark:bg-zinc-800 rounded-full overflow-hidden"
                >
                  <div
                    class="h-full bg-primary transition-all duration-200"
                    style="width: {item.progress}%"
                  ></div>
                </div>
              {:else if item.status === "complete"}
                <div
                  class="flex items-center gap-1 text-[11px] text-emerald-600 dark:text-emerald-400"
                >
                  <Check class="w-3 h-3" />
                  <span>Ready</span>
                </div>
              {/if}
            </div>

            <!-- Remove Action -->
            <button
              type="button"
              class="p-1 text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-200 hover:bg-zinc-100 dark:hover:bg-zinc-800 rounded transition-colors"
              onclick={() => removeFile(item.id)}
            >
              <X class="w-4 h-4" />
            </button>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
