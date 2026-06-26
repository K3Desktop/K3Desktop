package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/k3desktop/k3desktop/dto"
	"github.com/wailsapp/wails/v3/pkg/application"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/storage/driver"
	"sigs.k8s.io/yaml"
)

func init() {
	application.RegisterEvent[dto.BlueprintEventDTO]("blueprint:deploying")
	application.RegisterEvent[dto.BlueprintEventDTO]("blueprint:done")
	application.RegisterEvent[dto.BlueprintEventDTO]("blueprint:error")
}

type BlueprintService struct{}

// blueprintsDirFunc is the directory resolver; overridable in tests.
var blueprintsDirFunc = blueprintsDirDefault

func blueprintsDir() (string, error) {
	return blueprintsDirFunc()
}

func blueprintsDirDefault() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	dir := filepath.Join(base, "k3desktop", "blueprints")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create blueprints dir: %w", err)
	}
	return dir, nil
}

func (s *BlueprintService) ListBlueprints(_ context.Context) ([]dto.BlueprintDTO, error) {
	dir, err := blueprintsDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []dto.BlueprintDTO
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		stem := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		var bp dto.BlueprintDTO
		if err := yaml.Unmarshal(data, &bp); err != nil {
			continue
		}
		bp.FileName = name
		if bp.Name == "" {
			bp.Name = stem
		}
		out = append(out, bp)
	}
	return out, nil
}

func (s *BlueprintService) GetBlueprint(_ context.Context, name string) (dto.BlueprintDTO, error) {
	dir, err := blueprintsDir()
	if err != nil {
		return dto.BlueprintDTO{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, name+".yaml"))
	if err != nil {
		return dto.BlueprintDTO{}, err
	}
	var bp dto.BlueprintDTO
	if err := yaml.Unmarshal(data, &bp); err != nil {
		return dto.BlueprintDTO{}, err
	}
	bp.FileName = name + ".yaml"
	return bp, nil
}

func (s *BlueprintService) SaveBlueprint(_ context.Context, bp dto.BlueprintDTO) error {
	if bp.Name == "" {
		return fmt.Errorf("blueprint name required")
	}
	dir, err := blueprintsDir()
	if err != nil {
		return err
	}
	// clear FileName before writing — it's runtime-only
	toWrite := bp
	toWrite.FileName = ""
	data, err := yaml.Marshal(toWrite)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, bp.Name+".yaml"), data, 0o644)
}

func (s *BlueprintService) DeleteBlueprint(_ context.Context, name string) error {
	dir, err := blueprintsDir()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(dir, name+".yaml"))
}

func (s *BlueprintService) DeployBlueprint(ctx context.Context, req dto.BlueprintDeployRequest) (string, error) {
	bp, err := s.GetBlueprint(ctx, req.BlueprintName)
	if err != nil {
		return "", fmt.Errorf("load blueprint: %w", err)
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}

	id, done := StartOp("blueprint.deploy", req.BlueprintName)
	go func() {
		defer WithTarget(req.BlueprintName)()
		var opErr error
		defer func() { done(opErr) }()

		app := application.Get()
		if app != nil {
			app.Event.Emit("blueprint:deploying", dto.BlueprintEventDTO{Name: req.BlueprintName})
		}

		for _, entry := range bp.Charts {
			if err := deployChart(req, entry); err != nil {
				msg := fmt.Sprintf("[%s] %s", entry.ReleaseName, err.Error())
				slog.Error("blueprint chart deploy failed", "blueprint", req.BlueprintName, "release", entry.ReleaseName, "err", err)
				if app != nil {
					app.Event.Emit("blueprint:error", dto.BlueprintEventDTO{Name: req.BlueprintName, Message: msg})
				}
				opErr = fmt.Errorf("%s", msg)
				return
			}
		}

		if app != nil {
			app.Event.Emit("blueprint:done", dto.BlueprintEventDTO{Name: req.BlueprintName})
		}
	}()

	return id, nil
}

func deployChart(req dto.BlueprintDeployRequest, entry dto.ChartEntryDTO) error {
	settings := cli.New()
	settings.KubeContext = "k3d-" + req.ClusterName

	actionConfig := new(action.Configuration)
	helmLog := func(format string, v ...interface{}) {
		slog.Info(fmt.Sprintf(format, v...), "source", "helm", "target", req.BlueprintName)
	}
	if err := actionConfig.Init(settings.RESTClientGetter(), req.Namespace, "secrets", helmLog); err != nil {
		return fmt.Errorf("init helm action config: %w", err)
	}

	// pull chart to temp dir
	tmpDir, err := os.MkdirTemp("", "k3desktop-helm-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// strip optional "repo-alias/chart" prefix — only the chart name goes to pull.Run
	chartName := entry.Chart
	if idx := strings.LastIndex(chartName, "/"); idx >= 0 {
		chartName = chartName[idx+1:]
	}

	pull := action.NewPullWithOpts(action.WithConfig(actionConfig))
	pull.Settings = settings
	pull.RepoURL = entry.Repo
	pull.Version = entry.Version
	pull.Untar = true
	pull.UntarDir = tmpDir
	pull.DestDir = tmpDir

	slog.Info(fmt.Sprintf("pulling chart %s/%s@%s", entry.Repo, chartName, entry.Version), "target", req.BlueprintName)
	if _, err := pull.Run(chartName); err != nil {
		return fmt.Errorf("pull chart %s: %w", chartName, err)
	}

	// helm pull untars into tmpDir/<chart-name>/
	chartDir := filepath.Join(tmpDir, chartName)
	ch, err := loader.Load(chartDir)
	if err != nil {
		return fmt.Errorf("load chart: %w", err)
	}

	// parse inline values
	vals := map[string]interface{}{}
	if entry.Values != "" {
		if err := yaml.Unmarshal([]byte(entry.Values), &vals); err != nil {
			return fmt.Errorf("parse values YAML: %w", err)
		}
	}

	slog.Info(fmt.Sprintf("deploying release %s (chart %s) to namespace %s", entry.ReleaseName, entry.Chart, req.Namespace), "target", req.BlueprintName)

	// Check if release exists to decide install vs upgrade
	histClient := action.NewHistory(actionConfig)
	histClient.Max = 1
	_, histErr := histClient.Run(entry.ReleaseName)
	if histErr == driver.ErrReleaseNotFound {
		install := action.NewInstall(actionConfig)
		install.ReleaseName = entry.ReleaseName
		install.Namespace = req.Namespace
		install.CreateNamespace = true
		if _, err := install.Run(ch, vals); err != nil {
			return fmt.Errorf("install %s: %w", entry.ReleaseName, err)
		}
	} else {
		upgrade := action.NewUpgrade(actionConfig)
		upgrade.Namespace = req.Namespace
		if _, err := upgrade.Run(entry.ReleaseName, ch, vals); err != nil {
			return fmt.Errorf("upgrade %s: %w", entry.ReleaseName, err)
		}
	}

	slog.Info(fmt.Sprintf("release %s deployed successfully", entry.ReleaseName), "target", req.BlueprintName)
	return nil
}
