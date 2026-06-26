<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { ClusterService, VersionService, ProfileService } from "../../bindings/github.com/k3desktop/k3desktop/service";
  import { ClusterCreateRequest, ClusterCreateAdvancedRequest, NodeFilter, UlimitDTO, FileDTO, HostAliasDTO, ClusterDTO, ProfileDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";
  import ClusterForm from "../lib/ClusterForm.svelte";
  import OpLog from "../lib/OpLog.svelte";
  import LoadBalancerPanel from "../lib/LoadBalancerPanel.svelte";
  import { clusterFormPrefill, defaultAdv, advToDto, dtoToAdv } from "../lib/prefill";
  import { clearOpLog } from "../lib/logStore";
  import { operations, dismiss as dismissOp } from "../lib/operationsStore";
  import type { OperationState } from "../lib/operationsStore";
  import type { AdvState } from "../lib/prefill";
  import ErrorAlert from "../lib/ErrorAlert.svelte";

  let clusters: ClusterDTO[] = $state([]);
  let loading = $state(false);
  let error = $state("");
  let showCreate = $state(false);
  let showAdvanced = $state(false);
  let visibleLogs: Record<string, boolean> = $state({}); // names with log panel open

  // Derived from the global store — survives navigation.
  let clusterOps: Map<string, OperationState> = $state(new Map());
  let creatingOps: OperationState[] = $state([]);
  const unsubOps = operations.subscribe((m) => {
    const next = new Map<string, OperationState>();
    const creates: OperationState[] = [];
    for (const op of m.values()) {
      if (!op.kind.startsWith("cluster.")) continue;
      next.set(op.target, op);
      if (op.kind === "cluster.create") creates.push(op);
    }
    clusterOps = next;
    creatingOps = creates;
  });
  onDestroy(unsubOps);

  function opFor(name: string): OperationState | undefined {
    return clusterOps.get(name);
  }
  function isBusy(name: string): boolean {
    return opFor(name)?.phase === "start";
  }
  function errorFor(name: string): OperationState | undefined {
    const op = opFor(name);
    return op?.phase === "error" ? op : undefined;
  }

  let k3sVersions: string[] = $state([]);
  let versionsLoading = $state(false);

  let profiles: ProfileDTO[] = $state([]);
  let profileDropdownOpen = $state(false);

  async function loadProfiles() {
    try {
      profiles = await ProfileService.ListProfiles() ?? [];
    } catch { /* non-fatal */ }
  }

  async function createFromProfile(p: ProfileDTO) {
    profileDropdownOpen = false;
    error = "";
    try {
      const yaml = await ProfileService.GetProfile(p.name);
      const req = await ProfileService.YAMLToAdvancedRequest(yaml);
      adv = dtoToAdv(req);
      loadVersions();
      showAdvanced = true;
    } catch (e: any) {
      error = String(e);
    }
  }

  async function loadVersions() {
    if (k3sVersions.length > 0) return;
    versionsLoading = true;
    try {
      k3sVersions = await VersionService.ListK3sVersions(20);
    } catch { /* non-fatal */ } finally {
      versionsLoading = false;
    }
  }

  // Simple form
  let simple = $state({ name: "", servers: 1, agents: 0, image: "", apiPort: "" });

  // Advanced form
  let adv: AdvState = $state(defaultAdv());

  async function load() {
    loading = true;
    error = "";
    try {
      clusters = await ClusterService.ListClusters();
    } catch (e: any) {
      error = String(e);
    } finally {
      loading = false;
    }
  }

  async function silentLoad() {
    try {
      clusters = await ClusterService.ListClusters();
    } catch (e: any) {
      error = String(e);
    }
  }

  async function createSimple() {
    if (!simple.name) return;
    clearOpLog(simple.name);
    visibleLogs[simple.name] = true;
    showCreate = false;
    try {
      await ClusterService.CreateCluster(new ClusterCreateRequest(simple));
      simple = { name: "", servers: 1, agents: 0, image: "", apiPort: "" };
    } catch (e: any) {
      error = String(e);
    }
  }

  async function createAdvanced() {
    if (!adv.name) return;
    clearOpLog(adv.name);
    visibleLogs[adv.name] = true;
    showAdvanced = false;
    try {
      const dto = advToDto(adv);
      await ClusterService.CreateClusterAdvanced(new ClusterCreateAdvancedRequest(dto));
    } catch (e: any) {
      error = String(e);
    }
  }

  async function del(name: string) {
    if (!confirm(`Delete cluster "${name}"?`)) return;
    visibleLogs[name] = true;
    clearOpLog(name);
    try {
      await ClusterService.DeleteCluster(name);
    } catch (e: any) {
      error = String(e);
    }
  }

  async function toggle(cluster: ClusterDTO) {
    visibleLogs[cluster.name] = true;
    clearOpLog(cluster.name);
    try {
      if (cluster.status === "running") {
        await ClusterService.StopCluster(cluster.name);
      } else {
        await ClusterService.StartCluster(cluster.name);
      }
    } catch (e: any) {
      error = String(e);
    }
  }

  // Whenever any cluster operation finishes, refresh the list. Survives navigation.
  let lastSeenIds = new Set<string>();
  const unsubRefresh = operations.subscribe((m) => {
    const stillActive = new Set<string>();
    for (const op of m.values()) {
      if (op.kind.startsWith("cluster.")) stillActive.add(op.id);
    }
    // Detect ops that disappeared or transitioned out of "start" since last tick.
    let shouldReload = false;
    for (const id of lastSeenIds) {
      if (!stillActive.has(id)) shouldReload = true;
    }
    for (const op of m.values()) {
      if (op.kind.startsWith("cluster.") && (op.phase === "done" || op.phase === "error") && !lastSeenIds.has(op.id)) {
        shouldReload = true;
      }
    }
    lastSeenIds = stillActive;
    if (shouldReload) silentLoad();
  });
  onDestroy(unsubRefresh);

  onMount(() => {
    load();
    loadProfiles();

    // Check for profile-prefilled form state
    const unsub = clusterFormPrefill.subscribe((prefill) => {
      if (prefill) {
        adv = prefill;
        clusterFormPrefill.set(null);
        loadVersions();
        showAdvanced = true;
      }
    });
    return unsub;
  });
</script>

<div class="p-6">
  <div class="flex items-center justify-between mb-6">
    <h4 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">Clusters</h4>
    <div class="flex items-center gap-2">
      {#if profiles.length > 0}
        <div class="relative">
          <button
            onclick={() => profileDropdownOpen = !profileDropdownOpen}
            class="px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors flex items-center gap-1.5"
          >
            From profile
            <svg class="w-3.5 h-3.5 text-gray-400" viewBox="0 0 20 20" fill="none">
              <path d="M5 8l5 5 5-5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
          {#if profileDropdownOpen}
            <div class="absolute right-0 z-10 mt-1 min-w-40 rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 shadow-lg overflow-hidden">
              {#each profiles as p}
                <button
                  type="button"
                  onclick={() => createFromProfile(p)}
                  class="w-full text-left px-4 py-2 text-sm text-gray-900 dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                >
                  {p.name}
                </button>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
      <button
        onclick={() => (showCreate = true)}
        class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim transition-colors"
      >
        + New cluster
      </button>
    </div>
  </div>

  <ErrorAlert bind:message={error} />

  {#each creatingOps as op (op.id)}
    {#if op.phase === "start"}
      <div class="mb-2 p-3 rounded-lg bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 text-sm animate-pulse">Creating cluster <span class="font-medium">{op.target}</span>…</div>
      <OpLog active={true} target={op.target} onclose={() => {}} />
    {:else if op.phase === "error"}
      <div class="mb-2 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 text-sm flex items-start justify-between gap-3">
        <div><span class="font-medium">{op.target}</span> — {op.error ?? "creation failed"}</div>
        <button onclick={() => dismissOp(op.kind, op.target)} class="text-red-700 dark:text-red-300 hover:underline shrink-0">Dismiss</button>
      </div>
    {/if}
  {/each}

  {#if loading}
    <div class="text-sm text-gray-400 dark:text-gray-500">Loading…</div>
  {:else if clusters.length === 0}
    <div class="text-sm text-gray-400 dark:text-gray-500">No clusters. Create one to get started.</div>
  {:else}
    <div class="grid gap-4">
      {#each clusters as c (c.name)}
        {@const op = opFor(c.name)}
        {@const busy = op?.phase === "start"}
        {@const errOp = op?.phase === "error" ? op : undefined}
        {@const busyLabel = busy ? (op?.kind === "cluster.stop" ? "Stopping…" : op?.kind === "cluster.delete" ? "Deleting…" : op?.kind === "cluster.start" ? "Starting…" : "Working…") : ""}
        <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5">
          <div class="flex items-center gap-4">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span class="font-medium text-gray-900 dark:text-gray-100">{c.name}</span>
                <span class="px-2 py-0.5 rounded-full text-xs font-medium {c.status === 'running' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : c.status === 'partial' ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}">
                  {c.status}
                </span>
                {#if busy}
                  <span class="text-xs text-gray-400 dark:text-gray-500 animate-pulse">{busyLabel}</span>
                {/if}
              </div>
              <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {c.servers} server{c.servers !== 1 ? "s" : ""} · {c.agents} agent{c.agents !== 1 ? "s" : ""}
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button
                onclick={() => toggle(c)}
                disabled={busy}
                class="px-3 py-1.5 rounded-lg text-xs font-medium border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                {c.status === "running" ? "Stop" : "Start"}
              </button>
              <button
                onclick={() => del(c.name)}
                disabled={busy}
                class="px-3 py-1.5 rounded-lg text-xs font-medium border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              >
                Delete
              </button>
            </div>
          </div>
          {#if errOp}
            <div class="mt-3 p-2 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-300 text-xs flex items-start justify-between gap-3">
              <div>{errOp.error ?? "operation failed"}</div>
              <button onclick={() => dismissOp(errOp.kind, errOp.target)} class="hover:underline shrink-0">Dismiss</button>
            </div>
          {/if}
          {#if visibleLogs[c.name]}
            <OpLog active={busy} target={c.name} onclose={() => { delete visibleLogs[c.name]; }} />
          {/if}
          <LoadBalancerPanel clusterName={c.name} clusterStatus={c.status} />
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Simple create modal -->
{#if showCreate}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-6 w-full max-w-md mx-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">New cluster</h2>
      <div class="space-y-3">
        <label class="block">
          <span class="text-sm text-gray-600 dark:text-gray-400">Name *</span>
          <input bind:value={simple.name} class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent" placeholder="my-cluster" />
        </label>
        <div class="grid grid-cols-2 gap-3">
          <label class="block">
            <span class="text-sm text-gray-600 dark:text-gray-400">Servers</span>
            <input type="number" bind:value={simple.servers} min="1" max="5" class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent" />
          </label>
          <label class="block">
            <span class="text-sm text-gray-600 dark:text-gray-400">Agents</span>
            <input type="number" bind:value={simple.agents} min="0" max="10" class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent" />
          </label>
        </div>
        <label class="block">
          <span class="text-sm text-gray-600 dark:text-gray-400">Image <span class="text-gray-400">(optional)</span></span>
          <input bind:value={simple.image} onfocus={loadVersions} list="k3s-versions-simple"
            class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
            placeholder={versionsLoading ? "Loading versions…" : "docker.io/rancher/k3s:latest"} />
          <datalist id="k3s-versions-simple">
            {#each k3sVersions as v}
              <option value="docker.io/rancher/k3s:{v}">{v}</option>
            {/each}
          </datalist>
        </label>
        <label class="block">
          <span class="text-sm text-gray-600 dark:text-gray-400">API port <span class="text-gray-400">(optional)</span></span>
          <input bind:value={simple.apiPort} class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent" placeholder="6443" />
        </label>
      </div>
      <div class="mt-5 flex items-center justify-between">
        <button
          onclick={() => { showCreate = false; adv = { ...defaultAdv(), name: simple.name, servers: simple.servers, agents: simple.agents, image: simple.image, apiPort: simple.apiPort }; loadVersions(); showAdvanced = true; }}
          class="text-sm text-accent hover:underline"
        >Advanced options →</button>
        <div class="flex gap-3">
          <button onclick={() => (showCreate = false)} class="px-4 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors">Cancel</button>
          <button onclick={createSimple} disabled={!simple.name} class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 transition-colors">Create</button>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Advanced create page (full overlay) -->
{#if showAdvanced}
  <div class="fixed inset-0 bg-gray-50 dark:bg-gray-950 z-50 overflow-y-auto">
    <div class="max-w-2xl mx-auto px-6 py-8">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">New cluster — advanced</h2>
        <button onclick={() => (showAdvanced = false)} class="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">✕ Cancel</button>
      </div>

      <ClusterForm bind:adv {k3sVersions} {versionsLoading} datalistId="k3s-versions-adv" />

      <div class="flex justify-end gap-3 border-t border-gray-200 dark:border-gray-700 pt-6">
        <button onclick={() => (showAdvanced = false)} class="px-4 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">Cancel</button>
        <button onclick={createAdvanced} disabled={!adv.name} class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 transition-colors">Create cluster</button>
      </div>
    </div>
  </div>
{/if}
