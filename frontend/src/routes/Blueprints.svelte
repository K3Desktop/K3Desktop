<script lang="ts">
  import { onMount, tick } from "svelte";
  import { Events } from "@wailsio/runtime";
  import { BlueprintService, ClusterService } from "../../bindings/github.com/k3desktop/k3desktop/service";
  import { BlueprintDTO, BlueprintDeployRequest, ChartEntryDTO, ClusterDTO } from "../../bindings/github.com/k3desktop/k3desktop/dto";
  import ErrorAlert from "../lib/ErrorAlert.svelte";
  import OpLog from "../lib/OpLog.svelte";
  import YamlEditor from "../lib/YamlEditor.svelte";
  import { clearOpLog } from "../lib/logStore";

  let blueprints: BlueprintDTO[] = $state([]);
  let loading = $state(false);
  let error = $state("");
  let busy: Record<string, boolean> = $state({});
  let visibleLogs: Record<string, boolean> = $state({});
  let deploying: Record<string, boolean> = $state({});

  // Editor overlay state
  let showForm = $state(false);
  let editingName = $state("");
  let formName = $state("");
  let formDescription = $state("");
  let formCharts: ChartEntryDTO[] = $state([]);
  let saving = $state(false);

  // Deploy modal state
  let showDeploy = $state(false);
  let deployTarget = $state("");
  let clusters: ClusterDTO[] = $state([]);
  let deployCluster = $state("");
  let deployNamespace = $state("default");
  let deploying_ = $state(false);

  async function load() {
    loading = true;
    error = "";
    try {
      blueprints = await BlueprintService.ListBlueprints() ?? [];
    } catch (e: any) {
      error = String(e);
      blueprints = [];
    } finally {
      loading = false;
    }
  }

  function openNew() {
    editingName = "";
    formName = "";
    formDescription = "";
    formCharts = [new ChartEntryDTO({ releaseName: "", repo: "", chart: "", version: "", values: "" })];
    showForm = true;
  }

  async function openEdit(bp: BlueprintDTO) {
    error = "";
    try {
      const loaded = await BlueprintService.GetBlueprint(bp.name);
      editingName = loaded.name;
      formName = loaded.name;
      formDescription = loaded.description;
      formCharts = loaded.charts?.map(c => new ChartEntryDTO(c)) ?? [];
      if (formCharts.length === 0) {
        formCharts = [new ChartEntryDTO({ releaseName: "", repo: "", chart: "", version: "", values: "" })];
      }
      showForm = true;
    } catch (e: any) {
      error = String(e);
    }
  }

  function addChart() {
    formCharts.push(new ChartEntryDTO({ releaseName: "", repo: "", chart: "", version: "", values: "" }));
  }

  function removeChart(i: number) {
    formCharts.splice(i, 1);
  }

  async function saveBlueprint() {
    if (!formName.trim()) { error = "Blueprint name required."; return; }
    if (formCharts.length === 0) { error = "Add at least one chart."; return; }
    saving = true;
    error = "";
    try {
      const bp = new BlueprintDTO({
        name: formName.trim(),
        description: formDescription,
        fileName: "",
        charts: formCharts.map(c => new ChartEntryDTO(c)),
      });
      await BlueprintService.SaveBlueprint(bp);
      showForm = false;
    } catch (e: any) {
      error = String(e);
    } finally {
      saving = false;
    }
    if (!showForm) {
      await tick();
      load();
    }
  }

  async function deleteBlueprint(bp: BlueprintDTO) {
    if (!confirm(`Delete blueprint "${bp.name}"?`)) return;
    busy[bp.name] = true;
    error = "";
    try {
      await BlueprintService.DeleteBlueprint(bp.name);
      await load();
    } catch (e: any) {
      error = String(e);
    } finally {
      delete busy[bp.name];
    }
  }

  async function openDeploy(bp: BlueprintDTO) {
    deployTarget = bp.name;
    deployNamespace = "default";
    error = "";
    try {
      clusters = await ClusterService.ListClusters() ?? [];
      deployCluster = clusters.length > 0 ? clusters[0].name : "";
    } catch (e: any) {
      error = String(e);
      return;
    }
    showDeploy = true;
  }

  async function deployBlueprint() {
    if (!deployCluster) { error = "Select a cluster."; return; }
    deploying_ = true;
    error = "";
    clearOpLog(deployTarget);
    try {
      const req = new BlueprintDeployRequest({
        blueprintName: deployTarget,
        clusterName: deployCluster,
        namespace: deployNamespace || "default",
      });
      await BlueprintService.DeployBlueprint(req);
      showDeploy = false;
    } catch (e: any) {
      error = String(e);
    } finally {
      deploying_ = false;
    }
  }

  onMount(() => {
    load();

    Events.On("blueprint:deploying", (ev: any) => {
      const name: string = ev?.data?.name ?? ev?.data ?? "";
      deploying[name] = true;
      visibleLogs[name] = true;
    });

    Events.On("blueprint:done", (ev: any) => {
      const name: string = ev?.data?.name ?? ev?.data ?? "";
      deploying[name] = false;
    });

    Events.On("blueprint:error", (ev: any) => {
      const name: string = ev?.data?.name ?? ev?.data ?? "";
      const msg: string = ev?.data?.message ?? "";
      deploying[name] = false;
      if (msg) error = msg;
    });
  });
