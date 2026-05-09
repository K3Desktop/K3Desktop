package service

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	k3dlogger "github.com/k3d-io/k3d/v5/pkg/logger"
	"github.com/k3desktop/k3desktop/dto"
	"github.com/sirupsen/logrus"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func init() {
	application.RegisterEvent[dto.LogEntryDTO]("log:entry")
}

const ringSize = 500

// logLevel controls the active level for both app and k3d logs at runtime.
var logLevel = new(slog.LevelVar) // default: INFO

// ring is the global in-memory log buffer.
var ring = &ringBuffer{buf: make([]dto.LogEntryDTO, ringSize)}

// goroutineTarget maps goroutine ID → operation target name (e.g. cluster/node name).
var goroutineTarget sync.Map // map[uint64]string

// activeTargets counts how many ops are currently registered per target name.
var activeTargets sync.Map // map[string]int

// currentGoroutineID returns the ID of the calling goroutine by parsing runtime.Stack output.
func currentGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// format: "goroutine 18 [running]:\n..."
	s := strings.TrimPrefix(string(buf[:n]), "goroutine ")
	s = s[:strings.IndexByte(s, ' ')]
	id, _ := strconv.ParseUint(s, 10, 64)
	return id
}

// WithTarget registers a target name for the current goroutine and returns a cleanup func.
// Usage: defer WithTarget("my-cluster")()
func WithTarget(name string) func() {
	id := currentGoroutineID()
	goroutineTarget.Store(id, name)
	// increment active count for this target
	for {
		v, _ := activeTargets.LoadOrStore(name, 1)
		if v == 1 {
			break
		}
		if activeTargets.CompareAndSwap(name, v, v.(int)+1) {
			break
		}
	}
	return func() {
		goroutineTarget.Delete(id)
		for {
			v, ok := activeTargets.Load(name)
			if !ok {
				break
			}
			n := v.(int) - 1
			if n <= 0 {
				activeTargets.Delete(name)
				break
			}
			if activeTargets.CompareAndSwap(name, v, n) {
				break
			}
		}
	}
}

// resolveTarget returns the target for the current log entry.
// Priority: goroutine registry → sole active target fallback → "".
func resolveTarget(goroutineID uint64) string {
	if v, ok := goroutineTarget.Load(goroutineID); ok {
		return v.(string)
	}
	// Fallback: if exactly one target is active, attribute untagged logs to it.
	var sole string
	count := 0
	activeTargets.Range(func(k, _ any) bool {
		sole = k.(string)
		count++
		return count < 2 // stop early if we find two
	})
	if count == 1 {
		return sole
	}
	return ""
}

func init() {
	h := &multiHandler{
		stderr: slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}),
		ring:   ring,
		level:  logLevel,
	}
	slog.SetDefault(slog.New(h))

	k3dlogger.Logger.SetLevel(logrus.TraceLevel)
	k3dlogger.Logger.AddHook(&logrusToSlogHook{})
	k3dlogger.Logger.SetOutput(noopWriter{})
}

// multiHandler fans out to stderr text handler + ring buffer + Wails event.
type multiHandler struct {
	stderr slog.Handler
	ring   *ringBuffer
	level  *slog.LevelVar
}

func (h *multiHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	_ = h.stderr.Handle(ctx, r)

	source := "app"
	target := ""
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "source":
			source = a.Value.String()
		case "target":
			target = a.Value.String()
		}
		return true
	})
	if target == "" {
		target = resolveTarget(currentGoroutineID())
	}

	entry := dto.LogEntryDTO{
		Time:    r.Time.Format(time.RFC3339),
		Level:   r.Level.String(),
		Message: r.Message,
		Source:  source,
		Target:  target,
	}
	h.ring.push(entry)

	go func() {
		app := application.Get()
		if app == nil {
			return
		}
		app.Event.Emit("log:entry", entry)
	}()
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &multiHandler{stderr: h.stderr.WithAttrs(attrs), ring: h.ring, level: h.level}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	return &multiHandler{stderr: h.stderr.WithGroup(name), ring: h.ring, level: h.level}
}

// ringBuffer is a fixed-size FIFO of log entries, goroutine-safe.
type ringBuffer struct {
	mu   sync.Mutex
	buf  []dto.LogEntryDTO
	head int // next write position
	full bool
}

func (rb *ringBuffer) push(e dto.LogEntryDTO) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.buf[rb.head] = e
	rb.head = (rb.head + 1) % len(rb.buf)
	if rb.head == 0 {
		rb.full = true
	}
}

func (rb *ringBuffer) snapshot() []dto.LogEntryDTO {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if !rb.full {
		out := make([]dto.LogEntryDTO, rb.head)
		copy(out, rb.buf[:rb.head])
		return out
	}
	out := make([]dto.LogEntryDTO, len(rb.buf))
	copy(out, rb.buf[rb.head:])
	copy(out[len(rb.buf)-rb.head:], rb.buf[:rb.head])
	return out
}

// logrusToSlogHook forwards k3d logrus entries into slog.
type logrusToSlogHook struct{}

func (h *logrusToSlogHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *logrusToSlogHook) Fire(entry *logrus.Entry) error {
	level := logrusLevelToSlog(entry.Level)
	if !slog.Default().Enabled(context.Background(), level) {
		return nil
	}
	attrs := make([]slog.Attr, 0, len(entry.Data)+1)
	for k, v := range entry.Data {
		attrs = append(attrs, slog.Any(k, v))
	}
	attrs = append(attrs, slog.String("source", "k3d"))
	slog.Default().LogAttrs(context.Background(), level, entry.Message, attrs...)
	return nil
}

func logrusLevelToSlog(l logrus.Level) slog.Level {
	switch l {
	case logrus.TraceLevel:
		return slog.LevelDebug - 4
	case logrus.DebugLevel:
		return slog.LevelDebug
	case logrus.InfoLevel:
		return slog.LevelInfo
	case logrus.WarnLevel:
		return slog.LevelWarn
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }

// LogService exposes log-level control and history to the frontend.
type LogService struct{}

var validLevels = map[string]slog.Level{
	"DEBUG": slog.LevelDebug,
	"INFO":  slog.LevelInfo,
	"WARN":  slog.LevelWarn,
	"ERROR": slog.LevelError,
}

func (s *LogService) SetLevel(_ context.Context, level string) error {
	l, ok := validLevels[level]
	if !ok {
		return fmt.Errorf("invalid log level %q: choose DEBUG, INFO, WARN, or ERROR", level)
	}
	logLevel.Set(l)
	slog.Info("log level changed", "level", level)
	return nil
}

func (s *LogService) GetLevel(_ context.Context) string {
	return logLevel.Level().String()
}

func (s *LogService) GetRecentLogs(_ context.Context) []dto.LogEntryDTO {
	return ring.snapshot()
}
