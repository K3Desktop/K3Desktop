package service

import (
	"context"
	"fmt"
	"time"

	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
	k3dversion "github.com/k3d-io/k3d/v5/version"
	"github.com/k3desktop/k3desktop/dto"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type NodeService struct{}

func (s *NodeService) ListNodes(ctx context.Context, clusterName string) ([]dto.NodeDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	c, err := k3dclient.ClusterGet(ctx, GetRuntime(), &k3d.Cluster{Name: clusterName})
	if err != nil {
		return nil, err
	}
	result := make([]dto.NodeDTO, 0, len(c.Nodes))
	for _, n := range c.Nodes {
		if n.Role == k3d.LoadBalancerRole {
			continue
		}
		result = append(result, dto.NodeDTO{
			Name:  n.Name,
			Role:  string(n.Role),
			State: nodeState(n),
			Image: n.Image,
		})
	}
	return result, nil
}

func (s *NodeService) AddAgent(ctx context.Context, clusterName string) error {
	defer WithTarget(clusterName)()
	c, err := k3dclient.ClusterGet(ctx, GetRuntime(), &k3d.Cluster{Name: clusterName})
	if err != nil {
		return err
	}
	agentCount := 0
	for _, n := range c.Nodes {
		if n.Role == k3d.AgentRole {
			agentCount++
		}
	}
	nodeName := k3dclient.GenerateNodeName(clusterName, k3d.AgentRole, agentCount)
	node := &k3d.Node{
		Name:  nodeName,
		Role:  k3d.AgentRole,
		Image: fmt.Sprintf("%s:%s", k3d.DefaultK3sImageRepo, k3dversion.K3sVersion),
	}
	return k3dclient.NodeAddToCluster(ctx, GetRuntime(), node, c, k3d.NodeCreateOpts{})
}

func (s *NodeService) DeleteNode(ctx context.Context, nodeName string) error {
	defer WithTarget(nodeName)()
	node, err := k3dclient.NodeGet(ctx, GetRuntime(), &k3d.Node{Name: nodeName})
	if err != nil {
		return err
	}
	if node.Role == k3d.ServerRole {
		return fmt.Errorf("cannot delete server node; use DeleteCluster")
	}
	return k3dclient.NodeDelete(ctx, GetRuntime(), node, k3d.NodeDeleteOpts{})
}

func (s *NodeService) StartNode(ctx context.Context, nodeName string) error {
	defer WithTarget(nodeName)()
	node, err := k3dclient.NodeGet(ctx, GetRuntime(), &k3d.Node{Name: nodeName})
	if err != nil {
		return err
	}
	cluster, err := k3dclient.ClusterGet(ctx, GetRuntime(), &k3d.Cluster{Name: node.RuntimeLabels[k3d.LabelClusterName]})
	if err != nil {
		return err
	}
	envInfo, err := k3dclient.GatherEnvironmentInfo(ctx, GetRuntime(), cluster)
	if err != nil {
		return fmt.Errorf("gather environment info: %w", err)
	}
	return k3dclient.NodeStart(ctx, GetRuntime(), node, &k3d.NodeStartOpts{
		EnvironmentInfo: envInfo,
	})
}

func (s *NodeService) StopNode(ctx context.Context, nodeName string) error {
	defer WithTarget(nodeName)()
	node, err := k3dclient.NodeGet(ctx, GetRuntime(), &k3d.Node{Name: nodeName})
	if err != nil {
		return err
	}
	return GetRuntime().StopNode(ctx, node)
}

func (s *NodeService) UpgradeNode(ctx context.Context, nodeName, image string) error {
	defer WithTarget(nodeName)()
	rt := GetRuntime()

	oldNode, err := k3dclient.NodeGet(ctx, rt, &k3d.Node{Name: nodeName})
	if err != nil {
		return fmt.Errorf("node not found: %w", err)
	}

	cluster, err := k3dclient.ClusterGet(ctx, rt, &k3d.Cluster{Name: oldNode.RuntimeLabels[k3d.LabelClusterName]})
	if err != nil {
		return fmt.Errorf("cluster not found: %w", err)
	}

	// Rename the old node to free the original name for the replacement.
	tempName := nodeName + "-upgrading"
	if err := rt.RenameNode(ctx, oldNode, tempName); err != nil {
		return fmt.Errorf("rename old node: %w", err)
	}
	oldNode.Name = tempName

	// Re-fetch cluster after rename: cluster.Nodes still holds the pre-rename name,
	// so NodeAddToCluster's NodeGet call would fail to find the source container.
	cluster, err = k3dclient.ClusterGet(ctx, rt, &k3d.Cluster{Name: cluster.Name})
	if err != nil {
		_ = rt.RenameNode(ctx, oldNode, nodeName)
		return fmt.Errorf("re-fetch cluster after rename: %w", err)
	}

	// k3s stores a per-node password hash as a Kubernetes secret
	// (<nodename>.node-password.k3s in kube-system). A new container with the same
	// hostname generates a fresh password that won't match the stored hash, causing
	// the server to permanently reject registration. Delete the stale secret first.
	if err := cleanStaleNodeState(ctx, cluster, nodeName); err != nil {
		_ = rt.RenameNode(ctx, oldNode, nodeName)
		return fmt.Errorf("clean stale node state: %w", err)
	}

	// NodeAddToCluster performs the full node setup: it reads env vars (K3S_URL,
	// K3S_TOKEN), registry config and network settings from existing cluster nodes,
	// then creates and starts the new container and waits for the ready log message.
	// CopyNode omits all of that, causing the node to start but never join the cluster.
	newNode := &k3d.Node{
		Name:  nodeName,
		Role:  oldNode.Role,
		Image: image,
	}
	if err := k3dclient.NodeAddToCluster(ctx, rt, newNode, cluster, k3d.NodeCreateOpts{Wait: true}); err != nil {
		// rollback: restore old node name so the cluster stays functional
		_ = rt.RenameNode(ctx, oldNode, nodeName)
		return fmt.Errorf("create replacement node (rolled back): %w", err)
	}

	if err := rt.StopNode(ctx, oldNode); err != nil {
		return fmt.Errorf("stop old node: %w", err)
	}

	return k3dclient.NodeDelete(ctx, rt, oldNode, k3d.NodeDeleteOpts{SkipLBUpdate: true})
}

// cleanStaleNodeState removes k8s objects that would block a replacement container
// from joining cleanly:
//  1. The node-password secret — new container generates a fresh password; stale hash
//     causes permanent 403 from the server.
//  2. The Node object itself — it carries stale flannel/CNI annotations (e.g.
//     flannel.alpha.coreos.com/public-ip pointing to the old container's IP). flannel
//     in the new container reads those stale annotations, tries to find an interface
//     that doesn't exist, and crashes before kubelet can post status.
func cleanStaleNodeState(ctx context.Context, cluster *k3d.Cluster, nodeName string) error {
	kubecfg, err := k3dclient.KubeconfigGet(ctx, GetRuntime(), cluster)
	if err != nil {
		return fmt.Errorf("get kubeconfig: %w", err)
	}
	restCfg, err := clientcmd.NewDefaultClientConfig(*kubecfg, nil).ClientConfig()
	if err != nil {
		return fmt.Errorf("build rest config: %w", err)
	}
	kc, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("build k8s client: %w", err)
	}

	secretName := nodeName + ".node-password.k3s"
	err = kc.CoreV1().Secrets("kube-system").Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("delete node-password secret: %w", err)
	}

	err = kc.CoreV1().Nodes().Delete(ctx, nodeName, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return fmt.Errorf("delete node object: %w", err)
	}

	return nil
}
