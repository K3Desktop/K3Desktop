package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/k3desktop/k3desktop/dto"
	"sigs.k8s.io/yaml"
)

func TestBlueprintService_SaveAndGet(t *testing.T) {
	// Override blueprintsDir for test isolation
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	bp := dto.BlueprintDTO{
		Name:        "test-bp",
		Description: "A test blueprint",
		Charts: []dto.ChartEntryDTO{
			{ReleaseName: "nginx", Repo: "https://charts.bitnami.com/bitnami", Chart: "nginx", Version: "15.0.0"},
		},
	}

	// Save
	err := svc.SaveBlueprint(context.Background(), bp)
	if err != nil {
		t.Fatalf("SaveBlueprint: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, "test-bp.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("blueprint file not created at %s", path)
	}

	// Read raw file and verify FileName is cleared (empty) when written
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw: %v", err)
	}
	if fn, ok := raw["fileName"]; ok && fn != "" {
		t.Errorf("FileName should be empty in serialized YAML, got %q", fn)
	}

	// Get
	got, err := svc.GetBlueprint(context.Background(), "test-bp")
	if err != nil {
		t.Fatalf("GetBlueprint: %v", err)
	}
	if got.Name != "test-bp" {
		t.Errorf("Name = %q, want %q", got.Name, "test-bp")
	}
	if got.Description != "A test blueprint" {
		t.Errorf("Description = %q, want %q", got.Description, "A test blueprint")
	}
	if len(got.Charts) != 1 {
		t.Fatalf("Charts len = %d, want 1", len(got.Charts))
	}
	if got.Charts[0].ReleaseName != "nginx" {
		t.Errorf("Charts[0].ReleaseName = %q, want %q", got.Charts[0].ReleaseName, "nginx")
	}
	if got.FileName != "test-bp.yaml" {
		t.Errorf("FileName = %q, want %q", got.FileName, "test-bp.yaml")
	}
}

func TestBlueprintService_List(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	// Save two blueprints
	svc.SaveBlueprint(context.Background(), dto.BlueprintDTO{Name: "bp1", Description: "first"})
	svc.SaveBlueprint(context.Background(), dto.BlueprintDTO{Name: "bp2", Description: "second"})

	// Also add a non-YAML file that should be ignored
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("ignore me"), 0o644)

	// Also add a directory that should be ignored
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0o755)

	list, err := svc.ListBlueprints(context.Background())
	if err != nil {
		t.Fatalf("ListBlueprints: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}

	names := map[string]bool{}
	for _, bp := range list {
		names[bp.Name] = true
	}
	if !names["bp1"] || !names["bp2"] {
		t.Errorf("expected bp1 and bp2 in list, got %v", list)
	}
}

func TestBlueprintService_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	svc.SaveBlueprint(context.Background(), dto.BlueprintDTO{Name: "to-delete", Description: "will be deleted"})

	err := svc.DeleteBlueprint(context.Background(), "to-delete")
	if err != nil {
		t.Fatalf("DeleteBlueprint: %v", err)
	}

	path := filepath.Join(tmpDir, "to-delete.yaml")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("blueprint file should be deleted")
	}
}

func TestBlueprintService_SaveEmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	err := svc.SaveBlueprint(context.Background(), dto.BlueprintDTO{Name: ""})
	if err == nil {
		t.Error("SaveBlueprint with empty name should error")
	}
}

func TestBlueprintService_GetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	_, err := svc.GetBlueprint(context.Background(), "nonexistent")
	if err == nil {
		t.Error("GetBlueprint for nonexistent should error")
	}
}

func TestBlueprintService_ListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	list, err := svc.ListBlueprints(context.Background())
	if err != nil {
		t.Fatalf("ListBlueprints: %v", err)
	}
	if list != nil && len(list) != 0 {
		t.Errorf("list should be empty, got %d", len(list))
	}
}

func TestBlueprintService_ListIgnoresInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	// Save a valid blueprint
	svc.SaveBlueprint(context.Background(), dto.BlueprintDTO{Name: "valid", Description: "ok"})

	// Write an invalid YAML file
	os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte("{{{{not yaml"), 0o644)

	list, err := svc.ListBlueprints(context.Background())
	if err != nil {
		t.Fatalf("ListBlueprints: %v", err)
	}
	// Should only get the valid one
	if len(list) != 1 {
		t.Errorf("list len = %d, want 1 (invalid YAML should be skipped)", len(list))
	}
}

func TestBlueprintService_SaveOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	svc.SaveBlueprint(context.Background(), dto.BlueprintDTO{Name: "overwrite", Description: "v1"})
	svc.SaveBlueprint(context.Background(), dto.BlueprintDTO{Name: "overwrite", Description: "v2"})

	got, err := svc.GetBlueprint(context.Background(), "overwrite")
	if err != nil {
		t.Fatalf("GetBlueprint: %v", err)
	}
	if got.Description != "v2" {
		t.Errorf("Description = %q, want %q (should be overwritten)", got.Description, "v2")
	}
}

func TestBlueprintService_ChartValues(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	svc := &BlueprintService{}

	bp := dto.BlueprintDTO{
		Name: "with-values",
		Charts: []dto.ChartEntryDTO{
			{
				ReleaseName: "nginx",
				Repo:        "https://charts.bitnami.com/bitnami",
				Chart:       "nginx",
				Version:     "15.0.0",
				Values:      "replicaCount: 3\nservice:\n  type: ClusterIP",
			},
		},
	}

	svc.SaveBlueprint(context.Background(), bp)

	got, _ := svc.GetBlueprint(context.Background(), "with-values")
	if got.Charts[0].Values == "" {
		t.Error("Values should be preserved")
	}
}

func TestBlueprintService_ListUsesFileStem(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := blueprintsDirFunc
	blueprintsDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { blueprintsDirFunc = origFunc }()

	// Write a file with no name field in the YAML — name should come from filename stem
	content := "description: no-name-field\ncharts: []\n"
	os.WriteFile(filepath.Join(tmpDir, "from-file.yaml"), []byte(content), 0o644)

	svc := &BlueprintService{}
	list, err := svc.ListBlueprints(context.Background())
	if err != nil {
		t.Fatalf("ListBlueprints: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}
	if list[0].Name != "from-file" {
		t.Errorf("Name = %q, want %q (from filename stem)", list[0].Name, "from-file")
	}
}
