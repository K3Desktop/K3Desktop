<script lang="ts">
  import type { AdvState } from "./prefill.ts";

  let {
    adv = $bindable(),
    k3sVersions = [],
    versionsLoading = false,
    datalistId = "k3s-versions-form",
  }: {
    adv: AdvState;
    k3sVersions?: string[];
    versionsLoading?: boolean;
    datalistId?: string;
  } = $props();

  function addRow(arr: { value: string; nodeFilters: string }[]) {
    arr.push({ value: "", nodeFilters: "" });
  }
  function removeRow(arr: { value: string; nodeFilters: string }[], i: number) {
    arr.splice(i, 1);
  }
  function addUlimit() {
    adv.ulimits.push({ name: "", soft: 0, hard: 0 });
  }
  function removeUlimit(i: number) {
    adv.ulimits.splice(i, 1);
  }
  function addFile() {
    adv.files.push({ source: "", destination: "", description: "", nodeFilters: "" });
  }
  function removeFile(i: number) {
    adv.files.splice(i, 1);
  }
  function addHostAlias() {
    adv.hostAliases.push({ ip: "", hostnames: "" });
  }
  function removeHostAlias(i: number) {
    adv.hostAliases.splice(i, 1);
  }

  const input = "mt-1 block w-full rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-accent";
  const inputSm = "rounded-lg border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-accent";
  const label = "text-sm text-gray-600 dark:text-gray-400";
  const sectionHead = "text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500 mb-3";
</script>

