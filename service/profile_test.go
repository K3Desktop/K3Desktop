package service

import (
	"strings"
	"testing"
	"time"

	k3dconfigtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	confv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	k3dtypes "github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/k3desktop/k3desktop/dto"
)

// TestAdvancedRequestToSimpleConfig_Basic verifies the basic fields are mapped.
func TestAdvancedRequestToSimpleConfig_Basic(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "test-cluster",
		Servers: 3,
		Agents:  2,
		Image:   "rancher/k3s:v1.28.0-k3s1",
		Network: "my-net",
		Subnet:  "172.28.0.0/16",
		Token:   "secret-token",
	}

	sc := advancedRequestToSimpleConfig(req)

	if sc.ObjectMeta.Name != "test-cluster" {
		t.Errorf("Name = %q, want %q", sc.ObjectMeta.Name, "test-cluster")
	}
	if sc.Servers != 3 {
		t.Errorf("Servers = %d, want %d", sc.Servers, 3)
	}
	if sc.Agents != 2 {
		t.Errorf("Agents = %d, want %d", sc.Agents, 2)
	}
	if sc.Image != "rancher/k3s:v1.28.0-k3s1" {
		t.Errorf("Image = %q, want %q", sc.Image, "rancher/k3s:v1.28.0-k3s1")
	}
	if sc.Network != "my-net" {
		t.Errorf("Network = %q, want %q", sc.Network, "my-net")
	}
	if sc.Subnet != "172.28.0.0/16" {
		t.Errorf("Subnet = %q, want %q", sc.Subnet, "172.28.0.0/16")
	}
	if sc.ClusterToken != "secret-token" {
		t.Errorf("ClusterToken = %q, want %q", sc.ClusterToken, "secret-token")
	}
	if sc.TypeMeta.Kind != "Simple" {
		t.Errorf("Kind = %q, want %q", sc.TypeMeta.Kind, "Simple")
	}
	if sc.TypeMeta.APIVersion != confv1alpha5.ApiVersion {
		t.Errorf("APIVersion = %q, want %q", sc.TypeMeta.APIVersion, confv1alpha5.ApiVersion)
	}
}

// TestAdvancedRequestToSimpleConfig_API tests API exposure mapping.
func TestAdvancedRequestToSimpleConfig_API(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:      "api-test",
		Servers:   1,
		APIPort:   "6443",
		APIHost:   "myhost",
		APIHostIP: "127.0.0.1",
	}

	sc := advancedRequestToSimpleConfig(req)

	if sc.ExposeAPI.HostPort != "6443" {
		t.Errorf("ExposeAPI.HostPort = %q, want %q", sc.ExposeAPI.HostPort, "6443")
	}
	if sc.ExposeAPI.Host != "myhost" {
		t.Errorf("ExposeAPI.Host = %q, want %q", sc.ExposeAPI.Host, "myhost")
	}
	if sc.ExposeAPI.HostIP != "127.0.0.1" {
		t.Errorf("ExposeAPI.HostIP = %q, want %q", sc.ExposeAPI.HostIP, "127.0.0.1")
	}
}

// TestAdvancedRequestToSimpleConfig_Resources tests resource limits mapping.
func TestAdvancedRequestToSimpleConfig_Resources(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:          "res-test",
		Servers:       1,
		ServersMemory: "4g",
		AgentsMemory:  "2g",
		GPURequest:    "all",
		HostPidMode:   true,
	}

	sc := advancedRequestToSimpleConfig(req)

	if sc.Options.Runtime.ServersMemory != "4g" {
		t.Errorf("ServersMemory = %q, want %q", sc.Options.Runtime.ServersMemory, "4g")
	}
	if sc.Options.Runtime.AgentsMemory != "2g" {
		t.Errorf("AgentsMemory = %q, want %q", sc.Options.Runtime.AgentsMemory, "2g")
	}
	if sc.Options.Runtime.GPURequest != "all" {
		t.Errorf("GPURequest = %q, want %q", sc.Options.Runtime.GPURequest, "all")
	}
	if !sc.Options.Runtime.HostPidMode {
		t.Error("HostPidMode = false, want true")
	}
}

