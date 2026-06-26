<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { RegistryService } from "../../bindings/github.com/k3desktop/k3desktop/service";
  import { RegistryDTO, RegistryCreateRequest } from "../../bindings/github.com/k3desktop/k3desktop/dto";
  import { operations, dismiss as dismissOp } from "../lib/operationsStore";
  import type { OperationState } from "../lib/operationsStore";
  import ErrorAlert from "../lib/ErrorAlert.svelte";
  import ConfirmDialog from "../lib/ConfirmDialog.svelte";

  let registries: RegistryDTO[] = $state([]);
  let loading = $state(false);
  let error = $state("");
  let showCreate = $state(false);
  let form = $state({ name: "", port: 0 });
  let confirmOpen = $state(false);
  let pendingRegistry = $state<string | null>(null);

  let registryOps: Map<string, OperationState> = $state(new Map());
  const unsubOps = operations.subscribe((m) => {
    const next = new Map<string, OperationState>();
    for (const op of m.values()) {
      if (op.kind.startsWith("registry.")) next.set(op.target, op);
    }
    registryOps = next;
  });
  onDestroy(unsubOps);

  function opFor(name: string): OperationState | undefined {
    return registryOps.get(name);
  }

  async function load() {
    loading = true;
    error = "";
    try {
      registries = await RegistryService.ListRegistries();
    } catch (e: any) {
      error = String(e);
    } finally {
      loading = false;
    }
  }

  async function create() {
    if (!form.name) return;
    try {
      await RegistryService.CreateRegistry(new RegistryCreateRequest(form));
      showCreate = false;
      form = { name: "", port: 0 };
    } catch (e: any) {
      error = String(e);
    }
  }

  function del(name: string) {
    pendingRegistry = name;
    confirmOpen = true;
  }

  async function performDeleteRegistry() {
    const name = pendingRegistry;
    pendingRegistry = null;
    if (!name) return;
    try {
      await RegistryService.DeleteRegistry(name);
    } catch (e: any) {
      error = String(e);
    }
  }

  // Refresh whenever any registry op transitions out of "start".
  let lastSeenIds = new Set<string>();
  const unsubRefresh = operations.subscribe((m) => {
    const stillActive = new Set<string>();
    for (const op of m.values()) {
      if (op.kind.startsWith("registry.")) stillActive.add(op.id);
    }
    let shouldReload = false;
    for (const id of lastSeenIds) if (!stillActive.has(id)) shouldReload = true;
    for (const op of m.values()) {
      if (op.kind.startsWith("registry.") && (op.phase === "done" || op.phase === "error") && !lastSeenIds.has(op.id)) shouldReload = true;
    }
    lastSeenIds = stillActive;
    if (shouldReload) load();
  });
  onDestroy(unsubRefresh);

  onMount(load);
</script>

<div class="p-6">
  <div class="flex items-center justify-between mb-6">
    <h4 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">Registries</h4>
    <button onclick={() => (showCreate = true)} class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim transition-colors">
      + New registry
    </button>
  </div>

  <ErrorAlert bind:message={error} />

  {#each Array.from(registryOps.values()).filter(o => o.kind === "registry.create" && !registries.some(r => r.name === o.target)) as op (op.id)}
    {#if op.phase === "start"}
      <div class="mb-2 p-3 rounded-lg bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 text-sm animate-pulse">Creating registry <span class="font-medium">{op.target}</span>…</div>
    {:else if op.phase === "error"}
      <div class="mb-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 text-sm flex items-start justify-between gap-3">
        <div>Creating <span class="font-medium">{op.target}</span> failed — {op.error}</div>
        <button onclick={() => dismissOp(op.kind, op.target)} class="hover:underline shrink-0">Dismiss</button>
      </div>
    {/if}
  {/each}

  {#if loading}
    <div class="text-sm text-gray-400 dark:text-gray-500">Loading…</div>
  {:else if registries.length === 0}
    <div class="text-sm text-gray-400 dark:text-gray-500">No registries. Create one to get started.</div>
  {:else}
    <div class="grid gap-3">
      {#each registries as r}
        {@const op = opFor(r.name)}
        {@const busy = op?.phase === "start"}
        {@const errOp = op?.phase === "error" ? op : undefined}
        <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4">
          <div class="flex items-center gap-4">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span class="font-medium text-gray-900 dark:text-gray-100">{r.name}</span>
                <span class="px-2 py-0.5 rounded-full text-xs font-medium {r.state === 'running' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}">
                  {r.state}
                </span>
                {#if busy}
                  <span class="text-xs text-gray-400 dark:text-gray-500 animate-pulse">{op?.kind === "registry.delete" ? "Deleting…" : "Working…"}</span>
                {/if}
              </div>
              <div class="mt-1 font-mono text-sm text-gray-500 dark:text-gray-400">{r.protocol}://{r.host}{r.port ? `:${r.port}` : ""}</div>
            </div>
            <button onclick={() => del(r.name)} disabled={busy} class="px-3 py-1.5 rounded-lg text-xs font-medium border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 transition-colors">Delete</button>
          </div>
          {#if errOp}
            <div class="mt-3 p-2 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 text-xs flex items-start justify-between gap-3">
              <div>{errOp.error ?? "operation failed"}</div>
              <button onclick={() => dismissOp(errOp.kind, errOp.target)} class="hover:underline shrink-0">Dismiss</button>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<ConfirmDialog
  bind:open={confirmOpen}
  title="Delete registry"
  message={`Delete registry "${pendingRegistry ?? ""}"?`}
  confirmLabel="Delete"
  danger
  onconfirm={performDeleteRegistry}
  oncancel={() => { pendingRegistry = null; }}
/>

{#if showCreate}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-6 w-full max-w-sm mx-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">New registry</h2>
      <div class="space-y-3">
        <label class="block">
          <span class="text-sm text-gray-600 dark:text-gray-400">Name *</span>
          <input bind:value={form.name} class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent" placeholder="my-registry" />
        </label>
        <label class="block">
          <span class="text-sm text-gray-600 dark:text-gray-400">Host port <span class="text-gray-400">(0 = random)</span></span>
          <input type="number" bind:value={form.port} min="0" max="65535" class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent" />
        </label>
      </div>
      <div class="mt-5 flex justify-end gap-3">
        <button onclick={() => (showCreate = false)} class="px-4 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">Cancel</button>
        <button onclick={create} disabled={!form.name} class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 transition-colors">Create</button>
      </div>
    </div>
  </div>
{/if}
