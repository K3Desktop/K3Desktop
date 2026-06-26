<script lang="ts">
  import { onDestroy } from "svelte";
  import { ClusterService } from "../../bindings/github.com/k3desktop/k3desktop/service";
  import { LoadBalancerDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";
  import { operations, dismiss as dismissOp } from "./operationsStore";
  import type { OperationState } from "./operationsStore";
  import { clearOpLog } from "./logStore";
  import OpLog from "./OpLog.svelte";
  import ConfirmDialog from "./ConfirmDialog.svelte";

  let { clusterName, clusterStatus }: { clusterName: string; clusterStatus: string } = $props();

  let lb: LoadBalancerDTO | null = $state(null);
  let loadError = $state("");
  let open = $state(false);
  let confirmStopOpen = $state(false);

  let op = $state<OperationState | undefined>(undefined);
  const unsubOps = operations.subscribe((m) => {
    let found: OperationState | undefined;
    for (const v of m.values()) {
      if (v.target === clusterName && v.kind.startsWith("lb.")) {
        found = v;
        break;
      }
    }
    op = found;
  });
  onDestroy(unsubOps);

  const busy = $derived(op?.phase === "start");
  const errOp = $derived(op?.phase === "error" ? op : undefined);
  const busyLabel = $derived(
    op?.kind === "lb.start" ? "Starting…" :
    op?.kind === "lb.stop" ? "Stopping…" :
    op?.kind === "lb.restart" ? "Restarting…" : ""
  );

  async function load() {
    loadError = "";
    try {
      lb = await ClusterService.GetLoadBalancer(clusterName);
    } catch (e: any) {
      loadError = String(e);
    }
  }

  $effect(() => {
    // Refetch when cluster name/status changes, or when an LB op finishes.
    const _ = clusterName + clusterStatus + (op?.phase ?? "");
    if (!op || op.phase !== "start") load();
  });

  async function start() {
    clearOpLog(clusterName);
    try { await ClusterService.StartLoadBalancer(clusterName); }
    catch (e: any) { loadError = String(e); }
  }

  function stop() {
    confirmStopOpen = true;
  }

  async function performStop() {
    clearOpLog(clusterName);
    try { await ClusterService.StopLoadBalancer(clusterName); }
    catch (e: any) { loadError = String(e); }
  }

  async function restart() {
    clearOpLog(clusterName);
    try { await ClusterService.RestartLoadBalancer(clusterName); }
    catch (e: any) { loadError = String(e); }
  }
</script>

<ConfirmDialog
  bind:open={confirmStopOpen}
  title="Stop load balancer"
  message={"Stopping the loadbalancer will break kubectl access to this cluster (it fronts the API server). The cluster itself keeps running.\n\nStop the loadbalancer?"}
  confirmLabel="Stop"
  onconfirm={performStop}
/>

{#if lb}
  <div class="mt-3 rounded-lg border border-gray-200 dark:border-gray-700">
    <button
      type="button"
      onclick={() => (open = !open)}
      class="w-full flex items-center gap-2 px-3 py-2 text-left text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
    >
      <svg class="w-3.5 h-3.5 text-gray-400 transition-transform" style:transform={open ? "rotate(90deg)" : ""} viewBox="0 0 20 20" fill="none">
        <path d="M7 5l6 5-6 5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
      <span class="font-medium">Load balancer</span>
      <span class="px-2 py-0.5 rounded-full text-xs font-medium {lb.state === 'running' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}">
        {lb.state}
      </span>
      {#if busy}
        <span class="text-xs text-gray-400 dark:text-gray-500 animate-pulse">{busyLabel}</span>
      {/if}
    </button>
    {#if open}
      <div class="px-3 pb-3 pt-1 text-sm space-y-2">
        <div class="text-xs text-gray-500 dark:text-gray-400">
          <span class="font-medium">Name:</span>
          <span class="font-mono">{lb.name}</span>
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400">
          <span class="font-medium">Image:</span>
          <span class="font-mono break-all">{lb.image}</span>
        </div>
        {#if lb.ports.length > 0}
          <div class="text-xs text-gray-500 dark:text-gray-400">
            <span class="font-medium">Ports:</span>
            <ul class="mt-1 ml-3 font-mono space-y-0.5">
              {#each lb.ports as p}
                <li>{p}</li>
              {/each}
            </ul>
          </div>
        {/if}
        <div class="flex items-center gap-2 pt-1">
          {#if lb.state === "running"}
            <button
              onclick={stop}
              disabled={busy}
              class="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >Stop</button>
          {:else}
            <button
              onclick={start}
              disabled={busy || clusterStatus === "stopped"}
              title={clusterStatus === "stopped" ? "Start the cluster first" : ""}
              class="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >Start</button>
          {/if}
          <button
            onclick={restart}
            disabled={busy}
            class="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >Restart</button>
        </div>
        {#if errOp}
          <div class="p-2 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 text-xs flex items-start justify-between gap-3">
            <div>{errOp.error ?? "operation failed"}</div>
            <button onclick={() => dismissOp(errOp.kind, errOp.target)} class="hover:underline shrink-0">Dismiss</button>
          </div>
        {/if}
        {#if loadError}
          <div class="p-2 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 text-xs">{loadError}</div>
        {/if}
      </div>
    {/if}
  </div>
{/if}
