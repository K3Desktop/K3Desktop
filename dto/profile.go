package dto

type ProfileDTO struct {
	Name     string `json:"name"`     // filename stem, e.g. "dev-cluster"
	FileName string `json:"fileName"` // full filename, e.g. "dev-cluster.yaml"
}

type UlimitDTO struct {
	Name string `json:"name"`
	Soft int64  `json:"soft"`
	Hard int64  `json:"hard"`
}

type FileDTO struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Description string `json:"description"`
	NodeFilters string `json:"nodeFilters"` // comma-separated
}

type HostAliasDTO struct {
	IP        string   `json:"ip"`
	Hostnames []string `json:"hostnames"`
}
