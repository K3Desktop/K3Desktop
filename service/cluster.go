package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	k3dconfig "github.com/k3d-io/k3d/v5/pkg/config"
	k3dconfigtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	confv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
	k3dversion "github.com/k3d-io/k3d/v5/version"
	"github.com/k3desktop/k3desktop/dto"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	application.RegisterEvent[string]("cluster:creating")
	application.RegisterEvent[string]("cluster:ready")
	application.RegisterEvent[string]("cluster:error")
}

type ClusterService struct{}

func (s *ClusterService) ListClusters(ctx context.Context) ([]dto.ClusterDTO, error) {
	clusters, err := k3dclient.ClusterList(ctx, GetRuntime())
	if err != nil {
		return nil, err
	}
	result := make([]dto.ClusterDTO, 0, len(clusters))
	for _, c := range clusters {
		result = append(result, clusterToDTO(c))
	}
	return result, nil
}

func (s *ClusterService) CreateCluster(ctx context.Context, req dto.ClusterCreateRequest) (string, error) {
	if req.Name == "" {
		return "", fmt.Errorf("cluster name required")
	}
	if req.Servers < 1 {
		req.Servers = 1
	}
	if req.Image == "" {
		req.Image = fmt.Sprintf("%s:%s", k3d.DefaultK3sImageRepo, k3dversion.K3sVersion)
	}

	simpleConfig := confv1alpha5.SimpleConfig{
		ObjectMeta: k3dconfigtypes.ObjectMeta{Name: req.Name},
		Servers:    req.Servers,
		Agents:     req.Agents,
		Image:      req.Image,
		Options: confv1alpha5.SimpleConfigOptions{
			KubeconfigOptions: confv1alpha5.SimpleConfigOptionsKubeconfig{
				UpdateDefaultKubeconfig: true,
				SwitchCurrentContext:    true,
			},
		},
	}
	if req.APIPort != "" {
		simpleConfig.ExposeAPI = confv1alpha5.SimpleExposureOpts{HostPort: req.APIPort}
	}

	return runCluster(ctx, simpleConfig)
}

func (s *ClusterService) CreateClusterAdvanced(ctx context.Context, req dto.ClusterCreateAdvancedRequest) (string, error) {
	if req.Name == "" {
		return "", fmt.Errorf("cluster name required")
	}
	if req.Servers < 1 {
		req.Servers = 1
	}
	if req.Image == "" {
		req.Image = fmt.Sprintf("%s:%s", k3d.DefaultK3sImageRepo, k3dversion.K3sVersion)
	}
	return runCluster(ctx, advancedRequestToSimpleConfig(req))
}

func runCluster(ctx context.Context, sc confv1alpha5.SimpleConfig) (string, error) {
	clusterConfig, err := k3dconfig.TransformSimpleToClusterConfig(ctx, GetRuntime(), sc, "")
	if err != nil {
		return "", fmt.Errorf("config transform: %w", err)
	}
	clusterConfig, err = k3dconfig.ProcessClusterConfig(*clusterConfig)
	if err != nil {
		return "", fmt.Errorf("config process: %w", err)
	}

	kubeconfigOpts := clusterConfig.KubeconfigOpts
	id, done := StartOp("cluster.create", sc.Name)
	go func() {
		defer WithTarget(sc.Name)()
		var opErr error
		defer func() { done(opErr) }()

		app := application.Get()
		if app != nil {
			app.Event.Emit("cluster:creating", sc.Name)
		}
		if err := k3dclient.ClusterRun(context.Background(), GetRuntime(), clusterConfig); err != nil {
			slog.Error("cluster creation failed, attempting rollback", "cluster", sc.Name, "err", err)
			if sc.Options.K3dOptions.NoRollback {
				opErr = fmt.Errorf("cluster creation failed (rollback disabled): %w", err)
				if app != nil {
					app.Event.Emit("cluster:error", opErr.Error())
				}
				return
			}
			if rbErr := k3dclient.ClusterDelete(context.Background(), GetRuntime(), &clusterConfig.Cluster, k3d.ClusterDeleteOpts{SkipRegistryCheck: true}); rbErr != nil {
				slog.Error("rollback failed", "cluster", sc.Name, "err", rbErr)
				opErr = fmt.Errorf("cluster creation failed and rollback also failed: %v (rollback: %v)", err, rbErr)
				if app != nil {
					app.Event.Emit("cluster:error", opErr.Error())
				}
				return
			}
			slog.Info("rollback completed", "cluster", sc.Name)
			opErr = fmt.Errorf("cluster creation failed, changes rolled back: %w", err)
			if app != nil {
				app.Event.Emit("cluster:error", opErr.Error())
			}
			return
		}
		if kubeconfigOpts.UpdateDefaultKubeconfig {
			if _, err := k3dclient.KubeconfigGetWrite(
				context.Background(),
				GetRuntime(),
				&clusterConfig.Cluster,
				"",
				&k3dclient.WriteKubeConfigOptions{
					UpdateExisting:       true,
					OverwriteExisting:    false,
					UpdateCurrentContext: kubeconfigOpts.SwitchCurrentContext,
				},
			); err != nil {
				slog.Warn("kubeconfig merge failed", "cluster", sc.Name, "err", err)
			}
		}
		if app != nil {
			app.Event.Emit("cluster:ready", sc.Name)
		}
	}()
	return id, nil
}

func splitFilters(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func (s *ClusterService) DeleteCluster(_ context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("cluster name required")
	}
	id, done := StartOp("cluster.delete", name)
	go func() {
		defer WithTarget(name)()
		var err error
		defer func() { done(err) }()
		err = deleteClusterSync(name)
	}()
	return id, nil
}

