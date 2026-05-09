package dto

import (
	"encoding/json"
	"testing"
)

func TestClusterDTO_JSON(t *testing.T) {
	d := ClusterDTO{
		Name:    "test",
		Servers: 3,
		Agents:  2,
		Status:  "running",
		Nodes: []NodeDTO{
			{Name: "server-0", Role: "server", State: "running", Image: "img:v1"},
		},
		Created: "2024-01-01",
	}

	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got ClusterDTO
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Name != "test" {
		t.Errorf("Name = %q, want %q", got.Name, "test")
	}
	if got.Servers != 3 {
		t.Errorf("Servers = %d, want 3", got.Servers)
	}
	if got.Agents != 2 {
		t.Errorf("Agents = %d, want 2", got.Agents)
	}
	if got.Status != "running" {
		t.Errorf("Status = %q, want %q", got.Status, "running")
	}
	if len(got.Nodes) != 1 {
		t.Fatalf("Nodes len = %d, want 1", len(got.Nodes))
	}
	if got.Created != "2024-01-01" {
		t.Errorf("Created = %q, want %q", got.Created, "2024-01-01")
	}
}

func TestNodeDTO_JSON(t *testing.T) {
	n := NodeDTO{Name: "s0", Role: "server", State: "running", Image: "img:v1"}
	data, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got NodeDTO
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != n {
		t.Errorf("roundtrip failed: got %+v, want %+v", got, n)
	}
}

