package service

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
)

func TestFilterExternalRegistries(t *testing.T) {
	port := func(pub uint16) container.Port { return container.Port{PrivatePort: 5000, PublicPort: pub, Type: "tcp"} }
	in := []container.Summary{
		{Names: []string{"/registry.local"}, Image: "registry:2", State: "running", Ports: []container.Port{port(5000)}},
		{Names: []string{"/ipv4v6"}, Image: "docker.io/library/registry:2", State: "exited", Ports: []container.Port{port(5001), {PrivatePort: 5000, PublicPort: 5001, Type: "tcp", IP: "::"}}},
		{Names: []string{"/k3d-reg"}, Image: "registry:2", Labels: map[string]string{k3d.LabelRole: "registry"}, Ports: []container.Port{port(5002)}},
		{Names: []string{"/unpublished"}, Image: "registry:2", Ports: []container.Port{{PrivatePort: 5000, Type: "tcp"}}},
		{Names: []string{"/nginx"}, Image: "nginx:latest", Ports: []container.Port{port(8080)}},
	}
	in = append(in, container.Summary{
		Names: []string{"/pasta"}, Image: "registry:2", Ports: []container.Port{port(5003)},
		HostConfig: struct {
			NetworkMode string            `json:",omitempty"`
			Annotations map[string]string `json:",omitempty"`
		}{NetworkMode: "pasta"},
	})
	got := filterExternalRegistries(in)
	if last := got[len(got)-1]; last.Name != "pasta" || last.Reason == "" {
		t.Errorf("pasta registry should be flagged unattachable: %+v", last)
	}
	got = got[:len(got)-1]
	if len(got) != 2 {
		t.Fatalf("want 2 registries, got %+v", got)
	}
	if got[0].Name != "registry.local" || got[0].Port != 5000 || got[0].State != "running" {
		t.Errorf("unexpected first entry: %+v", got[0])
	}
	if got[1].Name != "ipv4v6" || got[1].Port != 5001 {
		t.Errorf("unexpected second entry: %+v", got[1])
	}
}
