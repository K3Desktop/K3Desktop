package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/k3desktop/k3desktop/dto"
	"github.com/sirupsen/logrus"
)

func TestRingBuffer_PushAndSnapshot(t *testing.T) {
	rb := &ringBuffer{buf: make([]dto.LogEntryDTO, 5)}

	rb.push(dto.LogEntryDTO{Message: "msg1"})
	rb.push(dto.LogEntryDTO{Message: "msg2"})
	rb.push(dto.LogEntryDTO{Message: "msg3"})

	snap := rb.snapshot()
	if len(snap) != 3 {
		t.Fatalf("snapshot len = %d, want 3", len(snap))
	}
	if snap[0].Message != "msg1" {
		t.Errorf("snap[0] = %q, want %q", snap[0].Message, "msg1")
	}
	if snap[2].Message != "msg3" {
		t.Errorf("snap[2] = %q, want %q", snap[2].Message, "msg3")
	}
}

func TestRingBuffer_Wrap(t *testing.T) {
	rb := &ringBuffer{buf: make([]dto.LogEntryDTO, 3)}

	rb.push(dto.LogEntryDTO{Message: "a"})
	rb.push(dto.LogEntryDTO{Message: "b"})
	rb.push(dto.LogEntryDTO{Message: "c"})
	rb.push(dto.LogEntryDTO{Message: "d"}) // overwrites "a"
	rb.push(dto.LogEntryDTO{Message: "e"}) // overwrites "b"

	snap := rb.snapshot()
	if len(snap) != 3 {
		t.Fatalf("snapshot len = %d, want 3", len(snap))
	}
	// Should be in order: c, d, e
	if snap[0].Message != "c" {
		t.Errorf("snap[0] = %q, want %q", snap[0].Message, "c")
	}
	if snap[1].Message != "d" {
		t.Errorf("snap[1] = %q, want %q", snap[1].Message, "d")
	}
	if snap[2].Message != "e" {
		t.Errorf("snap[2] = %q, want %q", snap[2].Message, "e")
	}
}

func TestRingBuffer_EmptySnapshot(t *testing.T) {
	rb := &ringBuffer{buf: make([]dto.LogEntryDTO, 5)}
	snap := rb.snapshot()
	if len(snap) != 0 {
		t.Errorf("empty snapshot len = %d, want 0", len(snap))
	}
}

func TestRingBuffer_SingleElement(t *testing.T) {
	rb := &ringBuffer{buf: make([]dto.LogEntryDTO, 1)}
	rb.push(dto.LogEntryDTO{Message: "only"})
	snap := rb.snapshot()
	if len(snap) != 1 || snap[0].Message != "only" {
		t.Errorf("single element snapshot failed: %v", snap)
	}

	rb.push(dto.LogEntryDTO{Message: "replaced"})
	snap = rb.snapshot()
	if len(snap) != 1 || snap[0].Message != "replaced" {
		t.Errorf("after wrap: %v", snap)
	}
}

func TestRingBuffer_ExactFill(t *testing.T) {
	rb := &ringBuffer{buf: make([]dto.LogEntryDTO, 3)}
	rb.push(dto.LogEntryDTO{Message: "1"})
	rb.push(dto.LogEntryDTO{Message: "2"})
	rb.push(dto.LogEntryDTO{Message: "3"})

	snap := rb.snapshot()
	if len(snap) != 3 {
		t.Fatalf("exact fill len = %d, want 3", len(snap))
	}
	if snap[0].Message != "1" || snap[1].Message != "2" || snap[2].Message != "3" {
		t.Errorf("exact fill = %v", snap)
	}
}

func TestRingBuffer_ConcurrentAccess(t *testing.T) {
	rb := &ringBuffer{buf: make([]dto.LogEntryDTO, 100)}
	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				rb.push(dto.LogEntryDTO{Message: fmt.Sprintf("writer-%d-%d", id, j)})
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				rb.snapshot()
			}
		}()
	}

	wg.Wait()

	snap := rb.snapshot()
	if len(snap) != 100 {
		// 10 * 50 = 500 writes into a buffer of 100, so it should be full
		t.Errorf("after concurrent writes, snapshot len = %d, want 100", len(snap))
	}
}

