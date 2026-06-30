package dto

type NodeDTO struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	State    string `json:"state"`
	Image    string `json:"image"`
	K8sReady *bool  `json:"k8sReady"`
}

type ClusterDTO struct {
	Name    string    `json:"name"`
	Servers int       `json:"servers"`
	Agents  int       `json:"agents"`
	Status  string    `json:"status"` // "running" | "stopped" | "partial"
	Nodes   []NodeDTO `json:"nodes"`
	Created string    `json:"created"`
}

// ClusterCreateRequest is the simple (quick-create) form.
type ClusterCreateRequest struct {
	Name    string `json:"name"`
	Servers int    `json:"servers"`
	Agents  int    `json:"agents"`
	Image   string `json:"image"`   // empty = default
	APIPort string `json:"apiPort"` // empty = random
}

// NodeFilter pairs a value with an optional k3d node filter expression (e.g. "@server:0", "@agent:*").
type NodeFilter struct {
	Value       string `json:"value"`
	NodeFilters string `json:"nodeFilters"` // single string, comma-separated filters
}

// NodeUpgradeRequest specifies a new image for an existing node.
type NodeUpgradeRequest struct {
	NodeName string `json:"nodeName"`
	Image    string `json:"image"`
}

// ClusterCreateAdvancedRequest covers all configurable aspects of k3d cluster create.
type ClusterCreateAdvancedRequest struct {
	Name string `json:"name"`

	// Topology
	Servers int    `json:"servers"`
	Agents  int    `json:"agents"`
	Image   string `json:"image"`

	// API exposure
	APIPort   string `json:"apiPort"`
	APIHost   string `json:"apiHost"`
	APIHostIP string `json:"apiHostIP"`

	// Networking
	Network string `json:"network"`
	Subnet  string `json:"subnet"`
	Token   string `json:"token"`

	// Resource limits
	ServersMemory string `json:"serversMemory"`
	AgentsMemory  string `json:"agentsMemory"`
	GPURequest    string `json:"gpuRequest"`

	// Port mappings  — "HOST:HOSTPORT:CONTAINERPORT/PROTO@nodefilter"
	Ports []NodeFilter `json:"ports"`

	// Volume mounts — "source:dest@nodefilter"
	Volumes []NodeFilter `json:"volumes"`

	// Env vars — "KEY=VALUE@nodefilter"
	Env []NodeFilter `json:"env"`

	// k3s extra args — "arg@nodefilter"
	K3sArgs []NodeFilter `json:"k3sArgs"`

	// k3s node labels — "key=value@nodefilter"
	K3sNodeLabels []NodeFilter `json:"k3sNodeLabels"`

	// Runtime labels — "key=value@nodefilter"
	RuntimeLabels []NodeFilter `json:"runtimeLabels"`

	// Registries
	RegistryCreate string   `json:"registryCreate"` // "name[:host][:port]"
	RegistryUse    []string `json:"registryUse"`

	// k3d behaviour flags
	NoLoadbalancer     bool     `json:"noLoadbalancer"`
	NoImageVolume      bool     `json:"noImageVolume"`
	NoRollback         bool     `json:"noRollback"`
	Timeout            string   `json:"timeout"`           // options.k3d.timeout, e.g. "60s"
	LBConfigOverrides  []string `json:"lbConfigOverrides"` // options.k3d.loadbalancer.configOverrides

	// Runtime extras
	HostPidMode bool        `json:"hostPidMode"` // options.runtime.hostPidMode
	Ulimits     []UlimitDTO `json:"ulimits"`     // options.runtime.ulimits

	// Files (injected into nodes)
	Files []FileDTO `json:"files"`

	// Host aliases (/etc/hosts injections)
	HostAliases []HostAliasDTO `json:"hostAliases"`

	// Registry extras
	RegistryCreateHost  string   `json:"registryCreateHost"`  // registries.create.host
	RegistryCreatePort  string   `json:"registryCreatePort"`  // registries.create.hostPort
	RegistryProxyURL    string   `json:"registryProxyURL"`    // registries.create.proxy.remoteURL
	RegistryProxyUser   string   `json:"registryProxyUser"`   // registries.create.proxy.username
	RegistryProxyPass   string   `json:"registryProxyPass"`   // registries.create.proxy.password
	RegistryVolumes     []string `json:"registryVolumes"`     // registries.create.volumes
	RegistryConfig      string   `json:"registryConfig"`      // registries.config (inline YAML or path)

	// Kubeconfig
	UpdateKubeconfig bool `json:"updateKubeconfig"`
	SwitchContext    bool `json:"switchContext"`
}