// TestAdvancedRequestToSimpleConfig_K3dFlags tests k3d behaviour flags.
func TestAdvancedRequestToSimpleConfig_K3dFlags(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:              "flags-test",
		Servers:           1,
		NoLoadbalancer:    true,
		NoImageVolume:     true,
		NoRollback:        true,
		Timeout:           "120s",
		LBConfigOverrides: []string{"override1", "override2"},
	}

	sc := advancedRequestToSimpleConfig(req)

	if !sc.Options.K3dOptions.DisableLoadbalancer {
		t.Error("DisableLoadbalancer = false, want true")
	}
	if !sc.Options.K3dOptions.DisableImageVolume {
		t.Error("DisableImageVolume = false, want true")
	}
	if !sc.Options.K3dOptions.NoRollback {
		t.Error("NoRollback = false, want true")
	}
	if sc.Options.K3dOptions.Timeout != 120*time.Second {
		t.Errorf("Timeout = %v, want %v", sc.Options.K3dOptions.Timeout, 120*time.Second)
	}
	if len(sc.Options.K3dOptions.Loadbalancer.ConfigOverrides) != 2 {
		t.Fatalf("ConfigOverrides len = %d, want 2", len(sc.Options.K3dOptions.Loadbalancer.ConfigOverrides))
	}
}

// TestAdvancedRequestToSimpleConfig_Kubeconfig tests kubeconfig option mapping.
func TestAdvancedRequestToSimpleConfig_Kubeconfig(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:             "kube-test",
		Servers:          1,
		UpdateKubeconfig: true,
		SwitchContext:    true,
	}

	sc := advancedRequestToSimpleConfig(req)

	if !sc.Options.KubeconfigOptions.UpdateDefaultKubeconfig {
		t.Error("UpdateDefaultKubeconfig = false, want true")
	}
	if !sc.Options.KubeconfigOptions.SwitchCurrentContext {
		t.Error("SwitchCurrentContext = false, want true")
	}
}

// TestAdvancedRequestToSimpleConfig_Ports tests port mapping conversion.
func TestAdvancedRequestToSimpleConfig_Ports(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "ports-test",
		Servers: 1,
		Ports: []dto.NodeFilter{
			{Value: "8080:80/tcp", NodeFilters: "server:0,agent:*"},
			{Value: "9090:90", NodeFilters: ""},
		},
	}

	sc := advancedRequestToSimpleConfig(req)

	if len(sc.Ports) != 2 {
		t.Fatalf("Ports len = %d, want 2", len(sc.Ports))
	}
	if sc.Ports[0].Port != "8080:80/tcp" {
		t.Errorf("Ports[0].Port = %q, want %q", sc.Ports[0].Port, "8080:80/tcp")
	}
	if len(sc.Ports[0].NodeFilters) != 2 {
		t.Errorf("Ports[0].NodeFilters len = %d, want 2", len(sc.Ports[0].NodeFilters))
	}
	if len(sc.Ports[1].NodeFilters) != 0 {
		t.Errorf("Ports[1].NodeFilters len = %d, want 0", len(sc.Ports[1].NodeFilters))
	}
}

// TestAdvancedRequestToSimpleConfig_Volumes tests volume mapping conversion.
func TestAdvancedRequestToSimpleConfig_Volumes(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "vol-test",
		Servers: 1,
		Volumes: []dto.NodeFilter{
			{Value: "/tmp/data:/data", NodeFilters: "server:0"},
		},
	}

	sc := advancedRequestToSimpleConfig(req)

	if len(sc.Volumes) != 1 {
		t.Fatalf("Volumes len = %d, want 1", len(sc.Volumes))
	}
	if sc.Volumes[0].Volume != "/tmp/data:/data" {
		t.Errorf("Volume = %q, want %q", sc.Volumes[0].Volume, "/tmp/data:/data")
	}
}

// TestAdvancedRequestToSimpleConfig_Env tests environment variable mapping.
func TestAdvancedRequestToSimpleConfig_Env(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "env-test",
		Servers: 1,
		Env: []dto.NodeFilter{
			{Value: "FOO=bar", NodeFilters: "agent:*"},
		},
	}

	sc := advancedRequestToSimpleConfig(req)

	if len(sc.Env) != 1 {
		t.Fatalf("Env len = %d, want 1", len(sc.Env))
	}
	if sc.Env[0].EnvVar != "FOO=bar" {
		t.Errorf("EnvVar = %q, want %q", sc.Env[0].EnvVar, "FOO=bar")
	}
}

