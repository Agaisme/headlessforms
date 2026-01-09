<script lang="ts">
  import { onMount, onDestroy } from "svelte";

  export let isOpen = false;
  export let title = "";
  export let size: "sm" | "md" | "lg" | "xl" | "2xl" = "md";
  export let closeOnBackdrop = true;
  export let closeOnEscape = true;
  export let onclose: (() => void) | null = null;

  let modalElement: HTMLDivElement | null = null;
  let contentElement: HTMLDivElement | null = null;

  const sizeClasses: Record<string, string> = {
    sm: "max-w-sm",
    md: "max-w-md",
    lg: "max-w-lg",
    xl: "max-w-xl",
    "2xl": "max-w-2xl",
  };

  function close() {
    isOpen = false;
    if (onclose) {
      onclose();
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if (closeOnBackdrop && e.target === e.currentTarget) {
      close();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (closeOnEscape && e.key === "Escape" && isOpen) {
      close();
    }
  }

  onMount(() => {
    document.addEventListener("keydown", handleKeydown);
  });

  onDestroy(() => {
    document.removeEventListener("keydown", handleKeydown);
  });

  $: if (isOpen && modalElement) {
    document.body.style.overflow = "hidden";
  } else {
    document.body.style.overflow = "";
  }
</script>

{#if isOpen}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/60 backdrop-blur-sm sm:p-4"
    onclick={handleBackdropClick}
    bind:this={modalElement}
  >
    <div
      bind:this={contentElement}
      class="bg-background border border-border rounded-t-2xl sm:rounded-xl shadow-2xl w-full sm:w-auto sm:{sizeClasses[
        size
      ] ||
        sizeClasses.md} max-h-[90vh] sm:max-h-[85vh] transform transition-all duration-200 animate-in slide-in-from-bottom sm:fade-in sm:zoom-in-95"
    >
      <!-- Header -->
      <div
        class="flex items-center justify-between px-6 py-4 border-b border-border"
      >
        <h3 class="text-lg font-semibold text-foreground">{title}</h3>
        <button
          type="button"
          class="p-2 -m-2 text-muted-foreground hover:text-foreground hover:bg-muted rounded-lg transition-colors"
          onclick={close}
          aria-label="Close"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
              clip-rule="evenodd"
            />
          </svg>
        </button>
      </div>

      <!-- Body -->
      <div class="px-6 py-5 max-h-[70vh] overflow-y-auto">
        <slot />
      </div>

      <!-- Footer (optional) -->
      {#if $$slots.footer}
        <div
          class="flex justify-end gap-3 px-6 py-4 border-t border-border bg-muted/30"
        >
          <slot name="footer" />
        </div>
      {/if}
    </div>
  </div>
{/if}
