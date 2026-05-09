package dto

type BlueprintDTO struct {
	Name        string          `json:"name"        yaml:"name"`
	Description string          `json:"description" yaml:"description"`
	FileName    string          `json:"fileName"    yaml:"-"`
	Charts      []ChartEntryDTO `json:"charts"      yaml:"charts"`
}

type ChartEntryDTO struct {
	ReleaseName string `json:"releaseName" yaml:"releaseName"`
	Repo        string `json:"repo"        yaml:"repo"`
	Chart       string `json:"chart"       yaml:"chart"`
	Version     string `json:"version"     yaml:"version"`
	Values      string `json:"values"      yaml:"values"`
}

type BlueprintDeployRequest struct {
	BlueprintName string `json:"blueprintName"`
	ClusterName   string `json:"clusterName"`
	Namespace     string `json:"namespace"`
}

type BlueprintEventDTO struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}
