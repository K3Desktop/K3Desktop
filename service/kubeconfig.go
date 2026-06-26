package service

import (
	"context"
	"os"
	"path/filepath"

	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	k3d "github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/k3desktop/k3desktop/dto"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeconfigService struct{}

// ExportKubeconfig merges cluster kubeconfig into ~/.kube/config and returns the path.
// Kept synchronous because the frontend needs the returned path; wrapped with StartOp
// so the operations store reflects in-flight state.
func (s *KubeconfigService) ExportKubeconfig(ctx context.Context, clusterName string) (path string, retErr error) {
	defer WithTarget(clusterName)()
	_, done := StartOp("kubeconfig.export", clusterName)
	defer func() { done(retErr) }()

	c, err := k3dclient.ClusterGet(ctx, GetRuntime(), &k3d.Cluster{Name: clusterName})
	if err != nil {
		return "", err
	}
	outPath, err := k3dclient.KubeconfigGetDefaultPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0750); err != nil {
		return "", err
	}
	return k3dclient.KubeconfigGetWrite(ctx, GetRuntime(), c, outPath, &k3dclient.WriteKubeConfigOptions{
		OverwriteExisting:    false,
		UpdateCurrentContext: true,
		UpdateExisting:       true,
	})
}

// GetKubeconfigYAML returns cluster kubeconfig as YAML string for display.
func (s *KubeconfigService) GetKubeconfigYAML(ctx context.Context, clusterName string) (string, error) {
	c, err := k3dclient.ClusterGet(ctx, GetRuntime(), &k3d.Cluster{Name: clusterName})
	if err != nil {
		return "", err
	}
	cfg, err := k3dclient.KubeconfigGet(ctx, GetRuntime(), c)
	if err != nil {
		return "", err
	}
	bw := &bytesWriter{}
	if err := k3dclient.KubeconfigWriteToStream(ctx, cfg, bw); err != nil {
		return "", err
	}
	return string(bw.data), nil
}

// ListContexts returns all contexts in ~/.kube/config.
func (s *KubeconfigService) ListContexts() ([]dto.KubeconfigContextDTO, error) {
	path, err := k3dclient.KubeconfigGetDefaultPath()
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []dto.KubeconfigContextDTO{}, nil
		}
		return nil, err
	}
	result := make([]dto.KubeconfigContextDTO, 0, len(cfg.Contexts))
	for name, ctx := range cfg.Contexts {
		result = append(result, dto.KubeconfigContextDTO{
			Name:    name,
			Cluster: ctx.Cluster,
			User:    ctx.AuthInfo,
			Current: name == cfg.CurrentContext,
		})
	}
	return result, nil
}

// SetCurrentContext sets the active context in ~/.kube/config.
func (s *KubeconfigService) SetCurrentContext(contextName string) (retErr error) {
	defer WithTarget(contextName)()
	_, done := StartOp("kubeconfig.setContext", contextName)
	defer func() { done(retErr) }()
	path, err := k3dclient.KubeconfigGetDefaultPath()
	if err != nil {
		return err
	}
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return err
	}
	cfg.CurrentContext = contextName
	return clientcmd.WriteToFile(*cfg, path)
}

// DeleteContext removes a context (and its cluster/user if exclusive) from ~/.kube/config.
func (s *KubeconfigService) DeleteContext(contextName string) (retErr error) {
	defer WithTarget(contextName)()
	_, done := StartOp("kubeconfig.deleteContext", contextName)
	defer func() { done(retErr) }()
	path, err := k3dclient.KubeconfigGetDefaultPath()
	if err != nil {
		return err
	}
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return err
	}
	ctx, exists := cfg.Contexts[contextName]
	if !exists {
		return nil
	}
	clusterName := ctx.Cluster
	userName := ctx.AuthInfo
	delete(cfg.Contexts, contextName)

	// Remove cluster if no other context references it
	clusterUsed := false
	for _, c := range cfg.Contexts {
		if c.Cluster == clusterName {
			clusterUsed = true
			break
		}
	}
	if !clusterUsed {
		delete(cfg.Clusters, clusterName)
	}

	// Remove user if no other context references it
	userUsed := false
	for _, c := range cfg.Contexts {
		if c.AuthInfo == userName {
			userUsed = true
			break
		}
	}
	if !userUsed {
		delete(cfg.AuthInfos, userName)
	}

	if cfg.CurrentContext == contextName {
		cfg.CurrentContext = ""
	}
	return clientcmd.WriteToFile(*cfg, path)
}

type bytesWriter struct{ data []byte }

func (b *bytesWriter) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}
