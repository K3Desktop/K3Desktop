package dto

// OperationEventDTO is the unified payload for op:start / op:done / op:error events.
// One operation has a stable ID across all three phases so the frontend can correlate.
type OperationEventDTO struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`              // e.g. "cluster.start", "node.upgrade"
	Target    string `json:"target"`            // cluster/node/registry name
	Phase     string `json:"phase"`             // "start" | "done" | "error"
	Message   string `json:"message,omitempty"` // optional human-readable status
	Error     string `json:"error,omitempty"`   // populated when phase == "error"
	StartedAt string `json:"startedAt"`         // RFC3339
}
