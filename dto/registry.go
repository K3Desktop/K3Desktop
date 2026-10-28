package dto

type RegistryDTO struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	State    string `json:"state"`
}

// ExternalRegistryDTO is a registry container not managed by k3d that can be attached to a cluster.
type ExternalRegistryDTO struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	Port  int    `json:"port"`
	// InternalPort is the port the registry listens on inside the container network.
	InternalPort int    `json:"internalPort"`
	State        string `json:"state"`
	// Reason is set when the container cannot be joined to a cluster network.
	Reason string `json:"reason"`
}

type RegistryCreateRequest struct {
	Name string `json:"name"`
	Port int    `json:"port"` // 0 = random
}
