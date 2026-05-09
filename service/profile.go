package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	k3dconfigtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	confv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	k3dtypes "github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/k3desktop/k3desktop/dto"
	"github.com/wailsapp/wails/v3/pkg/application"
	"sigs.k8s.io/yaml"
)

type ProfileService struct{}

// profilesDirFunc is the directory resolver; overridable in tests.
var profilesDirFunc = profilesDirDefault

func profilesDir() (string, error) {
	return profilesDirFunc()
}

func profilesDirDefault() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	dir := filepath.Join(base, "k3desktop", "profiles")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create profiles dir: %w", err)
	}
	return dir, nil
}

func (s *ProfileService) ListProfiles(_ context.Context) ([]dto.ProfileDTO, error) {
	dir, err := profilesDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []dto.ProfileDTO
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		stem := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		out = append(out, dto.ProfileDTO{Name: stem, FileName: name})
	}
	return out, nil
}

func (s *ProfileService) GetProfile(_ context.Context, name string) (string, error) {
	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SaveProfile validates the YAML (must unmarshal to SimpleConfig) then writes it.
func (s *ProfileService) SaveProfile(_ context.Context, name string, content string) error {
	if name == "" {
		return fmt.Errorf("profile name required")
	}
	// validate
	var sc confv1alpha5.SimpleConfig
	if err := yaml.Unmarshal([]byte(content), &sc); err != nil {
		return fmt.Errorf("invalid k3d config YAML: %w", err)
	}
	dir, err := profilesDir()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".yaml"), []byte(content), 0o644)
}

func (s *ProfileService) DeleteProfile(_ context.Context, name string) error {
	dir, err := profilesDir()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(dir, name+".yaml"))
}

// ImportProfile opens a file picker, copies the chosen YAML into the profiles dir,
// and returns the stem name of the imported profile.
func (s *ProfileService) ImportProfile(_ context.Context) (string, error) {
	app := application.Get()
	paths, err := app.Dialog.OpenFile().
		SetTitle("Import k3d config").
		AddFilter("YAML files", "*.yaml;*.yml").
		PromptForMultipleSelection()
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", nil
	}
	srcPath := paths[0]

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", err
	}
	// validate
	var sc confv1alpha5.SimpleConfig
	if err := yaml.Unmarshal(data, &sc); err != nil {
		return "", fmt.Errorf("not a valid k3d config file: %w", err)
	}

	dir, err := profilesDir()
	if err != nil {
		return "", err
	}
	base := filepath.Base(srcPath)
	stem := strings.TrimSuffix(strings.TrimSuffix(base, ".yaml"), ".yml")
	dst := filepath.Join(dir, stem+".yaml")

	// avoid stomping existing profile — append suffix if needed
	if _, err := os.Stat(dst); err == nil {
		stem = stem + "-imported"
		dst = filepath.Join(dir, stem+".yaml")
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		return "", err
	}
	return stem, nil
}