// TestAdvancedRequestToSimpleConfig_K3sArgs tests k3s extra args mapping.
func TestAdvancedRequestToSimpleConfig_K3sArgs(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "args-test",
		Servers: 1,
		K3sArgs: []dto.NodeFilter{
			{Value: "--disable=traefik", NodeFilters: "server:*"},
		},
	}

	sc := advancedRequestToSimpleConfig(req)

	if len(sc.Options.K3sOptions.ExtraArgs) != 1 {
		t.Fatalf("ExtraArgs len = %d, want 1", len(sc.Options.K3sOptions.ExtraArgs))
	}
	if sc.Options.K3sOptions.ExtraArgs[0].Arg != "--disable=traefik" {
		t.Errorf("Arg = %q, want %q", sc.Options.K3sOptions.ExtraArgs[0].Arg, "--disable=traefik")
	}
}

// TestAdvancedRequestToSimpleConfig_Ulimits tests ulimit mapping.
func TestAdvancedRequestToSimpleConfig_Ulimits(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "ulimit-test",
		Servers: 1,
		Ulimits: []dto.UlimitDTO{
			{Name: "nofile", Soft: 1024, Hard: 4096},
		},
	}

	sc := advancedRequestToSimpleConfig(req)

	if len(sc.Options.Runtime.Ulimits) != 1 {
		t.Fatalf("Ulimits len = %d, want 1", len(sc.Options.Runtime.Ulimits))
	}
	if sc.Options.Runtime.Ulimits[0].Name != "nofile" {
		t.Errorf("Ulimit name = %q, want %q", sc.Options.Runtime.Ulimits[0].Name, "nofile")
	}
	if sc.Options.Runtime.Ulimits[0].Soft != 1024 {
		t.Errorf("Ulimit soft = %d, want %d", sc.Options.Runtime.Ulimits[0].Soft, 1024)
	}
	if sc.Options.Runtime.Ulimits[0].Hard != 4096 {
		t.Errorf("Ulimit hard = %d, want %d", sc.Options.Runtime.Ulimits[0].Hard, 4096)
	}
}

// TestAdvancedRequestToSimpleConfig_Files tests file injection mapping.
func TestAdvancedRequestToSimpleConfig_Files(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "files-test",
		Servers: 1,
		Files: []dto.FileDTO{
			{Source: "/local/file.txt", Destination: "/node/file.txt", Description: "config", NodeFilters: "server:0"},
		},
	}

	sc := advancedRequestToSimpleConfig(req)

	if len(sc.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(sc.Files))
	}
	if sc.Files[0].Source != "/local/file.txt" {
		t.Errorf("Source = %q, want %q", sc.Files[0].Source, "/local/file.txt")
	}
	if sc.Files[0].Destination != "/node/file.txt" {
		t.Errorf("Destination = %q, want %q", sc.Files[0].Destination, "/node/file.txt")
	}
}

// TestAdvancedRequestToSimpleConfig_HostAliases tests host alias mapping.
func TestAdvancedRequestToSimpleConfig_HostAliases(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "hostalias-test",
		Servers: 1,
		HostAliases: []dto.HostAliasDTO{
			{IP: "10.0.0.1", Hostnames: []string{"myhost", "myhost.local"}},
		},
	}

	sc := advancedRequestToSimpleConfig(req)

	if len(sc.HostAliases) != 1 {
		t.Fatalf("HostAliases len = %d, want 1", len(sc.HostAliases))
	}
	if sc.HostAliases[0].IP != "10.0.0.1" {
		t.Errorf("IP = %q, want %q", sc.HostAliases[0].IP, "10.0.0.1")
	}
	if len(sc.HostAliases[0].Hostnames) != 2 {
		t.Errorf("Hostnames len = %d, want 2", len(sc.HostAliases[0].Hostnames))
	}
}

