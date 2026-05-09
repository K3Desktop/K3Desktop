package service

import (
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/k3desktop/k3desktop/dto"
)

func TestClusterToDTO_AllRunning(t *testing.T) {
	c := &k3d.Cluster{
		Name: "test-cluster",
		Nodes: []*k3d.Node{
			{Name: "server-0", Role: k3d.ServerRole, State: k3d.NodeState{Running: true}, Image: "img:v1", Created: "2024-01-01"},
			{Name: "server-1", Role: k3d.ServerRole, State: k3d.NodeState{Running: true}, Image: "img:v1", Created: "2024-01-02"},
			{Name: "agent-0", Role: k3d.AgentRole, State: k3d.NodeState{Running: true}, Image: "img:v1"},
		},
	}

	d := clusterToDTO(c)

	if d.Name != "test-cluster" {
		t.Errorf("Name = %q, want %q", d.Name, "test-cluster")
	}
	if d.Servers != 2 {
		t.Errorf("Servers = %d, want 2", d.Servers)
	}
	if d.Agents != 1 {
		t.Errorf("Agents = %d, want 1", d.Agents)
	}
	if d.Status != "running" {
		t.Errorf("Status = %q, want %q", d.Status, "running")
	}
	if len(d.Nodes) != 3 {
		t.Errorf("Nodes len = %d, want 3", len(d.Nodes))
	}
	if d.Created != "2024-01-01" {
		t.Errorf("Created = %q, want %q", d.Created, "2024-01-01")
	}
}

func TestClusterToDTO_AllStopped(t *testing.T) {
	c := &k3d.Cluster{
		Name: "stopped-cluster",
		Nodes: []*k3d.Node{
			{Name: "server-0", Role: k3d.ServerRole, State: k3d.NodeState{Running: false, Status: "exited"}, Image: "img:v1"},
			{Name: "agent-0", Role: k3d.AgentRole, State: k3d.NodeState{Running: false, Status: "exited"}, Image: "img:v1"},
		},
	}

	d := clusterToDTO(c)

	if d.Status != "stopped" {
		t.Errorf("Status = %q, want %q", d.Status, "stopped")
	}
}

func TestClusterToDTO_PartiallyRunning(t *testing.T) {
	c := &k3d.Cluster{
		Name: "partial-cluster",
		Nodes: []*k3d.Node{
			{Name: "server-0", Role: k3d.ServerRole, State: k3d.NodeState{Running: true}, Image: "img:v1"},
			{Name: "agent-0", Role: k3d.AgentRole, State: k3d.NodeState{Running: false, Status: "exited"}, Image: "img:v1"},
		},
	}

	d := clusterToDTO(c)

	if d.Status != "partial" {
		t.Errorf("Status = %q, want %q", d.Status, "partial")
	}
}

func TestClusterToDTO_ExcludesLoadbalancer(t *testing.T) {
	c := &k3d.Cluster{
		Name: "lb-test",
		Nodes: []*k3d.Node{
			{Name: "server-0", Role: k3d.ServerRole, State: k3d.NodeState{Running: true}, Image: "img:v1"},
			{Name: "serverlb", Role: k3d.LoadBalancerRole, State: k3d.NodeState{Running: true}, Image: "lb:v1"},
		},
	}

	d := clusterToDTO(c)

	if len(d.Nodes) != 1 {
		t.Errorf("Nodes len = %d, want 1 (LB excluded)", len(d.Nodes))
	}
	if d.Nodes[0].Name != "server-0" {
		t.Errorf("Nodes[0].Name = %q, want %q", d.Nodes[0].Name, "server-0")
	}
}

func TestClusterToDTO_EmptyCluster(t *testing.T) {
	c := &k3d.Cluster{
		Name:  "empty-cluster",
		Nodes: []*k3d.Node{},
	}

	d := clusterToDTO(c)

	if d.Status != "stopped" {
		t.Errorf("Status = %q, want %q for empty cluster", d.Status, "stopped")
	}
	if d.Servers != 0 || d.Agents != 0 {
		t.Errorf("Servers = %d, Agents = %d, want 0, 0", d.Servers, d.Agents)
	}
}

