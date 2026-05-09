package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/k3desktop/k3desktop/dto"
)

func TestProfileService_ListProfiles_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := profilesDirFunc
	profilesDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { profilesDirFunc = origFunc }()

	svc := &ProfileService{}
	list, err := svc.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if list != nil && len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestProfileService_SaveAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := profilesDirFunc
	profilesDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { profilesDirFunc = origFunc }()

	svc := &ProfileService{}

	content := `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: test-profile
servers: 1
agents: 2
`

	err := svc.SaveProfile(context.Background(), "test-profile", content)
	if err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	// Verify file exists
	path := filepath.Join(tmpDir, "test-profile.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("profile file not created")
	}

	got, err := svc.GetProfile(context.Background(), "test-profile")
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if got != content {
		t.Errorf("GetProfile content mismatch:\ngot: %q\nwant: %q", got, content)
	}
}

func TestProfileService_SaveEmptyName(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := profilesDirFunc
	profilesDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { profilesDirFunc = origFunc }()

	svc := &ProfileService{}

	err := svc.SaveProfile(context.Background(), "", "content")
	if err == nil {
		t.Error("SaveProfile with empty name should error")
	}
}

func TestProfileService_SaveInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := profilesDirFunc
	profilesDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { profilesDirFunc = origFunc }()

	svc := &ProfileService{}

	err := svc.SaveProfile(context.Background(), "bad", "{{{{not valid yaml")
	if err == nil {
		t.Error("SaveProfile with invalid YAML should error")
	}
}

func TestProfileService_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := profilesDirFunc
	profilesDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { profilesDirFunc = origFunc }()

	svc := &ProfileService{}

	content := `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: to-delete
servers: 1
`
	svc.SaveProfile(context.Background(), "to-delete", content)

	err := svc.DeleteProfile(context.Background(), "to-delete")
	if err != nil {
		t.Fatalf("DeleteProfile: %v", err)
	}

	path := filepath.Join(tmpDir, "to-delete.yaml")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("profile file should be deleted")
	}
}

func TestProfileService_ListProfiles(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := profilesDirFunc
	profilesDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { profilesDirFunc = origFunc }()

	svc := &ProfileService{}

	content := `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: profile1
servers: 1
`
	svc.SaveProfile(context.Background(), "profile1", content)
	svc.SaveProfile(context.Background(), "profile2", content)

	// Non-YAML files should be ignored
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("ignore"), 0o644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0o755)

	list, err := svc.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("list len = %d, want 2", len(list))
	}

	names := map[string]bool{}
	for _, p := range list {
		names[p.Name] = true
	}
	if !names["profile1"] || !names["profile2"] {
		t.Errorf("expected profile1 and profile2, got %v", list)
	}
}

func TestProfileService_ListProfiles_YMLExtension(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := profilesDirFunc
	profilesDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { profilesDirFunc = origFunc }()

	// Write a .yml file directly
	os.WriteFile(filepath.Join(tmpDir, "test.yml"), []byte("apiVersion: k3d.io/v1alpha5\nkind: Simple\nservers: 1\n"), 0o644)

	svc := &ProfileService{}
	list, err := svc.ListProfiles(context.Background())
	if err != nil {
		t.Fatalf("ListProfiles: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("list len = %d, want 1", len(list))
	}
	if list[0].Name != "test" {
		t.Errorf("Name = %q, want %q", list[0].Name, "test")
	}
	if list[0].FileName != "test.yml" {
		t.Errorf("FileName = %q, want %q", list[0].FileName, "test.yml")
	}
}

func TestProfileService_GetNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	origFunc := profilesDirFunc
	profilesDirFunc = func() (string, error) { return tmpDir, nil }
	defer func() { profilesDirFunc = origFunc }()

	svc := &ProfileService{}
	_, err := svc.GetProfile(context.Background(), "nonexistent")
	if err == nil {
		t.Error("GetProfile for nonexistent should error")
	}
}

func TestProfileService_AdvancedRequestToYAML(t *testing.T) {
	svc := &ProfileService{}

	req := dto.ClusterCreateAdvancedRequest{
		Name:    "yaml-test",
		Servers: 1,
		Agents:  2,
		Image:   "rancher/k3s:v1.28.0-k3s1",
	}

	yamlStr, err := svc.AdvancedRequestToYAML(context.Background(), req)
	if err != nil {
		t.Fatalf("AdvancedRequestToYAML: %v", err)
	}
	if yamlStr == "" {
		t.Error("YAML output should not be empty")
	}
	// Should contain the cluster name
	if !contains(yamlStr, "yaml-test") {
		t.Error("YAML should contain cluster name")
	}
}

func TestProfileService_YAMLToAdvancedRequest(t *testing.T) {
	svc := &ProfileService{}

	yamlContent := `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: from-yaml
servers: 3
agents: 5
image: rancher/k3s:v1.28.0-k3s1
`

	req, err := svc.YAMLToAdvancedRequest(context.Background(), yamlContent)
	if err != nil {
		t.Fatalf("YAMLToAdvancedRequest: %v", err)
	}
	if req.Name != "from-yaml" {
		t.Errorf("Name = %q, want %q", req.Name, "from-yaml")
	}
	if req.Servers != 3 {
		t.Errorf("Servers = %d, want 3", req.Servers)
	}
	if req.Agents != 5 {
		t.Errorf("Agents = %d, want 5", req.Agents)
	}
}

func TestProfileService_YAMLToAdvancedRequest_InvalidYAML(t *testing.T) {
	svc := &ProfileService{}

	_, err := svc.YAMLToAdvancedRequest(context.Background(), "{{not valid")
	if err == nil {
		t.Error("YAMLToAdvancedRequest with invalid YAML should error")
	}
}

func TestProfileService_RoundTrip_AdvancedToYAMLAndBack(t *testing.T) {
	svc := &ProfileService{}

	original := dto.ClusterCreateAdvancedRequest{
		Name:           "roundtrip-yaml",
		Servers:        2,
		Agents:         3,
		Image:          "rancher/k3s:v1.28.0-k3s1",
		APIPort:        "6443",
		NoLoadbalancer: true,
		Ports: []dto.NodeFilter{
			{Value: "8080:80", NodeFilters: "server:0"},
		},
	}

	yamlStr, err := svc.AdvancedRequestToYAML(context.Background(), original)
	if err != nil {
		t.Fatalf("AdvancedRequestToYAML: %v", err)
	}

	result, err := svc.YAMLToAdvancedRequest(context.Background(), yamlStr)
	if err != nil {
		t.Fatalf("YAMLToAdvancedRequest: %v", err)
	}

	if result.Name != original.Name {
		t.Errorf("Name = %q, want %q", result.Name, original.Name)
	}
	if result.Servers != original.Servers {
		t.Errorf("Servers = %d, want %d", result.Servers, original.Servers)
	}
	if result.Agents != original.Agents {
		t.Errorf("Agents = %d, want %d", result.Agents, original.Agents)
	}
	if result.NoLoadbalancer != original.NoLoadbalancer {
		t.Errorf("NoLoadbalancer = %v, want %v", result.NoLoadbalancer, original.NoLoadbalancer)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