// TestAdvancedRequestToSimpleConfig_Registry tests registry mapping.
func TestAdvancedRequestToSimpleConfig_Registry(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:               "reg-test",
		Servers:            1,
		RegistryCreate:     "myreg",
		RegistryCreateHost: "0.0.0.0",
		RegistryCreatePort: "5000",
		RegistryVolumes:    []string{"/data:/var/lib/registry"},
		RegistryProxyURL:   "https://registry-1.docker.io",
		RegistryProxyUser:  "user",
		RegistryProxyPass:  "pass",
		RegistryUse:        []string{"k3d-existing:5000"},
		RegistryConfig:     "mirrors: {}",
	}

	sc := advancedRequestToSimpleConfig(req)

	if sc.Registries.Create == nil {
		t.Fatal("Registries.Create is nil")
	}
	if sc.Registries.Create.Name != "myreg" {
		t.Errorf("Registry Name = %q, want %q", sc.Registries.Create.Name, "myreg")
	}
	if sc.Registries.Create.Host != "0.0.0.0" {
		t.Errorf("Registry Host = %q, want %q", sc.Registries.Create.Host, "0.0.0.0")
	}
	if sc.Registries.Create.HostPort != "5000" {
		t.Errorf("Registry HostPort = %q, want %q", sc.Registries.Create.HostPort, "5000")
	}
	if sc.Registries.Create.Proxy.RemoteURL != "https://registry-1.docker.io" {
		t.Errorf("Proxy RemoteURL = %q, want %q", sc.Registries.Create.Proxy.RemoteURL, "https://registry-1.docker.io")
	}
	if sc.Registries.Create.Proxy.Username != "user" {
		t.Errorf("Proxy Username = %q, want %q", sc.Registries.Create.Proxy.Username, "user")
	}
	if len(sc.Registries.Use) != 1 || sc.Registries.Use[0] != "k3d-existing:5000" {
		t.Errorf("Registries.Use = %v, want [k3d-existing:5000]", sc.Registries.Use)
	}
	if sc.Registries.Config != "mirrors: {}" {
		t.Errorf("Registries.Config = %q, want %q", sc.Registries.Config, "mirrors: {}")
	}
}

// TestAdvancedRequestToSimpleConfig_NoRegistry tests no registry when fields empty.
func TestAdvancedRequestToSimpleConfig_NoRegistry(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "noreg-test",
		Servers: 1,
	}

	sc := advancedRequestToSimpleConfig(req)

	if sc.Registries.Create != nil {
		t.Error("Registries.Create should be nil when no registry fields set")
	}
}

// TestAdvancedRequestToSimpleConfig_InvalidTimeout tests invalid timeout is ignored.
func TestAdvancedRequestToSimpleConfig_InvalidTimeout(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "bad-timeout",
		Servers: 1,
		Timeout: "not-a-duration",
	}

	sc := advancedRequestToSimpleConfig(req)

	if sc.Options.K3dOptions.Timeout != 0 {
		t.Errorf("Timeout = %v, want 0 for invalid string", sc.Options.K3dOptions.Timeout)
	}
}

// TestSimpleConfigToAdvancedRequest_Basic tests the reverse mapping.
func TestSimpleConfigToAdvancedRequest_Basic(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		ObjectMeta:   k3dconfigtypes.ObjectMeta{Name: "reverse-test"},
		Servers:      3,
		Agents:       2,
		Image:        "rancher/k3s:v1.28.0-k3s1",
		Network:      "my-net",
		Subnet:       "172.28.0.0/16",
		ClusterToken: "secret-token",
	}

	req := simpleConfigToAdvancedRequest(sc)

	if req.Name != "reverse-test" {
		t.Errorf("Name = %q, want %q", req.Name, "reverse-test")
	}
	if req.Servers != 3 {
		t.Errorf("Servers = %d, want 3", req.Servers)
	}
	if req.Agents != 2 {
		t.Errorf("Agents = %d, want 2", req.Agents)
	}
	if req.Image != "rancher/k3s:v1.28.0-k3s1" {
		t.Errorf("Image = %q, want %q", req.Image, "rancher/k3s:v1.28.0-k3s1")
	}
	if req.Network != "my-net" {
		t.Errorf("Network = %q, want %q", req.Network, "my-net")
	}
	if req.Token != "secret-token" {
		t.Errorf("Token = %q, want %q", req.Token, "secret-token")
	}
}