func TestWithTarget_RegistersAndCleans(t *testing.T) {
	cleanup := WithTarget("my-cluster")

	id := currentGoroutineID()
	val, ok := goroutineTarget.Load(id)
	if !ok {
		t.Fatal("goroutine target not registered")
	}
	if val != "my-cluster" {
		t.Errorf("target = %q, want %q", val, "my-cluster")
	}

	cleanup()

	_, ok = goroutineTarget.Load(id)
	if ok {
		t.Error("goroutine target still registered after cleanup")
	}
}

func TestWithTarget_ActiveTargetTracking(t *testing.T) {
	cleanup1 := WithTarget("cluster-a")
	defer cleanup1()

	val, ok := activeTargets.Load("cluster-a")
	if !ok {
		t.Fatal("active target not registered")
	}
	if val.(int) != 1 {
		t.Errorf("active count = %d, want 1", val.(int))
	}

	cleanup1()

	_, ok = activeTargets.Load("cluster-a")
	if ok {
		t.Error("active target still present after cleanup")
	}
}

func TestResolveTarget_FromGoroutine(t *testing.T) {
	cleanup := WithTarget("test-target")
	defer cleanup()

	target := resolveTarget(currentGoroutineID())
	if target != "test-target" {
		t.Errorf("resolveTarget = %q, want %q", target, "test-target")
	}
}

func TestResolveTarget_UnknownGoroutine_SoleTarget(t *testing.T) {
	// Ensure clean state
	activeTargets.Range(func(k, _ any) bool {
		activeTargets.Delete(k)
		return true
	})

	cleanup := WithTarget("sole-target")
	defer cleanup()

	// Use a goroutine ID that's not registered
	target := resolveTarget(999999999)
	if target != "sole-target" {
		t.Errorf("resolveTarget fallback = %q, want %q", target, "sole-target")
	}
}

func TestResolveTarget_UnknownGoroutine_MultipleTargets(t *testing.T) {
	// Ensure clean state
	activeTargets.Range(func(k, _ any) bool {
		activeTargets.Delete(k)
		return true
	})

	done := make(chan struct{})
	go func() {
		c1 := WithTarget("target-1")
		defer c1()
		done <- struct{}{}
		<-done // wait for signal to clean up
	}()
	<-done // wait for target-1 to be registered

	c2 := WithTarget("target-2")
	defer c2()

	target := resolveTarget(999999999)
	if target != "" {
		t.Errorf("resolveTarget with multiple targets = %q, want empty", target)
	}

	done <- struct{}{} // let goroutine clean up
	time.Sleep(10 * time.Millisecond)
}

func TestResolveTarget_NoTargets(t *testing.T) {
	// Ensure clean state
	activeTargets.Range(func(k, _ any) bool {
		activeTargets.Delete(k)
		return true
	})
	goroutineTarget.Range(func(k, _ any) bool {
		goroutineTarget.Delete(k)
		return true
	})

	target := resolveTarget(999999999)
	if target != "" {
		t.Errorf("resolveTarget with no targets = %q, want empty", target)
	}
}

func TestCurrentGoroutineID(t *testing.T) {
	id := currentGoroutineID()
	if id == 0 {
		t.Error("goroutine ID should not be 0")
	}

	// Different goroutines should have different IDs
	ch := make(chan uint64)
	go func() {
		ch <- currentGoroutineID()
	}()
	otherId := <-ch
	if id == otherId {
		t.Error("different goroutines returned same ID")
	}
}

