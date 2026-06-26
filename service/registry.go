package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/docker/go-connections/nat"
	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/k3desktop/k3desktop/dto"
)

type RegistryService struct{}

func (s *RegistryService) ListRegistries(ctx context.Context) ([]dto.RegistryDTO, error) {
	nodes, err := k3dclient.NodeList(ctx, GetRuntime())
	if err != nil {
		return nil, err
	}
	result := []dto.RegistryDTO{}
	for _, n := range nodes {
		if n.Role != k3d.RegistryRole {
			continue
		}
		reg, err := k3dclient.RegistryFromNode(n)
		if err != nil {
			continue
		}
		result = append(result, registryToDTO(reg, n))
	}
	return result, nil
}

func (s *RegistryService) CreateRegistry(_ context.Context, req dto.RegistryCreateRequest) (string, error) {
	if req.Name == "" {
		return "", fmt.Errorf("registry name required")
	}
	id, done := StartOp("registry.create", req.Name)
	go func() {
		defer WithTarget(req.Name)()
		var err error
		defer func() { done(err) }()
		reg := &k3d.Registry{
			Host:     req.Name,
			Image:    fmt.Sprintf("%s:%s", k3d.DefaultRegistryImageRepo, k3d.DefaultRegistryImageTag),
			Protocol: "http",
			ExposureOpts: k3d.ExposureOpts{
				Host: "0.0.0.0",
			},
		}
		reg.ExposureOpts.Port = nat.Port("5000/tcp")
		if req.Port > 0 {
			reg.ExposureOpts.Binding.HostPort = strconv.Itoa(req.Port)
		}
		// req.Port == 0: leave HostPort as "" so Docker assigns a random port
		_, err = k3dclient.RegistryRun(context.Background(), GetRuntime(), reg)
	}()
	return id, nil
}

func (s *RegistryService) DeleteRegistry(_ context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("registry name required")
	}
	id, done := StartOp("registry.delete", name)
	go func() {
		defer WithTarget(name)()
		var err error
		defer func() { done(err) }()
		ctx := context.Background()
		node, getErr := k3dclient.NodeGet(ctx, GetRuntime(), &k3d.Node{Name: name})
		if getErr != nil {
			node, getErr = k3dclient.NodeGet(ctx, GetRuntime(), &k3d.Node{Name: "k3d-" + name})
			if getErr != nil {
				err = getErr
				return
			}
		}
		err = k3dclient.NodeDelete(ctx, GetRuntime(), node, k3d.NodeDeleteOpts{})
	}()
	return id, nil
}

func registryToDTO(reg *k3d.Registry, n *k3d.Node) dto.RegistryDTO {
	port := 0
	if reg.ExposureOpts.Binding.HostPort != "" {
		p, _ := strconv.Atoi(reg.ExposureOpts.Binding.HostPort)
		port = p
	}
	host := reg.Host
	if !strings.HasPrefix(host, "k3d-") {
		host = "k3d-" + host
	}
	return dto.RegistryDTO{
		Name:     n.Name,
		Host:     host,
		Port:     port,
		Protocol: reg.Protocol,
		State:    nodeState(n),
	}
}