<!-- Topology -->
<section class="mb-6">
  <p class={sectionHead}>Topology</p>
  <div class="grid grid-cols-2 gap-3">
    <label class="block col-span-2">
      <span class={label}>Name *</span>
      <input bind:value={adv.name} class={input} placeholder="my-cluster" />
    </label>
    <label class="block">
      <span class={label}>Servers</span>
      <input type="number" bind:value={adv.servers} min="1" class={input} />
    </label>
    <label class="block">
      <span class={label}>Agents</span>
      <input type="number" bind:value={adv.agents} min="0" class={input} />
    </label>
    <label class="block col-span-2">
      <span class={label}>Image</span>
      <input bind:value={adv.image} list={datalistId}
        class={input}
        placeholder={versionsLoading ? "Loading versions…" : "docker.io/rancher/k3s:latest"} />
      <datalist id={datalistId}>
        {#each k3sVersions as v}
          <option value="docker.io/rancher/k3s:{v}">{v}</option>
        {/each}
      </datalist>
    </label>
  </div>
</section>

<!-- API / Networking -->
<section class="mb-6">
  <p class={sectionHead}>API & Networking</p>
  <div class="grid grid-cols-3 gap-3">
    <label class="block">
      <span class={label}>API port</span>
      <input bind:value={adv.apiPort} class={input} placeholder="6443" />
    </label>
    <label class="block">
      <span class={label}>API host</span>
      <input bind:value={adv.apiHost} class={input} placeholder="localhost" />
    </label>
    <label class="block">
      <span class={label}>API host IP</span>
      <input bind:value={adv.apiHostIP} class={input} placeholder="0.0.0.0" />
    </label>
    <label class="block">
      <span class={label}>Network</span>
      <input bind:value={adv.network} class={input} placeholder="existing-network" />
    </label>
    <label class="block">
      <span class={label}>Subnet</span>
      <input bind:value={adv.subnet} class={input} placeholder="172.28.0.0/16" />
    </label>
    <label class="block">
      <span class={label}>Token</span>
      <input bind:value={adv.token} class={input} placeholder="auto-generated" />
    </label>
  </div>
</section>

<!-- Resources -->
<section class="mb-6">
  <p class={sectionHead}>Resources</p>
  <div class="grid grid-cols-3 gap-3">
    <label class="block">
      <span class={label}>Servers memory</span>
      <input bind:value={adv.serversMemory} class={input} placeholder="2g" />
    </label>
    <label class="block">
      <span class={label}>Agents memory</span>
      <input bind:value={adv.agentsMemory} class={input} placeholder="2g" />
    </label>
    <label class="block">
      <span class={label}>GPU request</span>
      <input bind:value={adv.gpuRequest} class={input} placeholder="all" />
    </label>
  </div>
</section>

<!-- Node filter sections -->
{#each [
  { label: "Port mappings", hint: "HOST:HOSTPORT:CONTAINERPORT/PROTO", arr: adv.ports },
  { label: "Volume mounts", hint: "/host/path:/container/path", arr: adv.volumes },
  { label: "Environment variables", hint: "KEY=VALUE", arr: adv.env },
  { label: "k3s extra args", hint: "--disable=traefik", arr: adv.k3sArgs },
  { label: "k3s node labels", hint: "key=value", arr: adv.k3sNodeLabels },
  { label: "Runtime labels", hint: "key=value", arr: adv.runtimeLabels },
] as section}
  <section class="mb-6">
    <div class="flex items-center justify-between mb-2">
      <p class={sectionHead} style="margin:0">{section.label}</p>
      <button onclick={() => addRow(section.arr)} class="text-xs text-accent hover:underline">+ Add</button>
    </div>
    {#each section.arr as row, i}
      <div class="flex gap-2 mb-2">
        <input bind:value={row.value} placeholder={section.hint} class="flex-1 {inputSm}" />
        <input bind:value={row.nodeFilters} placeholder="@server:0" class="w-32 {inputSm}" />
        <button onclick={() => removeRow(section.arr, i)} class="px-2 text-gray-400 hover:text-red-500 transition-colors">✕</button>
      </div>
    {/each}
    {#if section.arr.length === 0}
      <p class="text-xs text-gray-400 dark:text-gray-600">None. Click + Add to add entries.</p>
    {/if}
  </section>
{/each}

<!-- Registries -->
<section class="mb-6">
  <p class={sectionHead}>Registries</p>
  <div class="grid grid-cols-2 gap-3 mb-3">
    <label class="block">
      <span class={label}>Create registry — name</span>
      <input bind:value={adv.registryCreate} class={input} placeholder="my-registry" />
    </label>
    <label class="block">
      <span class={label}>Use registries <span class="text-gray-400">(comma-separated)</span></span>
      <input bind:value={adv.registryUse} class={input} placeholder="k3d-reg1,k3d-reg2" />
    </label>
    <label class="block">
      <span class={label}>Registry host</span>
      <input bind:value={adv.registryCreateHost} class={input} placeholder="0.0.0.0" />
    </label>
    <label class="block">
      <span class={label}>Registry host port</span>
      <input bind:value={adv.registryCreatePort} class={input} placeholder="5000" />
    </label>
    <label class="block">
      <span class={label}>Proxy remote URL</span>
      <input bind:value={adv.registryProxyURL} class={input} placeholder="https://registry-1.docker.io" />
    </label>
    <label class="block">
      <span class={label}>Proxy username</span>
      <input bind:value={adv.registryProxyUser} class={input} placeholder="(optional)" />
    </label>
    <label class="block col-span-2">
      <span class={label}>Proxy password</span>
      <input type="password" bind:value={adv.registryProxyPass} class={input} placeholder="(optional)" />
    </label>
    <label class="block col-span-2">
      <span class={label}>Registry volumes <span class="text-gray-400">(one per line)</span></span>
      <textarea bind:value={adv.registryVolumes} rows="2" class={input} placeholder="/tmp/registry:/var/lib/registry"></textarea>
    </label>
    <label class="block col-span-2">
      <span class={label}>Registry config <span class="text-gray-400">(inline YAML or file path)</span></span>
      <textarea bind:value={adv.registryConfig} rows="3" class={input} placeholder="mirrors:&#10;  &quot;my.registry&quot;:&#10;    endpoint: [...]"></textarea>
    </label>
  </div>
</section>

<!-- Files -->
<section class="mb-6">
  <div class="flex items-center justify-between mb-2">
    <p class={sectionHead} style="margin:0">Files</p>
    <button onclick={addFile} class="text-xs text-accent hover:underline">+ Add</button>
  </div>
  {#each adv.files as f, i}
    <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-3 mb-2 grid grid-cols-2 gap-2">
      <label class="block col-span-2">
        <span class={label}>Source <span class="text-gray-400">(path or inline YAML)</span></span>
        <textarea bind:value={f.source} rows="2" class={input} placeholder="/path/to/file.yaml or inline YAML"></textarea>
      </label>
      <label class="block">
        <span class={label}>Destination</span>
        <input bind:value={f.destination} class={input} placeholder="/var/lib/rancher/k3s/server/manifests/foo.yaml" />
      </label>
      <label class="block">
        <span class={label}>Node filters</span>
        <input bind:value={f.nodeFilters} class={input} placeholder="server:*" />
      </label>
      <label class="block col-span-2">
        <span class={label}>Description <span class="text-gray-400">(optional)</span></span>
        <input bind:value={f.description} class={input} placeholder="(optional)" />
      </label>
      <div class="col-span-2 flex justify-end">
        <button onclick={() => removeFile(i)} class="text-xs text-red-500 hover:underline">Remove</button>
      </div>
    </div>
  {/each}
  {#if adv.files.length === 0}
    <p class="text-xs text-gray-400 dark:text-gray-600">None.</p>
  {/if}
</section>

<!-- Host aliases -->
<section class="mb-6">
  <div class="flex items-center justify-between mb-2">
    <p class={sectionHead} style="margin:0">Host aliases</p>
    <button onclick={addHostAlias} class="text-xs text-accent hover:underline">+ Add</button>
  </div>
  {#each adv.hostAliases as ha, i}
    <div class="flex gap-2 mb-2">
      <input bind:value={ha.ip} placeholder="1.2.3.4" class="w-36 {inputSm}" />
      <input bind:value={ha.hostnames} placeholder="host.local,other.local" class="flex-1 {inputSm}" />
      <button onclick={() => removeHostAlias(i)} class="px-2 text-gray-400 hover:text-red-500 transition-colors">✕</button>
    </div>
  {/each}
  {#if adv.hostAliases.length === 0}
    <p class="text-xs text-gray-400 dark:text-gray-600">None.</p>
  {/if}
</section>

<!-- Runtime ulimits -->
<section class="mb-6">
  <div class="flex items-center justify-between mb-2">
    <p class={sectionHead} style="margin:0">Ulimits</p>
    <button onclick={addUlimit} class="text-xs text-accent hover:underline">+ Add</button>
  </div>
  {#each adv.ulimits as u, i}
    <div class="flex gap-2 mb-2">
      <input bind:value={u.name} placeholder="nofile" class="flex-1 {inputSm}" />
      <input type="number" bind:value={u.soft} placeholder="soft" class="w-24 {inputSm}" />
      <input type="number" bind:value={u.hard} placeholder="hard" class="w-24 {inputSm}" />
      <button onclick={() => removeUlimit(i)} class="px-2 text-gray-400 hover:text-red-500 transition-colors">✕</button>
    </div>
  {/each}
  {#if adv.ulimits.length === 0}
    <p class="text-xs text-gray-400 dark:text-gray-600">None.</p>
  {/if}
</section>

<!-- Behaviour flags -->
<section class="mb-8">
  <p class={sectionHead}>Behaviour</p>
  <div class="grid grid-cols-2 gap-2 mb-3">
    {#each [
      { key: "noLoadbalancer", label: "Disable load balancer" },
      { key: "noImageVolume", label: "Disable image volume" },
      { key: "noRollback", label: "Disable rollback on failure" },
      { key: "hostPidMode", label: "Host PID mode" },
      { key: "updateKubeconfig", label: "Update default kubeconfig" },
      { key: "switchContext", label: "Switch kubeconfig context" },
    ] as flag}
      <label class="flex items-center gap-2 cursor-pointer select-none">
        <input type="checkbox" bind:checked={(adv as any)[flag.key]} class="rounded accent-brand" />
        <span class="text-sm text-gray-700 dark:text-gray-300">{flag.label}</span>
      </label>
    {/each}
  </div>
  <div class="grid grid-cols-2 gap-3">
    <label class="block">
      <span class={label}>Timeout <span class="text-gray-400">(e.g. 60s, 2m)</span></span>
      <input bind:value={adv.timeout} class={input} placeholder="60s" />
    </label>
    <label class="block">
      <span class={label}>LB config overrides <span class="text-gray-400">(one per line)</span></span>
      <textarea bind:value={adv.lbConfigOverrides} rows="2" class={input} placeholder="settings.workerConnections=2048"></textarea>
    </label>
  </div>
</section>
