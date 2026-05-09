package dto

type LogEntryDTO struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Source  string `json:"source"` // "app" | "k3d"
	Target  string `json:"target"` // cluster/node name this log belongs to, empty = unattributed
}