</script>

<div class="p-6">
  <div class="flex items-center justify-between mb-6">
    <h4 class="text-2xl font-semibold text-gray-900 dark:text-gray-100">Blueprints</h4>
    <button
      onclick={openNew}
      class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim transition-colors"
    >
      + New blueprint
    </button>
  </div>

  <ErrorAlert bind:message={error} />

  {#if loading}
    <div class="text-sm text-gray-400 dark:text-gray-500">Loading…</div>
  {:else if blueprints.length === 0}
    <div class="text-sm text-gray-400 dark:text-gray-500">
      No blueprints yet. Create a blueprint to bundle Helm charts with preconfigured values.
    </div>
  {:else}
    <div class="grid gap-4">
      {#each blueprints as bp}
        <div class="rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-5">
          <div class="flex items-center gap-4 flex-wrap">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <span class="font-medium text-gray-900 dark:text-gray-100">{bp.name}</span>
                <span class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-accent/10 text-accent">
                  {bp.charts?.length ?? 0} chart{(bp.charts?.length ?? 0) !== 1 ? "s" : ""}
                </span>
              </div>
              {#if bp.description}
                <p class="mt-0.5 text-sm text-gray-500 dark:text-gray-400 truncate">{bp.description}</p>
              {/if}
            </div>
            <div class="flex items-center gap-2 flex-wrap">
              <button
                onclick={() => openDeploy(bp)}
                disabled={!!busy[bp.name]}
                class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-brand text-gray-900 hover:bg-brand-dim disabled:opacity-50 transition-colors"
              >
                Deploy
              </button>
              <button
                onclick={() => { visibleLogs[bp.name] = !visibleLogs[bp.name]; }}
                class="px-3 py-1.5 rounded-lg text-xs border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
              >
                Logs
              </button>
              <button
                onclick={() => openEdit(bp)}
                class="px-3 py-1.5 rounded-lg text-xs border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
              >
                Edit
              </button>
              <button
                onclick={() => deleteBlueprint(bp)}
                disabled={!!busy[bp.name]}
                class="px-3 py-1.5 rounded-lg text-xs border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 disabled:opacity-50 transition-colors"
              >
                Delete
              </button>
            </div>
          </div>
          {#if visibleLogs[bp.name]}
            <div class="mt-3">
              <OpLog active={!!deploying[bp.name]} target={bp.name} onclose={() => { delete visibleLogs[bp.name]; }} />
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Deploy modal -->
{#if showDeploy}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md p-6">
      <h5 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Deploy blueprint — {deployTarget}</h5>

      <div class="space-y-4">
        <label class="block">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Cluster *</span>
          <select
            bind:value={deployCluster}
            class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
          >
            {#each clusters as c}
              <option value={c.name}>{c.name} ({c.status})</option>
            {/each}
          </select>
        </label>

        <label class="block">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Namespace</span>
          <input
            bind:value={deployNamespace}
            placeholder="default"
            class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
          />
        </label>
      </div>

      <div class="flex justify-end gap-3 mt-6">
        <button
          onclick={() => { showDeploy = false; }}
          class="px-4 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
        >
          Cancel
        </button>
        <button
          onclick={deployBlueprint}
          disabled={deploying_ || !deployCluster}
          class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 transition-colors"
        >
          {deploying_ ? "Starting…" : "Deploy"}
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- Blueprint editor overlay -->
{#if showForm}
  <div class="fixed inset-0 bg-gray-50 dark:bg-gray-950 z-50 overflow-y-auto">
    <div class="max-w-2xl mx-auto px-6 py-8">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100">
          {editingName ? `Edit blueprint — ${editingName}` : "New blueprint"}
        </h2>
        <button onclick={() => (showForm = false)} class="text-sm text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200">✕ Cancel</button>
      </div>

      <!-- Name & description -->
      <section class="mb-6 p-4 rounded-xl border border-accent/30 bg-accent/5 space-y-4">
        <label class="block">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Blueprint name *</span>
          <input
            bind:value={formName}
            placeholder="my-dev-stack"
            class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
          />
          <p class="mt-1 text-xs text-gray-400">Saved as <code>~/.config/k3desktop/blueprints/{formName || "…"}.yaml</code></p>
        </label>
        <label class="block">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Description</span>
          <input
            bind:value={formDescription}
            placeholder="Optional description"
            class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
          />
        </label>
      </section>

      <!-- Charts -->
      <section class="mb-6">
        <div class="flex items-center justify-between mb-3">
          <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Charts</h4>
          <button
            onclick={addChart}
            class="px-3 py-1 rounded-lg text-xs font-medium border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            + Add chart
          </button>
        </div>

        {#if formCharts.length === 0}
          <p class="text-sm text-gray-400 dark:text-gray-500">No charts yet. Add one above.</p>
        {:else}
          <div class="space-y-4">
            {#each formCharts as chart, i}
              <div class="p-4 rounded-xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800">
                <div class="flex items-center justify-between mb-3">
                  <span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Chart {i + 1}</span>
                  {#if formCharts.length > 1}
                    <button
                      onclick={() => removeChart(i)}
                      class="text-xs text-red-500 hover:text-red-700 dark:hover:text-red-300 transition-colors"
                    >✕ Remove</button>
                  {/if}
                </div>

                <div class="grid grid-cols-2 gap-3 mb-3">
                  <label class="block">
                    <span class="text-xs font-medium text-gray-600 dark:text-gray-400">Release name *</span>
                    <input
                      bind:value={chart.releaseName}
                      placeholder="nginx"
                      class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
                    />
                  </label>
                  <label class="block">
                    <span class="text-xs font-medium text-gray-600 dark:text-gray-400">Chart name *</span>
                    <input
                      bind:value={chart.chart}
                      placeholder="nginx"
                      class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
                    />
                  </label>
                </div>

                <div class="grid grid-cols-2 gap-3 mb-3">
                  <label class="block col-span-1">
                    <span class="text-xs font-medium text-gray-600 dark:text-gray-400">Repo URL *</span>
                    <input
                      bind:value={chart.repo}
                      placeholder="https://charts.bitnami.com/bitnami"
                      class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
                    />
                  </label>
                  <label class="block">
                    <span class="text-xs font-medium text-gray-600 dark:text-gray-400">Version <span class="font-normal text-gray-400">(empty = latest)</span></span>
                    <input
                      bind:value={chart.version}
                      placeholder="1.2.3"
                      class="mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
                    />
                  </label>
                </div>

                <label class="block">
                  <span class="text-xs font-medium text-gray-600 dark:text-gray-400">Values <span class="font-normal text-gray-400">(YAML)</span></span>
                  <div class="mt-1">
                    <YamlEditor bind:value={chart.values} placeholder="replicaCount: 1&#10;service:&#10;  type: ClusterIP&#10;" />
                  </div>
                </label>
              </div>
            {/each}
          </div>
        {/if}
      </section>

      <div class="flex justify-end gap-3 border-t border-gray-200 dark:border-gray-700 pt-6">
        <button onclick={() => (showForm = false)} class="px-4 py-2 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors">Cancel</button>
        <button
          onclick={saveBlueprint}
          disabled={saving || !formName.trim()}
          class="px-4 py-2 rounded-lg bg-brand text-gray-900 text-sm font-semibold hover:bg-brand-dim disabled:opacity-50 transition-colors"
        >
          {saving ? "Saving…" : "Save blueprint"}
        </button>
      </div>
    </div>
  </div>
{/if}
