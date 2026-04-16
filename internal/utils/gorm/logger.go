package gormutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"go-fx-template/internal/config"
	"go-fx-template/internal/logger"
	"go-fx-template/internal/utils/text"
)

type GormLogger struct {
	messageLevel     gormlogger.LogLevel
	showRuntimeTrace bool
	showTimestamp    bool
	showStartupTrace bool
}

var instance *GormLogger

func NewCustomLogger(messageLevel gormlogger.LogLevel, showRuntimeTrace, showStartupTrace bool) *GormLogger {
	instance = &GormLogger{
		messageLevel:     messageLevel,
		showRuntimeTrace: showRuntimeTrace,
		showStartupTrace: showStartupTrace,
	}
	return instance
}

func EnableTimestamps() {
	if instance != nil {
		instance.showTimestamp = true
	}
}

func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	cp := *l
	cp.messageLevel = level
	return &cp
}

func (l *GormLogger) Info(_ context.Context, msg string, data ...interface{}) {
	if l.messageLevel < gormlogger.Info {
		return
	}
	l.logf("%s%s | %s\n", l.tsPrefix(), text.Purple("GORM"), fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.messageLevel < gormlogger.Warn {
		return
	}
	l.logf("%s%s | %s\n", l.tsPrefix(), text.Purple("GORM"), fmt.Sprintf(msg, data...))
}

func (l *GormLogger) Error(_ context.Context, msg string, data ...interface{}) {
	if l.messageLevel < gormlogger.Error {
		return
	}
	l.logf("%s%s | %s\n", l.tsPrefix(), text.Purple("GORM"), fmt.Sprintf(msg, data...))
}

func (l *GormLogger) tsPrefix() string {
	if !l.showTimestamp {
		return ""
	}
	return time.Now().Format("2006/01/02 15:04:05") + " "
}

func (l *GormLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	sql, rows := fc()
	elapsed := time.Since(begin)
	ts := l.tsPrefix()
	reqID := requestIDFromContext(ctx)
	caller := shortCaller(findCaller())

	status := text.Green("Success")
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound) || rows == 0:
		status = text.Yellow("Not found")
	case err != nil:
		status = text.Red("Error")
	}

	meta := fmt.Sprintf("| %s | %.3fms | %s | %s",
		status, float64(elapsed.Nanoseconds())/1e6, formatRows(rows), caller)

	compacted := compactSQL(sql)

	// Startup (no timestamps) — single line with | separator
	if !l.showTimestamp {
		if !l.showStartupTrace {
			return
		}
		prefix := text.Purple("GORM DEBUG")
		l.logf("%s | %s %s\n", prefix, highlightSQL(compacted), meta)
		return
	}

	if !l.showRuntimeTrace {
		return
	}

	// Runtime — no | between prefix and SQL, single line, truncate if needed
	prefix := formatPrefix(l, reqID)

	if len(compacted) > maxSQLWidth {
		compacted = compacted[:maxSQLWidth-3] + "..."
	}

	l.logf("%s%s %s %s\n", ts, prefix, highlightSQL(compacted), meta)
}

// maxSQLWidth is the maximum visible SQL length before truncation.
const maxSQLWidth = 120

func formatPrefix(l *GormLogger, reqID string) string {
	if !l.showTimestamp {
		return text.Purple("GORM DEBUG")
	}
	if reqID == "" {
		return text.Purple("GORM ")
	}
	return fmt.Sprintf("%s  [%s]", text.Purple("GORM"), text.Gray(reqID))
}

func formatRows(rows int64) string {
	switch {
	case rows < 0:
		return "-"
	case rows == 1:
		return "1 row"
	default:
		return fmt.Sprintf("%d rows", rows)
	}
}

func requestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if v, ok := ctx.Value("request_id").(string); ok && v != "" {
		return v
	}

	return ""
}

func findCaller() string {
	for skip := 2; skip < 15; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}

		normalized := strings.ReplaceAll(file, "\\", "/")

		if strings.Contains(normalized, "/internal/utils/gorm/") {
			continue
		}

		if strings.Contains(normalized, "/gorm.io/") {
			continue
		}

		if idx := strings.Index(normalized, "internal/"); idx >= 0 {
			return fmt.Sprintf("%s:%d", normalized[idx:], line)
		}

		return fmt.Sprintf("%s:%d", normalized, line)
	}

	return "-"
}

func shortCaller(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")

	if idx := strings.Index(path, config.Project+"/"); idx >= 0 {
		return path[idx:]
	}

	return path
}

func compactSQL(sql string) string {
	return strings.Join(strings.Fields(sql), " ")
}

func (l *GormLogger) logf(format string, args ...any) {
	_, _ = io.WriteString(logger.Writer(), fmt.Sprintf(format, args...))
}
