package service

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/k3desktop/k3desktop/dto"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	application.RegisterEvent[dto.OperationEventDTO]("op:start")
	application.RegisterEvent[dto.OperationEventDTO]("op:done")
	application.RegisterEvent[dto.OperationEventDTO]("op:error")
}

type operation struct {
	ID        string
	Kind      string
	Target    string
	StartedAt time.Time
}

var opRegistry = struct {
	sync.Mutex
	m map[string]operation
}{m: make(map[string]operation)}

// StartOp registers a new in-flight operation and emits op:start.
// The returned closure deregisters and emits op:done (err == nil) or op:error.
// Pattern:
//
//	id, done := StartOp("cluster.start", name)
//	go func() {
//	    defer WithTarget(name)()
//	    var err error
//	    defer func() { done(err) }()
//	    err = doWork()
//	}()
func StartOp(kind, target string) (string, func(err error)) {
	op := operation{
		ID:        uuid.NewString(),
		Kind:      kind,
		Target:    target,
		StartedAt: time.Now(),
	}
	opRegistry.Lock()
	opRegistry.m[op.ID] = op
	opRegistry.Unlock()

	emit("op:start", op, "", nil)

	var once sync.Once
	return op.ID, func(err error) {
		once.Do(func() {
			opRegistry.Lock()
			delete(opRegistry.m, op.ID)
			opRegistry.Unlock()
			if err != nil {
				emit("op:error", op, "", err)
			} else {
				emit("op:done", op, "", nil)
			}
		})
	}
}

func emit(event string, op operation, message string, err error) {
	app := application.Get()
	if app == nil {
		return
	}
	payload := dto.OperationEventDTO{
		ID:        op.ID,
		Kind:      op.Kind,
		Target:    op.Target,
		Phase:     phaseFromEvent(event),
		Message:   message,
		StartedAt: op.StartedAt.UTC().Format(time.RFC3339),
	}
	if err != nil {
		payload.Error = err.Error()
	}
	app.Event.Emit(event, payload)
}

func phaseFromEvent(event string) string {
	switch event {
	case "op:start":
		return "start"
	case "op:done":
		return "done"
	case "op:error":
		return "error"
	}
	return ""
}

// OperationsService exposes the active-operation registry to the frontend.
type OperationsService struct{}

// ListActive returns a snapshot of currently in-flight operations.
// The frontend calls this at startup to rehydrate state after a window reload.
// Returned entries have Phase = "start".
func (s *OperationsService) ListActive(_ context.Context) []dto.OperationEventDTO {
	opRegistry.Lock()
	defer opRegistry.Unlock()
	out := make([]dto.OperationEventDTO, 0, len(opRegistry.m))
	for _, op := range opRegistry.m {
		out = append(out, dto.OperationEventDTO{
			ID:        op.ID,
			Kind:      op.Kind,
			Target:    op.Target,
			Phase:     "start",
			StartedAt: op.StartedAt.UTC().Format(time.RFC3339),
		})
	}
	return out
}
