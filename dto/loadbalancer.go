package dto

type LoadBalancerDTO struct {
	Name        string   `json:"name"`
	ClusterName string   `json:"clusterName"`
	State       string   `json:"state"` // "running" | "stopped"
	Image       string   `json:"image"`
	Ports       []string `json:"ports"` // "hostIP:hostPort→containerPort/proto"
}
