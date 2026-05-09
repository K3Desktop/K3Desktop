package dto

type RegistryDTO struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	State    string `json:"state"`
}

type RegistryCreateRequest struct {
	Name string `json:"name"`
	Port int    `json:"port"` // 0 = random
}
