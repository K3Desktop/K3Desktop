<script lang="ts">
  import { opLogs } from "./logStore";

  let {
    active = false,
    target = "",
    onclose,
  }: { active: boolean; target?: string; onclose?: () => void } = $props();

  // Derive lines directly from the store — no local state, no effects, no subscriptions.
  const lines = $derived(target ? ($opLogs.get(target) ?? []) : []);

  function levelClass(level: string) {
    switch (level) {
      case "ERROR": return "text-red-600 dark:text-red-400";
      case "WARN":  return "text-yellow-600 dark:text-yellow-300";
      case "DEBUG": return "text-gray-400 dark:text-gray-500";
      default:      return "text-gray-600 dark:text-gray-300";
    }
  }

  // Auto-scroll: run whenever lines changes and we're mounted.
  let scrollEl: HTMLDivElement | undefined = $state();
  $effect(() => {
    const _ = lines.length; // track length changes
    if (scrollEl) scrollEl.scrollTop = scrollEl.scrollHeight;
  });
</script>

{#if active || lines.length > 0}
  <div class="my-2 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
    {#if !active && lines.length > 0 && onclose}
      <div class="flex items-center justify-between px-3 py-1 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
        <span class="text-xs text-gray-400 dark:text-gray-500 font-medium">Logs</span>
        <button
          onclick={onclose}
          class="text-gray-400 dark:text-gray-500 hover:text-gray-700 dark:hover:text-gray-200 text-sm leading-none transition-colors"
        >✕</button>
      </div>
    {/if}
    <div
      bind:this={scrollEl}
      class="bg-gray-100 dark:bg-gray-900 overflow-y-auto font-mono text-xs p-3 space-y-0.5"
      style="max-height: 160px"
    >
      {#each lines as entry (entry.time + entry.message)}
        <div class="flex gap-2 leading-5">
          <span class="shrink-0 w-12 font-medium {levelClass(entry.level)}">{entry.level}</span>
          <span class="shrink-0 w-8 text-gray-500 dark:text-gray-600">{entry.source === "k3d" ? "k3d" : "app"}</span>
          <span class="text-gray-800 dark:text-gray-200 break-all">{entry.message}</span>
        </div>
      {/each}
      {#if active && lines.length === 0}
        <p class="text-gray-400 dark:text-gray-600 animate-pulse">Waiting for logs…</p>
      {/if}
    </div>
  </div>
{/if}
