<script lang="ts">
  let { message = $bindable(), onclose }: { message: string; onclose?: () => void } = $props();

  function dismiss() {
    if (onclose) onclose();
    else message = "";
  }
</script>

{#if message}
  <div class="mb-4 flex items-start gap-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 text-sm">
    <span class="flex-1 min-w-0 break-words">{message}</span>
    <button
      onclick={dismiss}
      class="shrink-0 leading-none hover:opacity-70 transition-opacity"
      aria-label="Dismiss error"
    >✕</button>
  </div>
{/if}
