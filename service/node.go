package service

import (
	"context"
	"fmt"

	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
	k3dversion "github.com/k3d-io/k3d/v5/version"
	"github.com/k3desktop/k3desktop/dto"
)

type NodeService struct{}

func (s *NodeService) ListNodes(ctx context.Context, clusterName string) ([]dto.NodeDTO, error) {
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
