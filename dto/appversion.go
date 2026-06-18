package dto

type AppVersionDTO struct {
	Version    string `json:"version"`
	BuildDate  string `json:"buildDate"`
	CommitHash string `json:"commitHash"`
}
