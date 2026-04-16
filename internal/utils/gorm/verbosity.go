package gormutils

import (
	"strings"

	"gorm.io/gorm/logger"
)

type verbosity int

const (
	verbositySilent verbosity = iota
	verbosityError
	verbosityWarn
	verbosityInfo
	verbosityDebug
)

func parseVerbosity(level string) verbosity {
	switch strings.ToUpper(level) {
	case "SILENT":
		return verbositySilent
	case "ERROR":
		return verbosityError
	case "WARN":
		return verbosityWarn
	case "DEBUG":
		return verbosityDebug
	case "INFO":
		fallthrough
	default:
		return verbosityInfo
	}
}

// RuntimeLoggerConfig maps a GORM log-level string to GORM's own level plus a
// runtime-trace flag. GORM has no real "Debug" level — per-query SQL trace is
// a separate toggle checked inside the adapter's Trace() method, so DEBUG maps
// to (Info, showRuntimeTrace=true). This function is independent of the
// application log level: GORM is as loud as GormLogLevel says, regardless of
// what app.log.level is set to.
func RuntimeLoggerConfig(gormLevel string) (logger.LogLevel, bool) {
	switch parseVerbosity(gormLevel) {
	case verbosityError:
		return logger.Error, false
	case verbosityWarn:
		return logger.Warn, false
	case verbosityInfo:
		return logger.Info, false
	case verbosityDebug:
		return logger.Info, true
	default:
		return logger.Silent, false
	}
}