func TestLogrusLevelToSlog(t *testing.T) {
	tests := []struct {
		input logrus.Level
		want  slog.Level
	}{
		{logrus.TraceLevel, slog.LevelDebug - 4},
		{logrus.DebugLevel, slog.LevelDebug},
		{logrus.InfoLevel, slog.LevelInfo},
		{logrus.WarnLevel, slog.LevelWarn},
		{logrus.ErrorLevel, slog.LevelError},
		{logrus.FatalLevel, slog.LevelError},
		{logrus.PanicLevel, slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.input.String(), func(t *testing.T) {
			got := logrusLevelToSlog(tt.input)
			if got != tt.want {
				t.Errorf("logrusLevelToSlog(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLogService_SetLevel(t *testing.T) {
	svc := &LogService{}

	tests := []struct {
		level   string
		wantErr bool
	}{
		{"DEBUG", false},
		{"INFO", false},
		{"WARN", false},
		{"ERROR", false},
		{"INVALID", true},
		{"debug", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			err := svc.SetLevel(context.Background(), tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetLevel(%q) err = %v, wantErr = %v", tt.level, err, tt.wantErr)
			}
		})
	}
}

func TestLogService_GetLevel(t *testing.T) {
	svc := &LogService{}

	svc.SetLevel(context.Background(), "WARN")
	if got := svc.GetLevel(context.Background()); got != "WARN" {
		t.Errorf("GetLevel = %q, want %q", got, "WARN")
	}

	svc.SetLevel(context.Background(), "DEBUG")
	if got := svc.GetLevel(context.Background()); got != "DEBUG" {
		t.Errorf("GetLevel = %q, want %q", got, "DEBUG")
	}

	// Reset
	svc.SetLevel(context.Background(), "INFO")
}

func TestLogService_GetRecentLogs(t *testing.T) {
	svc := &LogService{}

	logs := svc.GetRecentLogs(context.Background())
	if logs == nil {
		t.Error("GetRecentLogs returned nil, want empty slice or populated")
	}
}

func TestNoopWriter(t *testing.T) {
	w := noopWriter{}
	data := []byte("hello world")
	n, err := w.Write(data)
	if err != nil {
		t.Errorf("Write err = %v, want nil", err)
	}
	if n != len(data) {
		t.Errorf("Write n = %d, want %d", n, len(data))
	}
}

func TestLogrusToSlogHook_Levels(t *testing.T) {
	h := &logrusToSlogHook{}
	levels := h.Levels()
	if len(levels) != len(logrus.AllLevels) {
		t.Errorf("Levels() len = %d, want %d", len(levels), len(logrus.AllLevels))
	}
}

func TestMultiHandler_WithAttrs(t *testing.T) {
	h := &multiHandler{
		stderr: slog.NewTextHandler(noopWriter{}, nil),
		ring:   &ringBuffer{buf: make([]dto.LogEntryDTO, 10)},
		level:  new(slog.LevelVar),
	}

	attrs := []slog.Attr{slog.String("key", "value")}
	wrapped := h.WithAttrs(attrs)
	if wrapped == nil {
		t.Error("WithAttrs returned nil")
	}
	if _, ok := wrapped.(*multiHandler); !ok {
		t.Error("WithAttrs should return *multiHandler")
	}
}

func TestMultiHandler_WithGroup(t *testing.T) {
	h := &multiHandler{
		stderr: slog.NewTextHandler(noopWriter{}, nil),
		ring:   &ringBuffer{buf: make([]dto.LogEntryDTO, 10)},
		level:  new(slog.LevelVar),
	}

	wrapped := h.WithGroup("testgroup")
	if wrapped == nil {
		t.Error("WithGroup returned nil")
	}
	if _, ok := wrapped.(*multiHandler); !ok {
		t.Error("WithGroup should return *multiHandler")
	}
}

func TestMultiHandler_Enabled(t *testing.T) {
	lv := new(slog.LevelVar)
	lv.Set(slog.LevelWarn)

	h := &multiHandler{
		stderr: slog.NewTextHandler(noopWriter{}, nil),
		ring:   &ringBuffer{buf: make([]dto.LogEntryDTO, 10)},
		level:  lv,
	}

	if h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Enabled(INFO) should be false when level is WARN")
	}
	if !h.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("Enabled(WARN) should be true when level is WARN")
	}
	if !h.Enabled(context.Background(), slog.LevelError) {
		t.Error("Enabled(ERROR) should be true when level is WARN")
	}
}

func TestMultiHandler_Handle(t *testing.T) {
	rb := &ringBuffer{buf: make([]dto.LogEntryDTO, 10)}
	lv := new(slog.LevelVar)

	h := &multiHandler{
		stderr: slog.NewTextHandler(noopWriter{}, nil),
		ring:   rb,
		level:  lv,
	}

	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Errorf("Handle err = %v", err)
	}

	// Give async goroutine time to fire (the event Emit is in a goroutine)
	time.Sleep(10 * time.Millisecond)

	snap := rb.snapshot()
	if len(snap) < 1 {
		t.Fatal("ring buffer should have at least 1 entry")
	}
	if snap[len(snap)-1].Message != "test message" {
		t.Errorf("last message = %q, want %q", snap[len(snap)-1].Message, "test message")
	}
	if snap[len(snap)-1].Source != "app" {
		t.Errorf("source = %q, want %q", snap[len(snap)-1].Source, "app")
	}
}