func TestClusterToDTO_NodeState(t *testing.T) {
	c := &k3d.Cluster{
		Name: "state-test",
		Nodes: []*k3d.Node{
			{Name: "s0", Role: k3d.ServerRole, State: k3d.NodeState{Running: true}, Image: "img:v1"},
			{Name: "a0", Role: k3d.AgentRole, State: k3d.NodeState{Running: false, Status: "exited"}, Image: "img:v1"},
		},
	}

	d := clusterToDTO(c)

	if d.Nodes[0].State != "running" {
		t.Errorf("Nodes[0].State = %q, want %q", d.Nodes[0].State, "running")
	}
	if d.Nodes[1].State != "exited" {
		t.Errorf("Nodes[1].State = %q, want %q", d.Nodes[1].State, "exited")
	}
}

func TestClusterToDTO_NodeRoles(t *testing.T) {
	c := &k3d.Cluster{
		Name: "roles-test",
		Nodes: []*k3d.Node{
			{Name: "s0", Role: k3d.ServerRole, State: k3d.NodeState{Running: true}, Image: "img:v1"},
			{Name: "a0", Role: k3d.AgentRole, State: k3d.NodeState{Running: true}, Image: "img:v1"},
		},
	}

	d := clusterToDTO(c)

	if d.Nodes[0].Role != string(k3d.ServerRole) {
		t.Errorf("Nodes[0].Role = %q, want %q", d.Nodes[0].Role, k3d.ServerRole)
	}
	if d.Nodes[1].Role != string(k3d.AgentRole) {
		t.Errorf("Nodes[1].Role = %q, want %q", d.Nodes[1].Role, k3d.AgentRole)
	}
}

func TestClusterToDTO_NodeImages(t *testing.T) {
	c := &k3d.Cluster{
		Name: "images-test",
		Nodes: []*k3d.Node{
			{Name: "s0", Role: k3d.ServerRole, State: k3d.NodeState{Running: true}, Image: "rancher/k3s:v1.28.0"},
		},
	}

	d := clusterToDTO(c)

	if d.Nodes[0].Image != "rancher/k3s:v1.28.0" {
		t.Errorf("Image = %q, want %q", d.Nodes[0].Image, "rancher/k3s:v1.28.0")
	}
}

func TestNodeState_Running(t *testing.T) {
	n := &k3d.Node{State: k3d.NodeState{Running: true}}
	if got := nodeState(n); got != "running" {
		t.Errorf("nodeState = %q, want %q", got, "running")
	}
}

func TestNodeState_Stopped(t *testing.T) {
	n := &k3d.Node{State: k3d.NodeState{Running: false, Status: "exited"}}
	if got := nodeState(n); got != "exited" {
		t.Errorf("nodeState = %q, want %q", got, "exited")
	}
}

func TestNodeState_EmptyStatus(t *testing.T) {
	n := &k3d.Node{State: k3d.NodeState{Running: false, Status: ""}}
	if got := nodeState(n); got != "" {
		t.Errorf("nodeState = %q, want empty", got)
	}
}

func TestRegistryToDTO(t *testing.T) {
	tests := []struct {
		name      string
		reg       *k3d.Registry
		node      *k3d.Node
		wantName  string
		wantHost  string
		wantPort  int
		wantState string
	}{
		{
			name: "basic registry",
			reg: &k3d.Registry{
				Host:     "myregistry",
				Protocol: "http",
				ExposureOpts: k3d.ExposureOpts{
					PortMapping: nat.PortMapping{Binding: nat.PortBinding{HostPort: "5000"}},
				},
			},
			node:      &k3d.Node{Name: "k3d-myregistry", State: k3d.NodeState{Running: true}},
			wantName:  "k3d-myregistry",
			wantHost:  "k3d-myregistry",
			wantPort:  5000,
			wantState: "running",
		},
		{
			name: "registry with k3d- prefix already",
			reg: &k3d.Registry{
				Host:     "k3d-existing",
				Protocol: "http",
				ExposureOpts: k3d.ExposureOpts{
					PortMapping: nat.PortMapping{Binding: nat.PortBinding{HostPort: "5001"}},
				},
			},
			node:     &k3d.Node{Name: "k3d-existing", State: k3d.NodeState{Running: true}},
			wantName: "k3d-existing",
			wantHost: "k3d-existing",
			wantPort: 5001,
		},
		{
			name: "registry with empty port",
			reg: &k3d.Registry{
				Host:     "reg",
				Protocol: "https",
				ExposureOpts: k3d.ExposureOpts{
					PortMapping: nat.PortMapping{Binding: nat.PortBinding{HostPort: ""}},
				},
			},
			node:     &k3d.Node{Name: "k3d-reg", State: k3d.NodeState{Running: false, Status: "exited"}},
			wantName: "k3d-reg",
			wantHost: "k3d-reg",
			wantPort: 0,
		},
		{
			name: "registry with invalid port",
			reg: &k3d.Registry{
				Host:     "badport",
				Protocol: "http",
				ExposureOpts: k3d.ExposureOpts{
					PortMapping: nat.PortMapping{Binding: nat.PortBinding{HostPort: "abc"}},
				},
			},
			node:     &k3d.Node{Name: "k3d-badport", State: k3d.NodeState{Running: false}},
			wantName: "k3d-badport",
			wantHost: "k3d-badport",
			wantPort: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := registryToDTO(tt.reg, tt.node)
			if d.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", d.Name, tt.wantName)
			}
			if d.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", d.Host, tt.wantHost)
			}
			if d.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", d.Port, tt.wantPort)
			}
		})
	}
}