// TestSimpleConfigToAdvancedRequest_API tests API exposure reverse mapping.
func TestSimpleConfigToAdvancedRequest_API(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		ExposeAPI: confv1alpha5.SimpleExposureOpts{
			HostPort: "6443",
			Host:     "myhost",
			HostIP:   "127.0.0.1",
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if req.APIPort != "6443" {
		t.Errorf("APIPort = %q, want %q", req.APIPort, "6443")
	}
	if req.APIHost != "myhost" {
		t.Errorf("APIHost = %q, want %q", req.APIHost, "myhost")
	}
	if req.APIHostIP != "127.0.0.1" {
		t.Errorf("APIHostIP = %q, want %q", req.APIHostIP, "127.0.0.1")
	}
}

// TestSimpleConfigToAdvancedRequest_RuntimeOptions tests runtime option reverse mapping.
func TestSimpleConfigToAdvancedRequest_RuntimeOptions(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		Options: confv1alpha5.SimpleConfigOptions{
			Runtime: confv1alpha5.SimpleConfigOptionsRuntime{
				ServersMemory: "4g",
				AgentsMemory:  "2g",
				GPURequest:    "all",
				HostPidMode:   true,
			},
			K3dOptions: confv1alpha5.SimpleConfigOptionsK3d{
				DisableLoadbalancer: true,
				DisableImageVolume:  true,
				NoRollback:          true,
				Timeout:             60 * time.Second,
			},
			KubeconfigOptions: confv1alpha5.SimpleConfigOptionsKubeconfig{
				UpdateDefaultKubeconfig: true,
				SwitchCurrentContext:    true,
			},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if req.ServersMemory != "4g" {
		t.Errorf("ServersMemory = %q, want %q", req.ServersMemory, "4g")
	}
	if req.AgentsMemory != "2g" {
		t.Errorf("AgentsMemory = %q, want %q", req.AgentsMemory, "2g")
	}
	if req.GPURequest != "all" {
		t.Errorf("GPURequest = %q, want %q", req.GPURequest, "all")
	}
	if !req.HostPidMode {
		t.Error("HostPidMode = false, want true")
	}
	if !req.NoLoadbalancer {
		t.Error("NoLoadbalancer = false, want true")
	}
	if !req.NoImageVolume {
		t.Error("NoImageVolume = false, want true")
	}
	if !req.NoRollback {
		t.Error("NoRollback = false, want true")
	}
	if req.Timeout != "1m0s" {
		t.Errorf("Timeout = %q, want %q", req.Timeout, "1m0s")
	}
	if !req.UpdateKubeconfig {
		t.Error("UpdateKubeconfig = false, want true")
	}
	if !req.SwitchContext {
		t.Error("SwitchContext = false, want true")
	}
}

// TestSimpleConfigToAdvancedRequest_Ports tests port reverse mapping.
func TestSimpleConfigToAdvancedRequest_Ports(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		Ports: []confv1alpha5.PortWithNodeFilters{
			{Port: "8080:80/tcp", NodeFilters: []string{"server:0", "agent:*"}},
			{Port: "9090:90"},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if len(req.Ports) != 2 {
		t.Fatalf("Ports len = %d, want 2", len(req.Ports))
	}
	if req.Ports[0].Value != "8080:80/tcp" {
		t.Errorf("Ports[0].Value = %q, want %q", req.Ports[0].Value, "8080:80/tcp")
	}
	if req.Ports[0].NodeFilters != "server:0,agent:*" {
		t.Errorf("Ports[0].NodeFilters = %q, want %q", req.Ports[0].NodeFilters, "server:0,agent:*")
	}
	if req.Ports[1].NodeFilters != "" {
		t.Errorf("Ports[1].NodeFilters = %q, want empty", req.Ports[1].NodeFilters)
	}
}

// TestSimpleConfigToAdvancedRequest_Ulimits tests ulimit reverse mapping.
func TestSimpleConfigToAdvancedRequest_Ulimits(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		Options: confv1alpha5.SimpleConfigOptions{
			Runtime: confv1alpha5.SimpleConfigOptionsRuntime{
				Ulimits: []confv1alpha5.Ulimit{
					{Name: "nofile", Soft: 1024, Hard: 4096},
				},
			},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if len(req.Ulimits) != 1 {
		t.Fatalf("Ulimits len = %d, want 1", len(req.Ulimits))
	}
	if req.Ulimits[0].Name != "nofile" {
		t.Errorf("Ulimit name = %q, want %q", req.Ulimits[0].Name, "nofile")
	}
}

// TestSimpleConfigToAdvancedRequest_HostAliases tests host aliases reverse mapping.
func TestSimpleConfigToAdvancedRequest_HostAliases(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		HostAliases: []k3dtypes.HostAlias{
			{IP: "10.0.0.1", Hostnames: []string{"foo", "bar"}},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if len(req.HostAliases) != 1 {
		t.Fatalf("HostAliases len = %d, want 1", len(req.HostAliases))
	}
	if req.HostAliases[0].IP != "10.0.0.1" {
		t.Errorf("IP = %q, want %q", req.HostAliases[0].IP, "10.0.0.1")
	}
	if len(req.HostAliases[0].Hostnames) != 2 {
		t.Errorf("Hostnames len = %d, want 2", len(req.HostAliases[0].Hostnames))
	}
}

// TestSimpleConfigToAdvancedRequest_Registry tests registry reverse mapping.
func TestSimpleConfigToAdvancedRequest_Registry(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		Registries: confv1alpha5.SimpleConfigRegistries{
			Use:    []string{"k3d-existing:5000"},
			Config: "mirrors: {}",
			Create: &confv1alpha5.SimpleConfigRegistryCreateConfig{
				Name:     "myreg",
				Host:     "0.0.0.0",
				HostPort: "5000",
				Volumes:  []string{"/data:/var/lib/registry"},
				Proxy: k3dtypes.RegistryProxy{
					RemoteURL: "https://registry-1.docker.io",
					Username:  "user",
					Password:  "pass",
				},
			},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if req.RegistryCreate != "myreg" {
		t.Errorf("RegistryCreate = %q, want %q", req.RegistryCreate, "myreg")
	}
	if req.RegistryCreateHost != "0.0.0.0" {
		t.Errorf("RegistryCreateHost = %q, want %q", req.RegistryCreateHost, "0.0.0.0")
	}
	if req.RegistryProxyURL != "https://registry-1.docker.io" {
		t.Errorf("RegistryProxyURL = %q, want %q", req.RegistryProxyURL, "https://registry-1.docker.io")
	}
	if len(req.RegistryUse) != 1 || req.RegistryUse[0] != "k3d-existing:5000" {
		t.Errorf("RegistryUse = %v, want [k3d-existing:5000]", req.RegistryUse)
	}
}

// TestSimpleConfigToAdvancedRequest_ZeroTimeout tests that zero timeout is not serialized.
func TestSimpleConfigToAdvancedRequest_ZeroTimeout(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{}
	req := simpleConfigToAdvancedRequest(sc)
	if req.Timeout != "" {
		t.Errorf("Timeout = %q, want empty for zero duration", req.Timeout)
	}
}

// TestRoundTrip_AdvancedRequest tests that converting to SimpleConfig and back preserves data.
func TestRoundTrip_AdvancedRequest(t *testing.T) {
	original := dto.ClusterCreateAdvancedRequest{
		Name:             "roundtrip",
		Servers:          3,
		Agents:           5,
		Image:            "rancher/k3s:v1.28.0-k3s1",
		APIPort:          "6443",
		APIHost:          "myhost",
		APIHostIP:        "0.0.0.0",
		Network:          "net1",
		Subnet:           "172.28.0.0/16",
		Token:            "token123",
		ServersMemory:    "4g",
		AgentsMemory:     "2g",
		GPURequest:       "all",
		NoLoadbalancer:   true,
		NoImageVolume:    true,
		NoRollback:       true,
		Timeout:          "120s",
		HostPidMode:      true,
		UpdateKubeconfig: true,
		SwitchContext:    true,
		Ports: []dto.NodeFilter{
			{Value: "8080:80", NodeFilters: "server:0"},
		},
		Volumes: []dto.NodeFilter{
			{Value: "/tmp:/data", NodeFilters: "agent:*"},
		},
		Env: []dto.NodeFilter{
			{Value: "FOO=bar", NodeFilters: "server:0"},
		},
		K3sArgs: []dto.NodeFilter{
			{Value: "--disable=traefik", NodeFilters: "server:*"},
		},
		K3sNodeLabels: []dto.NodeFilter{
			{Value: "foo=bar", NodeFilters: "server:0"},
		},
		RuntimeLabels: []dto.NodeFilter{
			{Value: "env=test", NodeFilters: "server:0"},
		},
		Ulimits: []dto.UlimitDTO{
			{Name: "nofile", Soft: 1024, Hard: 4096},
		},
		Files: []dto.FileDTO{
			{Source: "/src", Destination: "/dst", Description: "test", NodeFilters: "server:0"},
		},
		HostAliases: []dto.HostAliasDTO{
			{IP: "10.0.0.1", Hostnames: []string{"foo"}},
		},
	}

	sc := advancedRequestToSimpleConfig(original)
	result := simpleConfigToAdvancedRequest(sc)

	if result.Name != original.Name {
		t.Errorf("Name = %q, want %q", result.Name, original.Name)
	}
	if result.Servers != original.Servers {
		t.Errorf("Servers = %d, want %d", result.Servers, original.Servers)
	}
	if result.Agents != original.Agents {
		t.Errorf("Agents = %d, want %d", result.Agents, original.Agents)
	}
	if result.Image != original.Image {
		t.Errorf("Image = %q, want %q", result.Image, original.Image)
	}
	if result.APIPort != original.APIPort {
		t.Errorf("APIPort = %q, want %q", result.APIPort, original.APIPort)
	}
	if result.Network != original.Network {
		t.Errorf("Network = %q, want %q", result.Network, original.Network)
	}
	if result.NoLoadbalancer != original.NoLoadbalancer {
		t.Errorf("NoLoadbalancer = %v, want %v", result.NoLoadbalancer, original.NoLoadbalancer)
	}
	if len(result.Ports) != len(original.Ports) {
		t.Errorf("Ports len = %d, want %d", len(result.Ports), len(original.Ports))
	}
	if len(result.Volumes) != len(original.Volumes) {
		t.Errorf("Volumes len = %d, want %d", len(result.Volumes), len(original.Volumes))
	}
	if len(result.Ulimits) != len(original.Ulimits) {
		t.Errorf("Ulimits len = %d, want %d", len(result.Ulimits), len(original.Ulimits))
	}
	if len(result.HostAliases) != len(original.HostAliases) {
		t.Errorf("HostAliases len = %d, want %d", len(result.HostAliases), len(original.HostAliases))
	}

	// Timeout round-trip: "120s" → time.Duration → "2m0s"
	if result.Timeout == "" {
		t.Error("Timeout should not be empty after round-trip")
	}
}

// TestSplitFilters tests the splitFilters helper.
func TestSplitFilters(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"server:0", 1},
		{"server:0,agent:*", 2},
		{"a,b,c,d", 4},
	}

	for _, tt := range tests {
		got := splitFilters(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitFilters(%q) len = %d, want %d", tt.input, len(got), tt.want)
		}
	}
}

// TestSplitFilters_Values verifies exact values from split.
func TestSplitFilters_Values(t *testing.T) {
	got := splitFilters("server:0,agent:*")
	if got[0] != "server:0" || got[1] != "agent:*" {
		t.Errorf("splitFilters = %v, want [server:0 agent:*]", got)
	}
}

// TestAdvancedRequestToSimpleConfig_K3sNodeLabels tests node label mapping.
func TestAdvancedRequestToSimpleConfig_K3sNodeLabels(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{
		Name:    "labels-test",
		Servers: 1,
		K3sNodeLabels: []dto.NodeFilter{
			{Value: "tier=frontend", NodeFilters: "agent:0,agent:1"},
		},
		RuntimeLabels: []dto.NodeFilter{
			{Value: "env=dev", NodeFilters: "server:0"},
		},
	}

	sc := advancedRequestToSimpleConfig(req)

	if len(sc.Options.K3sOptions.NodeLabels) != 1 {
		t.Fatalf("NodeLabels len = %d, want 1", len(sc.Options.K3sOptions.NodeLabels))
	}
	if sc.Options.K3sOptions.NodeLabels[0].Label != "tier=frontend" {
		t.Errorf("Label = %q, want %q", sc.Options.K3sOptions.NodeLabels[0].Label, "tier=frontend")
	}
	if len(sc.Options.K3sOptions.NodeLabels[0].NodeFilters) != 2 {
		t.Errorf("NodeFilters len = %d, want 2", len(sc.Options.K3sOptions.NodeLabels[0].NodeFilters))
	}
	if len(sc.Options.Runtime.Labels) != 1 {
		t.Fatalf("Runtime.Labels len = %d, want 1", len(sc.Options.Runtime.Labels))
	}
}

// TestSimpleConfigToAdvancedRequest_Volumes tests volume reverse mapping.
func TestSimpleConfigToAdvancedRequest_Volumes(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		Volumes: []confv1alpha5.VolumeWithNodeFilters{
			{Volume: "/host:/container", NodeFilters: []string{"server:0"}},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if len(req.Volumes) != 1 {
		t.Fatalf("Volumes len = %d, want 1", len(req.Volumes))
	}
	if req.Volumes[0].Value != "/host:/container" {
		t.Errorf("Volume = %q, want %q", req.Volumes[0].Value, "/host:/container")
	}
	if req.Volumes[0].NodeFilters != "server:0" {
		t.Errorf("NodeFilters = %q, want %q", req.Volumes[0].NodeFilters, "server:0")
	}
}

// TestSimpleConfigToAdvancedRequest_Env tests env reverse mapping.
func TestSimpleConfigToAdvancedRequest_Env(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		Env: []confv1alpha5.EnvVarWithNodeFilters{
			{EnvVar: "DEBUG=true", NodeFilters: []string{"agent:*"}},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if len(req.Env) != 1 {
		t.Fatalf("Env len = %d, want 1", len(req.Env))
	}
	if req.Env[0].Value != "DEBUG=true" {
		t.Errorf("Env = %q, want %q", req.Env[0].Value, "DEBUG=true")
	}
}

// TestSimpleConfigToAdvancedRequest_Files tests file reverse mapping.
func TestSimpleConfigToAdvancedRequest_Files(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		Files: []confv1alpha5.FileWithNodeFilters{
			{Source: "/a", Destination: "/b", Description: "test", NodeFilters: []string{"server:0"}},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if len(req.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(req.Files))
	}
	if req.Files[0].Source != "/a" {
		t.Errorf("Source = %q, want %q", req.Files[0].Source, "/a")
	}
	if req.Files[0].Destination != "/b" {
		t.Errorf("Destination = %q, want %q", req.Files[0].Destination, "/b")
	}
	if req.Files[0].NodeFilters != "server:0" {
		t.Errorf("NodeFilters = %q, want %q", req.Files[0].NodeFilters, "server:0")
	}
}

// TestSimpleConfigToAdvancedRequest_K3sArgs tests k3s args reverse mapping.
func TestSimpleConfigToAdvancedRequest_K3sArgs(t *testing.T) {
	sc := confv1alpha5.SimpleConfig{
		Options: confv1alpha5.SimpleConfigOptions{
			K3sOptions: confv1alpha5.SimpleConfigOptionsK3s{
				ExtraArgs: []confv1alpha5.K3sArgWithNodeFilters{
					{Arg: "--tls-san=10.0.0.1", NodeFilters: []string{"server:*"}},
				},
				NodeLabels: []confv1alpha5.LabelWithNodeFilters{
					{Label: "key=value", NodeFilters: []string{"agent:0"}},
				},
			},
			Runtime: confv1alpha5.SimpleConfigOptionsRuntime{
				Labels: []confv1alpha5.LabelWithNodeFilters{
					{Label: "runtime=docker", NodeFilters: []string{"server:0"}},
				},
			},
		},
	}

	req := simpleConfigToAdvancedRequest(sc)

	if len(req.K3sArgs) != 1 {
		t.Fatalf("K3sArgs len = %d, want 1", len(req.K3sArgs))
	}
	if req.K3sArgs[0].Value != "--tls-san=10.0.0.1" {
		t.Errorf("K3sArgs[0].Value = %q, want %q", req.K3sArgs[0].Value, "--tls-san=10.0.0.1")
	}
	if len(req.K3sNodeLabels) != 1 {
		t.Fatalf("K3sNodeLabels len = %d, want 1", len(req.K3sNodeLabels))
	}
	if len(req.RuntimeLabels) != 1 {
		t.Fatalf("RuntimeLabels len = %d, want 1", len(req.RuntimeLabels))
	}
}

// TestAdvancedRequestToSimpleConfig_WaitIsTrue verifies Wait is always set.
func TestAdvancedRequestToSimpleConfig_WaitIsTrue(t *testing.T) {
	sc := advancedRequestToSimpleConfig(dto.ClusterCreateAdvancedRequest{Name: "w", Servers: 1})
	if !sc.Options.K3dOptions.Wait {
		t.Error("Wait = false, want true")
	}
}

// TestAdvancedRequestToSimpleConfig_Empty tests with minimal fields.
func TestAdvancedRequestToSimpleConfig_Empty(t *testing.T) {
	req := dto.ClusterCreateAdvancedRequest{Name: "empty", Servers: 1}
	sc := advancedRequestToSimpleConfig(req)

	if sc.ObjectMeta.Name != "empty" {
		t.Errorf("Name = %q, want %q", sc.ObjectMeta.Name, "empty")
	}
	if len(sc.Ports) != 0 {
		t.Errorf("Ports should be empty, got %d", len(sc.Ports))
	}
	if len(sc.Volumes) != 0 {
		t.Errorf("Volumes should be empty, got %d", len(sc.Volumes))
	}
	if len(sc.Env) != 0 {
		t.Errorf("Env should be empty, got %d", len(sc.Env))
	}

	// Ensure filter fields produce nil NodeFilters
	if sc.ExposeAPI.Host != "" || sc.ExposeAPI.HostPort != "" {
		t.Error("ExposeAPI should be empty when not set")
	}

	// Ensure API is not set unless any APIPort/Host/HostIP is given.
	parts := strings.Join([]string{sc.ExposeAPI.Host, sc.ExposeAPI.HostIP, sc.ExposeAPI.HostPort}, "")
	if parts != "" {
		t.Errorf("ExposeAPI should be empty, got combined %q", parts)
	}
}