// AdvancedRequestToYAML converts a ClusterCreateAdvancedRequest to a k3d v1alpha5 YAML string.
func (s *ProfileService) AdvancedRequestToYAML(_ context.Context, req dto.ClusterCreateAdvancedRequest) (string, error) {
	sc := advancedRequestToSimpleConfig(req)
	data, err := yaml.Marshal(sc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// YAMLToAdvancedRequest parses a k3d v1alpha5 YAML string into a ClusterCreateAdvancedRequest.
func (s *ProfileService) YAMLToAdvancedRequest(_ context.Context, content string) (dto.ClusterCreateAdvancedRequest, error) {
	var sc confv1alpha5.SimpleConfig
	if err := yaml.Unmarshal([]byte(content), &sc); err != nil {
		return dto.ClusterCreateAdvancedRequest{}, fmt.Errorf("invalid k3d config YAML: %w", err)
	}
	return simpleConfigToAdvancedRequest(sc), nil
}

// advancedRequestToSimpleConfig is the canonical mapping used by both ProfileService and ClusterService.
func advancedRequestToSimpleConfig(req dto.ClusterCreateAdvancedRequest) confv1alpha5.SimpleConfig {
	sc := confv1alpha5.SimpleConfig{
		TypeMeta: k3dconfigtypes.TypeMeta{
			APIVersion: confv1alpha5.ApiVersion,
			Kind:       "Simple",
		},
		ObjectMeta:   k3dconfigtypes.ObjectMeta{Name: req.Name},
		Servers:      req.Servers,
		Agents:       req.Agents,
		Image:        req.Image,
		Network:      req.Network,
		Subnet:       req.Subnet,
		ClusterToken: req.Token,
		Options: confv1alpha5.SimpleConfigOptions{
			Runtime: confv1alpha5.SimpleConfigOptionsRuntime{
				ServersMemory: req.ServersMemory,
				AgentsMemory:  req.AgentsMemory,
				GPURequest:    req.GPURequest,
				HostPidMode:   req.HostPidMode,
			},
			K3dOptions: confv1alpha5.SimpleConfigOptionsK3d{
				Wait:                true,
				DisableLoadbalancer: req.NoLoadbalancer,
				DisableImageVolume:  req.NoImageVolume,
				NoRollback:          req.NoRollback,
				Loadbalancer: confv1alpha5.SimpleConfigOptionsK3dLoadbalancer{
					ConfigOverrides: req.LBConfigOverrides,
				},
			},
			KubeconfigOptions: confv1alpha5.SimpleConfigOptionsKubeconfig{
				UpdateDefaultKubeconfig: req.UpdateKubeconfig,
				SwitchCurrentContext:    req.SwitchContext,
			},
		},
	}

	if req.Timeout != "" {
		if d, err := time.ParseDuration(req.Timeout); err == nil {
			sc.Options.K3dOptions.Timeout = d
		}
	}

	if req.APIPort != "" || req.APIHost != "" || req.APIHostIP != "" {
		sc.ExposeAPI = confv1alpha5.SimpleExposureOpts{
			Host:     req.APIHost,
			HostIP:   req.APIHostIP,
			HostPort: req.APIPort,
		}
	}

	for _, p := range req.Ports {
		sc.Ports = append(sc.Ports, confv1alpha5.PortWithNodeFilters{
			Port:        p.Value,
			NodeFilters: splitFilters(p.NodeFilters),
		})
	}
	for _, v := range req.Volumes {
		sc.Volumes = append(sc.Volumes, confv1alpha5.VolumeWithNodeFilters{
			Volume:      v.Value,
			NodeFilters: splitFilters(v.NodeFilters),
		})
	}
	for _, e := range req.Env {
		sc.Env = append(sc.Env, confv1alpha5.EnvVarWithNodeFilters{
			EnvVar:      e.Value,
			NodeFilters: splitFilters(e.NodeFilters),
		})
	}
	for _, a := range req.K3sArgs {
		sc.Options.K3sOptions.ExtraArgs = append(sc.Options.K3sOptions.ExtraArgs, confv1alpha5.K3sArgWithNodeFilters{
			Arg:         a.Value,
			NodeFilters: splitFilters(a.NodeFilters),
		})
	}
	for _, l := range req.K3sNodeLabels {
		sc.Options.K3sOptions.NodeLabels = append(sc.Options.K3sOptions.NodeLabels, confv1alpha5.LabelWithNodeFilters{
			Label:       l.Value,
			NodeFilters: splitFilters(l.NodeFilters),
		})
	}
	for _, l := range req.RuntimeLabels {
		sc.Options.Runtime.Labels = append(sc.Options.Runtime.Labels, confv1alpha5.LabelWithNodeFilters{
			Label:       l.Value,
			NodeFilters: splitFilters(l.NodeFilters),
		})
	}
	for _, u := range req.Ulimits {
		sc.Options.Runtime.Ulimits = append(sc.Options.Runtime.Ulimits, confv1alpha5.Ulimit{
			Name: u.Name,
			Soft: u.Soft,
			Hard: u.Hard,
		})
	}
	for _, f := range req.Files {
		sc.Files = append(sc.Files, confv1alpha5.FileWithNodeFilters{
			Source:      f.Source,
			Destination: f.Destination,
			Description: f.Description,
			NodeFilters: splitFilters(f.NodeFilters),
		})
	}
	for _, ha := range req.HostAliases {
		sc.HostAliases = append(sc.HostAliases, k3dtypes.HostAlias{
			IP:        ha.IP,
			Hostnames: ha.Hostnames,
		})
	}

	// Registry
	sc.Registries.Use = req.RegistryUse
	sc.Registries.Config = req.RegistryConfig
	if req.RegistryCreate != "" || req.RegistryCreateHost != "" || req.RegistryCreatePort != "" {
		parts := strings.SplitN(req.RegistryCreate, ":", 2)
		reg := &confv1alpha5.SimpleConfigRegistryCreateConfig{
			Host:     req.RegistryCreateHost,
			HostPort: req.RegistryCreatePort,
			Volumes:  req.RegistryVolumes,
		}
		if len(parts) > 0 && parts[0] != "" {
			reg.Name = parts[0]
		}
		if req.RegistryProxyURL != "" {
			reg.Proxy = k3dtypes.RegistryProxy{
				RemoteURL: req.RegistryProxyURL,
				Username:  req.RegistryProxyUser,
				Password:  req.RegistryProxyPass,
			}
		}
		sc.Registries.Create = reg
	}

	return sc
}

func simpleConfigToAdvancedRequest(sc confv1alpha5.SimpleConfig) dto.ClusterCreateAdvancedRequest {
	req := dto.ClusterCreateAdvancedRequest{
		Name:              sc.ObjectMeta.Name,
		Servers:           sc.Servers,
		Agents:            sc.Agents,
		Image:             sc.Image,
		Network:           sc.Network,
		Subnet:            sc.Subnet,
		Token:             sc.ClusterToken,
		APIPort:           sc.ExposeAPI.HostPort,
		APIHost:           sc.ExposeAPI.Host,
		APIHostIP:         sc.ExposeAPI.HostIP,
		ServersMemory:     sc.Options.Runtime.ServersMemory,
		AgentsMemory:      sc.Options.Runtime.AgentsMemory,
		GPURequest:        sc.Options.Runtime.GPURequest,
		HostPidMode:       sc.Options.Runtime.HostPidMode,
		NoLoadbalancer:    sc.Options.K3dOptions.DisableLoadbalancer,
		NoImageVolume:     sc.Options.K3dOptions.DisableImageVolume,
		NoRollback:        sc.Options.K3dOptions.NoRollback,
		UpdateKubeconfig:  sc.Options.KubeconfigOptions.UpdateDefaultKubeconfig,
		SwitchContext:     sc.Options.KubeconfigOptions.SwitchCurrentContext,
		RegistryConfig:    sc.Registries.Config,
		RegistryUse:       sc.Registries.Use,
		LBConfigOverrides: sc.Options.K3dOptions.Loadbalancer.ConfigOverrides,
	}

	if sc.Options.K3dOptions.Timeout != 0 {
		req.Timeout = sc.Options.K3dOptions.Timeout.String()
	}

	for _, p := range sc.Ports {
		req.Ports = append(req.Ports, dto.NodeFilter{Value: p.Port, NodeFilters: strings.Join(p.NodeFilters, ",")})
	}
	for _, v := range sc.Volumes {
		req.Volumes = append(req.Volumes, dto.NodeFilter{Value: v.Volume, NodeFilters: strings.Join(v.NodeFilters, ",")})
	}
	for _, e := range sc.Env {
		req.Env = append(req.Env, dto.NodeFilter{Value: e.EnvVar, NodeFilters: strings.Join(e.NodeFilters, ",")})
	}
	for _, a := range sc.Options.K3sOptions.ExtraArgs {
		req.K3sArgs = append(req.K3sArgs, dto.NodeFilter{Value: a.Arg, NodeFilters: strings.Join(a.NodeFilters, ",")})
	}
	for _, l := range sc.Options.K3sOptions.NodeLabels {
		req.K3sNodeLabels = append(req.K3sNodeLabels, dto.NodeFilter{Value: l.Label, NodeFilters: strings.Join(l.NodeFilters, ",")})
	}
	for _, l := range sc.Options.Runtime.Labels {
		req.RuntimeLabels = append(req.RuntimeLabels, dto.NodeFilter{Value: l.Label, NodeFilters: strings.Join(l.NodeFilters, ",")})
	}
	for _, u := range sc.Options.Runtime.Ulimits {
		req.Ulimits = append(req.Ulimits, dto.UlimitDTO{Name: u.Name, Soft: u.Soft, Hard: u.Hard})
	}
	for _, f := range sc.Files {
		req.Files = append(req.Files, dto.FileDTO{
			Source:      f.Source,
			Destination: f.Destination,
			Description: f.Description,
			NodeFilters: strings.Join(f.NodeFilters, ","),
		})
	}
	for _, ha := range sc.HostAliases {
		req.HostAliases = append(req.HostAliases, dto.HostAliasDTO{IP: ha.IP, Hostnames: ha.Hostnames})
	}

	if sc.Registries.Create != nil {
		reg := sc.Registries.Create
		req.RegistryCreate = reg.Name
		req.RegistryCreateHost = reg.Host
		req.RegistryCreatePort = reg.HostPort
		req.RegistryVolumes = reg.Volumes
		req.RegistryProxyURL = reg.Proxy.RemoteURL
		req.RegistryProxyUser = reg.Proxy.Username
		req.RegistryProxyPass = reg.Proxy.Password
	}

	return req
}