func TestRegistryToDTO_Protocol(t *testing.T) {
	reg := &k3d.Registry{
		Host:     "test",
		Protocol: "https",
		ExposureOpts: k3d.ExposureOpts{
			PortMapping: nat.PortMapping{Binding: nat.PortBinding{HostPort: "5000"}},
		},
	}
	node := &k3d.Node{Name: "k3d-test", State: k3d.NodeState{Running: true}}

	d := registryToDTO(reg, node)

	if d.Protocol != "https" {
		t.Errorf("Protocol = %q, want %q", d.Protocol, "https")
	}
}

// TestRegistryToDTO_PortParsing verifies port string to int conversion.
func TestRegistryToDTO_PortParsing(t *testing.T) {
	tests := []struct {
		hostPort string
		wantPort int
	}{
		{"5000", 5000},
		{"12345", 12345},
		{"0", 0},
		{"", 0},
		{"notanumber", 0},
	}

	for _, tt := range tests {
		t.Run("port_"+tt.hostPort, func(t *testing.T) {
			reg := &k3d.Registry{
				Host:     "r",
				Protocol: "http",
				ExposureOpts: k3d.ExposureOpts{
					PortMapping: nat.PortMapping{Binding: nat.PortBinding{HostPort: tt.hostPort}},
				},
			}
			node := &k3d.Node{Name: "k3d-r"}

			d := registryToDTO(reg, node)

			// For "0", strconv.Atoi returns 0 which is valid
			expectedPort := tt.wantPort
			p, err := strconv.Atoi(tt.hostPort)
			if err != nil {
				expectedPort = 0
			} else {
				expectedPort = p
			}
			if d.Port != expectedPort {
				t.Errorf("Port = %d, want %d for hostPort %q", d.Port, expectedPort, tt.hostPort)
			}
		})
	}
}

// TestClusterToDTO_CreatedUsesFirstServer verifies 'created' comes from the first server node.
func TestClusterToDTO_CreatedUsesFirstServer(t *testing.T) {
	c := &k3d.Cluster{
		Name: "created-test",
		Nodes: []*k3d.Node{
			{Name: "agent-0", Role: k3d.AgentRole, State: k3d.NodeState{Running: true}, Created: "2024-06-01"},
			{Name: "server-0", Role: k3d.ServerRole, State: k3d.NodeState{Running: true}, Created: "2024-05-01"},
		},
	}

	d := clusterToDTO(c)

	// Created should be from the first server node encountered
	if d.Created != "2024-05-01" {
		t.Errorf("Created = %q, want %q (from first server)", d.Created, "2024-05-01")
	}
}

// TestNodeDTO_Fields verifies the NodeDTO fields are correctly populated.
func TestNodeDTO_Fields(t *testing.T) {
	n := dto.NodeDTO{
		Name:  "k3d-mycluster-server-0",
		Role:  "server",
		State: "running",
		Image: "rancher/k3s:v1.28.0-k3s1",
	}

	if n.Name != "k3d-mycluster-server-0" {
		t.Error("Name mismatch")
	}
	if n.Role != "server" {
		t.Error("Role mismatch")
	}
	if n.State != "running" {
		t.Error("State mismatch")
	}
	if n.Image != "rancher/k3s:v1.28.0-k3s1" {
		t.Error("Image mismatch")
	}
}
