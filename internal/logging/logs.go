package logs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"
)

// colors up top
const (
	reset  = "\033[0m"
	gray   = "\033[38;5;240m"
	blue   = "\033[38;5;75m"
	green  = "\033[38;5;114m"
	yellow = "\033[38;5;221m"
	red    = "\033[38;5;203m"
	purple = "\033[38;5;176m"
	cyan   = "\033[38;5;80m"
)

/*
exposed:
Start() - begins logging to file and handles output quality control

internals:
*/

func Start(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	logger := slog.New(NewHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
}

type Handler struct {
	w    io.Writer
	mu   sync.Mutex
	opts slog.HandlerOptions
	pre  []slog.Attr
}

func NewHandler(w io.Writer, opts *slog.HandlerOptions) *Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &Handler{w: w, opts: *opts}
}

func (h *Handler) Enabled(_ context.Context, l slog.Level) bool {
	min := h.opts.Level
	if min == nil {
		min = slog.LevelInfo
	}
	return l >= min.Level()
}

func (h *Handler) WithGroup(_ string) slog.Handler { return h }

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{w: h.w, opts: h.opts, pre: append(h.pre, attrs...)}
}

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer

	// timestamp
	fmt.Fprintf(&buf, "%s%s%s ", gray, r.Time.Format("2006-01-02 15:04:05.000"), reset)

	// level
	switch r.Level {
	case slog.LevelDebug:
		fmt.Fprintf(&buf, "%sDEBG%s ", blue, reset)
	case slog.LevelInfo:
		fmt.Fprintf(&buf, "%sINFO%s ", green, reset)
	case slog.LevelWarn:
		fmt.Fprintf(&buf, "%sWARN%s ", yellow, reset)
	case slog.LevelError:
		fmt.Fprintf(&buf, "%sERRO%s ", red, reset)
	default:
		fmt.Fprintf(&buf, "%s%s%s ", gray, r.Level, reset)
	}

	// message
	fmt.Fprintf(&buf, "%s", r.Message)

	writeAttr := func(a slog.Attr) {
		fmt.Fprintf(&buf, " %s%s%s=%s", purple, a.Key, reset, colorVal(a.Value))
	}

	for _, a := range h.pre {
		writeAttr(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		writeAttr(a)
		return true
	})

	buf.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	return err
}

func colorVal(v slog.Value) string {
	switch v.Kind() {
	case slog.KindString:
		return fmt.Sprintf(`%s"%s"%s`, green, v.String(), reset)
	case slog.KindInt64:
		return fmt.Sprintf("%s%s%s", cyan, strconv.FormatInt(v.Int64(), 10), reset)
	case slog.KindFloat64:
		return fmt.Sprintf("%s%s%s", cyan, strconv.FormatFloat(v.Float64(), 'f', -1, 64), reset)
	case slog.KindBool:
		return fmt.Sprintf("%s%s%s", cyan, strconv.FormatBool(v.Bool()), reset)
	case slog.KindDuration:
		return fmt.Sprintf("%s%s%s", cyan, v.Duration().Round(time.Microsecond), reset)
	case slog.KindAny:
		if err, ok := v.Any().(error); ok {
			return fmt.Sprintf("%s%s%s", red, err.Error(), reset)
		}
		return fmt.Sprintf("%s%v%s", gray, v.Any(), reset)
	default:
		return fmt.Sprintf("%s%s%s", gray, v.String(), reset)
	}
}