func TestClusterCreateRequest_JSON(t *testing.T) {
	r := ClusterCreateRequest{
		Name:    "test",
		Servers: 1,
		Agents:  2,
		Image:   "rancher/k3s:v1.28.0",
		APIPort: "6443",
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got ClusterCreateRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != r {
		t.Errorf("roundtrip failed: got %+v, want %+v", got, r)
	}
}

func TestNodeFilter_JSON(t *testing.T) {
	nf := NodeFilter{Value: "8080:80", NodeFilters: "server:0,agent:*"}
	data, err := json.Marshal(nf)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got NodeFilter
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != nf {
		t.Errorf("got %+v, want %+v", got, nf)
	}
}

func TestClusterCreateAdvancedRequest_JSONBasic(t *testing.T) {
	r := ClusterCreateAdvancedRequest{
		Name:           "adv",
		Servers:        2,
		Agents:         3,
		Image:          "rancher/k3s:v1.28.0",
		APIPort:        "6443",
		NoLoadbalancer: true,
		Ports: []NodeFilter{
			{Value: "8080:80", NodeFilters: "server:0"},
		},
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got ClusterCreateAdvancedRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Name != "adv" {
		t.Errorf("Name = %q, want %q", got.Name, "adv")
	}
	if got.NoLoadbalancer != true {
		t.Error("NoLoadbalancer should be true")
	}
	if len(got.Ports) != 1 {
		t.Fatalf("Ports len = %d, want 1", len(got.Ports))
	}
}

func TestClusterDTO_JSONTags(t *testing.T) {
	d := ClusterDTO{Name: "test", Servers: 1, Status: "running"}
	data, _ := json.Marshal(d)
	m := map[string]interface{}{}
	json.Unmarshal(data, &m)

	requiredKeys := []string{"name", "servers", "agents", "status", "nodes", "created"}
	for _, k := range requiredKeys {
		if _, ok := m[k]; !ok {
			t.Errorf("missing JSON key %q", k)
		}
	}
}

func TestRegistryDTO_JSON(t *testing.T) {
	r := RegistryDTO{
		Name:     "k3d-myreg",
		Host:     "k3d-myreg",
		Port:     5000,
		Protocol: "http",
		State:    "running",
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got RegistryDTO
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got != r {
		t.Errorf("got %+v, want %+v", got, r)
	}
}

func TestRegistryCreateRequest_JSON(t *testing.T) {
	r := RegistryCreateRequest{Name: "myreg", Port: 5000}
	data, _ := json.Marshal(r)
	var got RegistryCreateRequest
	json.Unmarshal(data, &got)
	if got != r {
		t.Errorf("got %+v, want %+v", got, r)
	}
}

func TestKubeconfigContextDTO_JSON(t *testing.T) {
	k := KubeconfigContextDTO{
		Name:    "k3d-test",
		Cluster: "k3d-test",
		User:    "admin@k3d-test",
		Current: true,
	}
	data, _ := json.Marshal(k)
	var got KubeconfigContextDTO
	json.Unmarshal(data, &got)
	if got != k {
		t.Errorf("got %+v, want %+v", got, k)
	}
}

func TestLogEntryDTO_JSON(t *testing.T) {
	e := LogEntryDTO{
		Time:    "2024-01-01T00:00:00Z",
		Level:   "INFO",
		Message: "test message",
		Source:  "app",
		Target:  "my-cluster",
	}
	data, _ := json.Marshal(e)
	var got LogEntryDTO
	json.Unmarshal(data, &got)
	if got != e {
		t.Errorf("got %+v, want %+v", got, e)
	}
}

func TestBlueprintDTO_JSON(t *testing.T) {
	bp := BlueprintDTO{
		Name:        "test-bp",
		Description: "desc",
		FileName:    "test-bp.yaml",
		Charts: []ChartEntryDTO{
			{ReleaseName: "nginx", Repo: "https://example.com", Chart: "nginx", Version: "1.0.0", Values: "a: b"},
		},
	}
	data, err := json.Marshal(bp)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got BlueprintDTO
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Name != bp.Name {
		t.Errorf("Name = %q, want %q", got.Name, bp.Name)
	}
	if len(got.Charts) != 1 {
		t.Fatalf("Charts len = %d, want 1", len(got.Charts))
	}
	if got.Charts[0].ReleaseName != "nginx" {
		t.Errorf("Charts[0].ReleaseName = %q, want %q", got.Charts[0].ReleaseName, "nginx")
	}
}

func TestBlueprintDeployRequest_JSON(t *testing.T) {
	r := BlueprintDeployRequest{
		BlueprintName: "bp",
		ClusterName:   "cluster",
		Namespace:     "default",
	}
	data, _ := json.Marshal(r)
	var got BlueprintDeployRequest
	json.Unmarshal(data, &got)
	if got != r {
		t.Errorf("got %+v, want %+v", got, r)
	}
}

func TestBlueprintEventDTO_JSON(t *testing.T) {
	e := BlueprintEventDTO{Name: "bp", Message: "done"}
	data, _ := json.Marshal(e)
	var got BlueprintEventDTO
	json.Unmarshal(data, &got)
	if got != e {
		t.Errorf("got %+v, want %+v", got, e)
	}
}

func TestProfileDTO_JSON(t *testing.T) {
	p := ProfileDTO{Name: "dev", FileName: "dev.yaml"}
	data, _ := json.Marshal(p)
	var got ProfileDTO
	json.Unmarshal(data, &got)
	if got != p {
		t.Errorf("got %+v, want %+v", got, p)
	}
}

func TestUlimitDTO_JSON(t *testing.T) {
	u := UlimitDTO{Name: "nofile", Soft: 1024, Hard: 4096}
	data, _ := json.Marshal(u)
	var got UlimitDTO
	json.Unmarshal(data, &got)
	if got != u {
		t.Errorf("got %+v, want %+v", got, u)
	}
}

func TestFileDTO_JSON(t *testing.T) {
	f := FileDTO{Source: "/src", Destination: "/dst", Description: "test", NodeFilters: "server:0"}
	data, _ := json.Marshal(f)
	var got FileDTO
	json.Unmarshal(data, &got)
	if got != f {
		t.Errorf("got %+v, want %+v", got, f)
	}
}

func TestHostAliasDTO_JSON(t *testing.T) {
	h := HostAliasDTO{IP: "10.0.0.1", Hostnames: []string{"foo", "bar"}}
	data, _ := json.Marshal(h)
	var got HostAliasDTO
	json.Unmarshal(data, &got)
	if got.IP != h.IP {
		t.Errorf("IP = %q, want %q", got.IP, h.IP)
	}
	if len(got.Hostnames) != 2 {
		t.Fatalf("Hostnames len = %d, want 2", len(got.Hostnames))
	}
}

func TestClusterDTO_EmptyNodes(t *testing.T) {
	d := ClusterDTO{Name: "empty"}
	data, _ := json.Marshal(d)
	var got ClusterDTO
	json.Unmarshal(data, &got)
	if got.Nodes != nil {
		t.Error("nil nodes should marshal/unmarshal as null/nil")
	}
}

func TestClusterCreateAdvancedRequest_AllFields(t *testing.T) {
	r := ClusterCreateAdvancedRequest{
		Name:               "full",
		Servers:            3,
		Agents:             5,
		Image:              "img",
		APIPort:            "6443",
		APIHost:            "host",
		APIHostIP:          "0.0.0.0",
		Network:            "net",
		Subnet:             "10.0.0.0/16",
		Token:              "tok",
		ServersMemory:      "4g",
		AgentsMemory:       "2g",
		GPURequest:         "all",
		NoLoadbalancer:     true,
		NoImageVolume:      true,
		NoRollback:         true,
		Timeout:            "60s",
		HostPidMode:        true,
		UpdateKubeconfig:   true,
		SwitchContext:      true,
		RegistryCreate:     "reg",
		RegistryCreateHost: "0.0.0.0",
		RegistryCreatePort: "5000",
		RegistryProxyURL:   "https://example.com",
		RegistryProxyUser:  "u",
		RegistryProxyPass:  "p",
		RegistryConfig:     "mirrors: {}",
		LBConfigOverrides:  []string{"o1"},
		RegistryUse:        []string{"r1"},
		RegistryVolumes:    []string{"v1"},
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got ClusterCreateAdvancedRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.Name != "full" {
		t.Errorf("Name = %q, want %q", got.Name, "full")
	}
	if !got.NoLoadbalancer {
		t.Error("NoLoadbalancer should be true")
	}
	if !got.HostPidMode {
		t.Error("HostPidMode should be true")
	}
	if got.RegistryProxyURL != "https://example.com" {
		t.Errorf("RegistryProxyURL = %q, want %q", got.RegistryProxyURL, "https://example.com")
	}
}
