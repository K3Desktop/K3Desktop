package service

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	confv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	dockerrt "github.com/k3d-io/k3d/v5/pkg/runtimes/docker"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/k3desktop/k3desktop/dto"
	"sigs.k8s.io/yaml"
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

// ListExternalRegistries returns registry containers not managed by k3d that
// k3d can attach to a cluster via `registries.use`.
func (s *RegistryService) ListExternalRegistries(ctx context.Context) ([]dto.ExternalRegistryDTO, error) {
	GetRuntime()
	c, err := dockerrt.GetDockerClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	containers, err := c.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	return filterExternalRegistries(containers), nil
}

// filterExternalRegistries keeps non-k3d containers running a registry image
// with exactly one published port (the shape k3d.RegistryFromNode requires).
func filterExternalRegistries(containers []container.Summary) []dto.ExternalRegistryDTO {
	result := []dto.ExternalRegistryDTO{}
	for _, ct := range containers {
		if _, managed := ct.Labels[k3d.LabelRole]; managed || len(ct.Names) == 0 {
			continue
		}
		repo, _, _ := strings.Cut(ct.Image, ":")
		if !strings.Contains(path.Base(repo), "registry") {
			continue
		}
		var published []container.Port
		for _, p := range ct.Ports {
			if p.PublicPort != 0 && p.Type == "tcp" && !slices.ContainsFunc(published, func(q container.Port) bool { return q.PublicPort == p.PublicPort }) {
				published = append(published, p)
			}
		}
		if len(published) != 1 {
			continue
		}
		result = append(result, dto.ExternalRegistryDTO{
			Name:         strings.TrimPrefix(ct.Names[0], "/"),
			Image:        ct.Image,
			Port:         int(published[0].PublicPort),
			InternalPort: int(published[0].PrivatePort),
			State:        ct.State,
			Reason:       unattachableReason(ct.HostConfig.NetworkMode),
		})
	}
	return result
}

// unattachableReason explains why a container in the given network mode can't
// be connected to the cluster network ("" if it can). pasta/slirp4netns are the
// rootless Podman defaults, which have no joinable network.
func unattachableReason(mode string) string {
	if mode == "pasta" || mode == "slirp4netns" || mode == "host" || mode == "none" || strings.HasPrefix(mode, "container:") {
		return fmt.Sprintf("network mode %q cannot join other networks; recreate it with --network <bridge network>", mode)
	}
	return ""
}

// attachExternalRegistries moves non-k3d registry containers out of
// sc.Registries.Use (k3d refuses containers without its labels) and into
// registries.config mirrors. It returns the container names, which must be
// connected to the cluster network once it exists (see connectExternalRegistries).
func attachExternalRegistries(ctx context.Context, sc *confv1alpha5.SimpleConfig) ([]string, error) {
	if len(sc.Registries.Use) == 0 {
		return nil, nil
	}
	external, err := (&RegistryService{}).ListExternalRegistries(ctx)
	if err != nil {
		return nil, err
	}
	var names []string
	mirrors := map[string]any{}
	use := sc.Registries.Use[:0:0]
	for _, ref := range sc.Registries.Use {
		i := slices.IndexFunc(external, func(r dto.ExternalRegistryDTO) bool { return r.Name == ref })
		if i < 0 {
			use = append(use, ref)
			continue
		}
		r := external[i]
		if r.Reason != "" {
			return nil, fmt.Errorf("registry %s: %s", r.Name, r.Reason)
		}
		endpoint := map[string]any{"endpoint": []string{fmt.Sprintf("http://%s:%d", r.Name, r.InternalPort)}}
		mirrors[fmt.Sprintf("%s:%d", r.Name, r.InternalPort)] = endpoint
		mirrors[fmt.Sprintf("localhost:%d", r.Port)] = endpoint
		names = append(names, r.Name)
	}
	if len(names) == 0 {
		return nil, nil
	}
	sc.Registries.Use = use

	cfg := map[string]any{}
	if sc.Registries.Config != "" {
		raw := []byte(sc.Registries.Config)
		if !strings.Contains(sc.Registries.Config, "\n") { // single line = file path
			if raw, err = os.ReadFile(sc.Registries.Config); err != nil {
				return nil, fmt.Errorf("read registries config: %w", err)
			}
		}
		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			return nil, fmt.Errorf("parse registries config: %w", err)
		}
	}
	merged, _ := cfg["mirrors"].(map[string]any)
	if merged == nil {
		merged = map[string]any{}
	}
	maps.Copy(merged, mirrors)
	cfg["mirrors"] = merged
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	sc.Registries.Config = string(out)
	return names, nil
}

// connectExternalRegistries joins the given containers to the cluster network
// so nodes can resolve them by name.
func connectExternalRegistries(ctx context.Context, networkName string, names []string) {
	dc, err := dockerrt.GetDockerClient()
	if err != nil {
		slog.Warn("cannot connect registries to cluster network", "err", err)
		return
	}
	defer dc.Close()
	for _, name := range names {
		if err := dc.NetworkConnect(ctx, networkName, name, nil); err != nil && !strings.Contains(err.Error(), "already exists") {
			slog.Warn(fmt.Sprintf("failed to connect registry %s to cluster network %s: %v", name, networkName, err))
		}
	}
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
