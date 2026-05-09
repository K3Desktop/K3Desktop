<script lang="ts">
  import { onMount, tick } from "svelte";
  import { logEntries, clearLogs } from "./logStore";
  import type { LogEntryDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";

  let {
    open = $bindable(false),
    sidebarWidth = 224,
  }: { open: boolean; sidebarWidth?: number } = $props();

  let scrollEl: HTMLDivElement | undefined = $state();
  let autoScroll = $state(true);

  function levelClass(level: string) {
    switch (level) {
      case "ERROR": return "text-red-500 dark:text-red-400";
      case "WARN":  return "text-yellow-600 dark:text-yellow-400";
      case "DEBUG": return "text-gray-400 dark:text-gray-500";
      default:      return "text-gray-700 dark:text-gray-300";
    }
  }

  function formatTime(iso: string) {
    try { return new Date(iso).toLocaleTimeString(); } catch { return iso; }
  }

  // Auto-scroll when new entries arrive, if user hasn't scrolled up.
  $effect(() => {
    // subscribe to logEntries to re-run on change
    const _ = $logEntries;
    if (autoScroll && scrollEl) {
      tick().then(() => {
        if (scrollEl) scrollEl.scrollTop = scrollEl.scrollHeight;
      });
    }
  });

  function onScroll() {
    if (!scrollEl) return;
    const { scrollTop, scrollHeight, clientHeight } = scrollEl;
    autoScroll = scrollHeight - scrollTop - clientHeight < 40;
  }
</script>

{#if open}
  <div class="fixed bottom-0 right-0 z-40 flex flex-col
              bg-white dark:bg-gray-900
              border-t border-gray-200 dark:border-gray-700
              shadow-2xl"
       style="height: 280px; left: {sidebarWidth}px">
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-700 shrink-0">
      <span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Logs</span>
      <div class="flex items-center gap-3">
        <span class="text-xs text-gray-400 dark:text-gray-500">{$logEntries.length} entries</span>
        <button onclick={clearLogs} class="text-xs text-gray-400 dark:text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 transition-colors">Clear</button>
        <button onclick={() => (open = false)} class="text-gray-400 dark:text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 transition-colors text-sm leading-none">✕</button>
      </div>
    </div>

    <!-- Log lines -->
    <div
      bind:this={scrollEl}
      onscroll={onScroll}
      class="flex-1 overflow-y-auto font-mono text-xs px-4 py-2 space-y-0.5"
    >
      {#each $logEntries as entry}
        <div class="flex gap-3 leading-5 hover:bg-gray-100 dark:hover:bg-white/5 rounded px-1">
          <span class="text-gray-400 dark:text-gray-600 shrink-0 w-20">{formatTime(entry.time)}</span>
          <span class="shrink-0 w-12 {levelClass(entry.level)} font-medium">{entry.level}</span>
          <span class="shrink-0 w-8 text-gray-400 dark:text-gray-600">{entry.source === "k3d" ? "k3d" : "app"}</span>
          <span class="text-gray-800 dark:text-gray-200 break-all">{entry.message}</span>
        </div>
      {/each}
      {#if $logEntries.length === 0}
        <p class="text-gray-400 dark:text-gray-600 italic">No logs yet.</p>
      {/if}
    </div>

    <!-- Scroll-to-bottom hint -->
    {#if !autoScroll}
      <button
        onclick={() => { autoScroll = true; if (scrollEl) scrollEl.scrollTop = scrollEl.scrollHeight; }}
        class="absolute bottom-2 right-4 text-xs bg-brand text-gray-900 px-2 py-1 rounded shadow hover:bg-brand-dim transition-colors font-semibold"
      >
        ↓ latest
      </button>
    {/if}
  </div>
{/if}
