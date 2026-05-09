import { writable } from "svelte/store";
import type { ClusterCreateAdvancedRequest } from "../../bindings/github.com/k3desktop/k3desktop/dto";

export type AdvState = {
  name: string;
  servers: number;
  agents: number;
  image: string;
  apiPort: string;
  apiHost: string;
  apiHostIP: string;
  network: string;
  subnet: string;
  token: string;
  serversMemory: string;
  agentsMemory: string;
  gpuRequest: string;
  ports: { value: string; nodeFilters: string }[];
  volumes: { value: string; nodeFilters: string }[];
  env: { value: string; nodeFilters: string }[];
  k3sArgs: { value: string; nodeFilters: string }[];
  k3sNodeLabels: { value: string; nodeFilters: string }[];
  runtimeLabels: { value: string; nodeFilters: string }[];
  registryCreate: string;
  registryCreateHost: string;
  registryCreatePort: string;
  registryProxyURL: string;
  registryProxyUser: string;
  registryProxyPass: string;
  registryVolumes: string; // newline-separated
  registryConfig: string;
  registryUse: string;
  noLoadbalancer: boolean;
  noImageVolume: boolean;
  noRollback: boolean;
  timeout: string;
  lbConfigOverrides: string; // newline-separated
  hostPidMode: boolean;
  ulimits: { name: string; soft: number; hard: number }[];
  files: { source: string; destination: string; description: string; nodeFilters: string }[];
  hostAliases: { ip: string; hostnames: string }[]; // hostnames comma-separated
  updateKubeconfig: boolean;
  switchContext: boolean;
};

export const defaultAdv = (): AdvState => ({
  name: "",
  servers: 1,
  agents: 0,
  image: "",
  apiPort: "",
  apiHost: "",
  apiHostIP: "",
  network: "",
  subnet: "",
  token: "",
  serversMemory: "",
  agentsMemory: "",
  gpuRequest: "",
  ports: [],
  volumes: [],
  env: [],
  k3sArgs: [],
  k3sNodeLabels: [],
  runtimeLabels: [],
  registryCreate: "",
  registryCreateHost: "",
  registryCreatePort: "",
  registryProxyURL: "",
  registryProxyUser: "",
  registryProxyPass: "",
  registryVolumes: "",
  registryConfig: "",
  registryUse: "",
  noLoadbalancer: false,
  noImageVolume: false,
  noRollback: false,
  timeout: "",
  lbConfigOverrides: "",
  hostPidMode: false,
  ulimits: [],
  files: [],
  hostAliases: [],
  updateKubeconfig: true,
  switchContext: true,
});

// Set this before navigating to Clusters to pre-fill the advanced form.
export const clusterFormPrefill = writable<AdvState | null>(null);

export function dtoToAdv(req: ClusterCreateAdvancedRequest): AdvState {
  const adv = defaultAdv();
  adv.name = req.name ?? "";
  adv.servers = req.servers ?? 1;
  adv.agents = req.agents ?? 0;
  adv.image = req.image ?? "";
  adv.apiPort = req.apiPort ?? "";
  adv.apiHost = req.apiHost ?? "";
  adv.apiHostIP = req.apiHostIP ?? "";
  adv.network = req.network ?? "";
  adv.subnet = req.subnet ?? "";
  adv.token = req.token ?? "";
  adv.serversMemory = req.serversMemory ?? "";
  adv.agentsMemory = req.agentsMemory ?? "";
  adv.gpuRequest = req.gpuRequest ?? "";
  adv.noLoadbalancer = req.noLoadbalancer ?? false;
  adv.noImageVolume = req.noImageVolume ?? false;
  adv.noRollback = req.noRollback ?? false;
  adv.updateKubeconfig = req.updateKubeconfig ?? true;
  adv.switchContext = req.switchContext ?? true;
  adv.timeout = req.timeout ?? "";
  adv.lbConfigOverrides = (req.lbConfigOverrides ?? []).join("\n");
  adv.hostPidMode = req.hostPidMode ?? false;
  adv.registryCreate = req.registryCreate ?? "";
  adv.registryCreateHost = req.registryCreateHost ?? "";
  adv.registryCreatePort = req.registryCreatePort ?? "";
  adv.registryProxyURL = req.registryProxyURL ?? "";
  adv.registryProxyUser = req.registryProxyUser ?? "";
  adv.registryProxyPass = req.registryProxyPass ?? "";
  adv.registryVolumes = (req.registryVolumes ?? []).join("\n");
  adv.registryConfig = req.registryConfig ?? "";
  adv.registryUse = (req.registryUse ?? []).join(",");
  adv.ports = (req.ports ?? []).map((p) => ({ value: p.value ?? "", nodeFilters: p.nodeFilters ?? "" }));
  adv.volumes = (req.volumes ?? []).map((v) => ({ value: v.value ?? "", nodeFilters: v.nodeFilters ?? "" }));
  adv.env = (req.env ?? []).map((e) => ({ value: e.value ?? "", nodeFilters: e.nodeFilters ?? "" }));
  adv.k3sArgs = (req.k3sArgs ?? []).map((a) => ({ value: a.value ?? "", nodeFilters: a.nodeFilters ?? "" }));
  adv.k3sNodeLabels = (req.k3sNodeLabels ?? []).map((l) => ({ value: l.value ?? "", nodeFilters: l.nodeFilters ?? "" }));
  adv.runtimeLabels = (req.runtimeLabels ?? []).map((l) => ({ value: l.value ?? "", nodeFilters: l.nodeFilters ?? "" }));
  adv.ulimits = (req.ulimits ?? []).map((u) => ({ name: u.name ?? "", soft: u.soft ?? 0, hard: u.hard ?? 0 }));
  adv.files = (req.files ?? []).map((f) => ({
    source: f.source ?? "",
    destination: f.destination ?? "",
    description: f.description ?? "",
    nodeFilters: f.nodeFilters ?? "",
  }));
  adv.hostAliases = (req.hostAliases ?? []).map((ha) => ({
    ip: ha.ip ?? "",
    hostnames: (ha.hostnames ?? []).join(","),
  }));
  return adv;
}

