package dto

type KubeconfigContextDTO struct {
	Name    string `json:"name"`
	Cluster string `json:"cluster"`
	User    string `json:"user"`
	Current bool   `json:"current"`
}
