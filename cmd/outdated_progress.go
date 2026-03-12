package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// outdatedProgressConsole 维护按工具分配的进度行；终端环境下原地刷新，其他环境降级为追加日志。
type outdatedProgressConsole struct {
	out         io.Writer
	lines       []string
	interactive bool
	printed     bool
	mu          sync.Mutex
}

func newOutdatedProgressConsole(out io.Writer, initialLines []string) *outdatedProgressConsole {
	console := &outdatedProgressConsole{
		out:         out,
		lines:       append([]string{}, initialLines...),
		interactive: isInteractiveProgressWriter(out),
	}
	if console.interactive && len(console.lines) > 0 {
		for _, line := range console.lines {
			_, _ = fmt.Fprintln(console.out, line)
		}
		console.printed = true
	}
	return console
}

func (c *outdatedProgressConsole) SetLine(index int, text string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if index < 0 || index >= len(c.lines) {
		return
	}
	c.lines[index] = text

	if !c.interactive || !c.printed {
		_, _ = fmt.Fprintln(c.out, text)
		return
	}

	linesUp := len(c.lines) - index
	if linesUp > 0 {
		_, _ = fmt.Fprintf(c.out, "\x1b[%dA", linesUp)
	}
	_, _ = fmt.Fprint(c.out, "\r\x1b[2K")
	_, _ = fmt.Fprint(c.out, text)
	if linesUp > 0 {
		_, _ = fmt.Fprintf(c.out, "\x1b[%dB", linesUp)
	}
	_, _ = fmt.Fprint(c.out, "\r")
}

// outdatedLineWriter 将多次写入的日志汇总到同一条进度行，并在末尾追加最终结果。
type outdatedLineWriter struct {
	console *outdatedProgressConsole
	index   int
	name    string

	mu      sync.Mutex
	entries []string
}

func newOutdatedLineWriter(console *outdatedProgressConsole, index int, name string) *outdatedLineWriter {
	return &outdatedLineWriter{
		console: console,
		index:   index,
		name:    name,
	}
}

func (w *outdatedLineWriter) Write(p []byte) (int, error) {
	w.Add(string(p))
	return len(p), nil
}

func (w *outdatedLineWriter) Add(text string) {
	entry := sanitizeProgressText(text)
	if entry == "" {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries = append(w.entries, entry)
	w.console.SetLine(w.index, w.renderLocked(""))
}

func (w *outdatedLineWriter) Finish(result string) {
	sanitized := sanitizeProgressText(result)

	w.mu.Lock()
	defer w.mu.Unlock()
	w.console.SetLine(w.index, w.renderLocked(sanitized))
}

func (w *outdatedLineWriter) renderLocked(result string) string {
	parts := append([]string{}, w.entries...)
	if result != "" {
		parts = append(parts, result)
	}
	if len(parts) == 0 {
		return w.name + " - pending"
	}
	return w.name + " - " + strings.Join(parts, " | ")
}

func sanitizeProgressText(raw string) string {
	if raw == "" {
		return ""
	}
	return strings.Join(strings.Fields(raw), " ")
}

func isInteractiveProgressWriter(out io.Writer) bool {
	file, ok := out.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