func deleteClusterSync(name string) error {
	ctx := context.Background()
	c, err := k3dclient.ClusterGet(ctx, GetRuntime(), &k3d.Cluster{Name: name})
	if err != nil {
		return err
	}

	// Disconnect registries from cluster network before deletion to avoid "network has active endpoints" error
	nodes, _ := k3dclient.NodeList(ctx, GetRuntime())
	for _, n := range nodes {
		if n.Role == k3d.RegistryRole {
			for _, net := range n.Networks {
				if net == c.Network.Name {
					slog.Info("disconnecting registry from cluster network", "registry", n.Name, "network", c.Network.Name)
					if err := GetRuntime().DisconnectNodeFromNetwork(ctx, n, c.Network.Name); err != nil {
						slog.Warn("failed to disconnect registry from network", "registry", n.Name, "network", c.Network.Name, "err", err)
					}
					break
				}
			}
		}
	}

	return k3dclient.ClusterDelete(ctx, GetRuntime(), c, k3d.ClusterDeleteOpts{SkipRegistryCheck: false})
}

func (s *ClusterService) StartCluster(_ context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("cluster name required")
	}
	id, done := StartOp("cluster.start", name)
	go func() {
		defer WithTarget(name)()
		var err error
		defer func() { done(err) }()
		err = startClusterSync(name)
	}()
	return id, nil
}

func startClusterSync(name string) error {
	ctx := context.Background()
	rt := GetRuntime()

	slog.Info("starting cluster", "cluster", name)

	c, err := k3dclient.ClusterGet(ctx, rt, &k3d.Cluster{Name: name})
	if err != nil {
		return err
	}

	slog.Debug("gathering environment info for cluster start", "cluster", name)

	envInfo, err := k3dclient.GatherEnvironmentInfo(ctx, rt, c)
	if err != nil {
		return fmt.Errorf("gather environment info: %w", err)
	}

	slog.Debug("starting server nodes", "cluster", name)
	// Start the server nodes first and wait for them to be ready, so the apiserver
	// is reachable before we try to clean stale agent state.
	for _, node := range c.Nodes {
		if node.State.Running || node.Role != k3d.ServerRole {
			continue
		}
		if err := k3dclient.NodeStart(ctx, rt, node, &k3d.NodeStartOpts{
			Wait:            true,
			EnvironmentInfo: envInfo,
		}); err != nil {
			return fmt.Errorf("start server node %s: %w", node.Name, err)
		}
	}

	slog.Debug("re-fetch cluster to gather its state")
	// Re-fetch cluster so we have the updated server node state (running=true, fresh IPs).
	c, err = k3dclient.ClusterGet(ctx, rt, &k3d.Cluster{Name: name})
	if err != nil {
		return fmt.Errorf("re-fetch cluster after server start: %w", err)
	}

	// After a cluster restart Docker assigns new IPs to containers. Agent Node objects
	// in Kubernetes still carry the old IPs in flannel annotations, causing flannel to
	// fail with "failed to find interface with specified node ip" and shut down k3s.
	// Delete the stale Node objects so agents rejoin cleanly with their new IPs.

	slog.Debug("cleaning stale agent node state", "cluster", name)
	for _, node := range c.Nodes {
		if node.Role != k3d.AgentRole {
			continue
		}
		if err := cleanStaleNodeState(ctx, c, node.Name); err != nil {
			// Non-fatal: log and continue. If the node object is missing or apiserver
			// not yet ready, k3s will still attempt to start; the error will surface there.
			slog.Warn("clean stale agent node state on cluster start", "node", node.Name, "err", err)
		}
	}

	slog.Debug("starting agent nodes", "cluster", name)
	return k3dclient.ClusterStart(ctx, rt, c, k3d.ClusterStartOpts{
		EnvironmentInfo: envInfo,
	})
}

func (s *ClusterService) StopCluster(_ context.Context, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("cluster name required")
	}
	id, done := StartOp("cluster.stop", name)
	go func() {
		defer WithTarget(name)()
		var err error
		defer func() { done(err) }()
		ctx := context.Background()
		c, getErr := k3dclient.ClusterGet(ctx, GetRuntime(), &k3d.Cluster{Name: name})
		if getErr != nil {
			err = getErr
			return
		}
		err = k3dclient.ClusterStop(ctx, GetRuntime(), c)
	}()
	return id, nil
}

func clusterToDTO(c *k3d.Cluster) dto.ClusterDTO {
	nodes := make([]dto.NodeDTO, 0, len(c.Nodes))
	servers, agents := 0, 0
	running, total := 0, 0
	created := ""
	for _, n := range c.Nodes {
		if n.Role == k3d.LoadBalancerRole {
			continue
		}
		total++
		if n.State.Running {
			running++
		}
		switch n.Role {
		case k3d.ServerRole:
			servers++
			if created == "" {
				created = n.Created
			}
		case k3d.AgentRole:
			agents++
		}
		nodes = append(nodes, dto.NodeDTO{
			Name:  n.Name,
			Role:  string(n.Role),
			State: nodeState(n),
			Image: n.Image,
		})
	}

	status := "stopped"
	if running == total && total > 0 {
		status = "running"
	} else if running > 0 {
		status = "partial"
	}

	return dto.ClusterDTO{
		Name:    c.Name,
		Servers: servers,
		Agents:  agents,
		Status:  status,
		Nodes:   nodes,
		Created: created,
	}
}

func nodeState(n *k3d.Node) string {
	if n.State.Running {
		return "running"
	}
	return n.State.Status
}