export function advToDto(adv: AdvState): ClusterCreateAdvancedRequest {
  return {
    name: adv.name,
    servers: adv.servers,
    agents: adv.agents,
    image: adv.image,
    apiPort: adv.apiPort,
    apiHost: adv.apiHost,
    apiHostIP: adv.apiHostIP,
    network: adv.network,
    subnet: adv.subnet,
    token: adv.token,
    serversMemory: adv.serversMemory,
    agentsMemory: adv.agentsMemory,
    gpuRequest: adv.gpuRequest,
    noLoadbalancer: adv.noLoadbalancer,
    noImageVolume: adv.noImageVolume,
    noRollback: adv.noRollback,
    updateKubeconfig: adv.updateKubeconfig,
    switchContext: adv.switchContext,
    timeout: adv.timeout,
    lbConfigOverrides: adv.lbConfigOverrides ? adv.lbConfigOverrides.split("\n").filter(Boolean) : [],
    hostPidMode: adv.hostPidMode,
    registryCreate: adv.registryCreate,
    registryCreateHost: adv.registryCreateHost,
    registryCreatePort: adv.registryCreatePort,
    registryProxyURL: adv.registryProxyURL,
    registryProxyUser: adv.registryProxyUser,
    registryProxyPass: adv.registryProxyPass,
    registryVolumes: adv.registryVolumes ? adv.registryVolumes.split("\n").filter(Boolean) : [],
    registryConfig: adv.registryConfig,
    registryUse: adv.registryUse ? adv.registryUse.split(",").map((s) => s.trim()).filter(Boolean) : [],
    ports: adv.ports.filter((r) => r.value).map((r) => ({ value: r.value, nodeFilters: r.nodeFilters })),
    volumes: adv.volumes.filter((r) => r.value).map((r) => ({ value: r.value, nodeFilters: r.nodeFilters })),
    env: adv.env.filter((r) => r.value).map((r) => ({ value: r.value, nodeFilters: r.nodeFilters })),
    k3sArgs: adv.k3sArgs.filter((r) => r.value).map((r) => ({ value: r.value, nodeFilters: r.nodeFilters })),
    k3sNodeLabels: adv.k3sNodeLabels.filter((r) => r.value).map((r) => ({ value: r.value, nodeFilters: r.nodeFilters })),
    runtimeLabels: adv.runtimeLabels.filter((r) => r.value).map((r) => ({ value: r.value, nodeFilters: r.nodeFilters })),
    ulimits: adv.ulimits.filter((u) => u.name).map((u) => ({ name: u.name, soft: u.soft, hard: u.hard })),
    files: adv.files.filter((f) => f.source || f.destination).map((f) => ({
      source: f.source,
      destination: f.destination,
      description: f.description,
      nodeFilters: f.nodeFilters,
    })),
    hostAliases: adv.hostAliases.filter((ha) => ha.ip).map((ha) => ({
      ip: ha.ip,
      hostnames: ha.hostnames ? ha.hostnames.split(",").map((s) => s.trim()).filter(Boolean) : [],
    })),
  } as unknown as ClusterCreateAdvancedRequest;
}
