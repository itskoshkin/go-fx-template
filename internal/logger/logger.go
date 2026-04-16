package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/viper"

	"go-fx-template/internal/config"
	"go-fx-template/internal/utils/colors"
	"go-fx-template/internal/utils/text"
)

var (
	logFile      *os.File
	panicLogFile *os.File
	currentLevel = LevelInfo
	consoleLog   *log.Logger
	fileLog      *log.Logger
	jsonFileLog  *log.Logger
)

const actionDoneColumn = 60

type Level int

const (
	LevelError Level = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

type jsonEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func SetupLogger() {
	fmt.Print("Setting up logger...")

	switch viper.GetString(config.LogLevel) {
	case "DEBUG":
		currentLevel = LevelDebug
	case "ERROR":
		currentLevel = LevelError
	case "WARN":
		currentLevel = LevelWarn
	default:
		currentLevel = LevelInfo
	}

	if viper.GetBool(config.LogToFile) {
		filePath := viper.GetString(config.LogFilePath)

		flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
		if viper.GetString(config.LogFileMode) == "overwrite" {
			flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
		}
		if viper.GetString(config.LogFileMode) == "rotate" {
			rotateLogFile(filePath, viper.GetString(config.LogFilesFolder))
		}

		var err error
		logFile, err = os.OpenFile(filePath, flags, 0666)
		if err != nil {
			fmt.Println()
			log.Fatalf("Fatal: failed to open log file: %v", err)
		}

		fileInfo, err := logFile.Stat()
		if err != nil {
			fmt.Println()
			log.Fatalf("Fatal: failed to stat log file: %v", err)
		}

		if viper.GetString(config.LogFormat) == "json" {
			jsonFileLog = log.New(logFile, "", 0)
			if viper.GetString(config.LogFileMode) == "append" && fileInfo.Size() > 0 {
				b, _ := json.Marshal(jsonEntry{Time: ts(), Level: "INFO", Message: "==== new run ===="})
				jsonFileLog.Println(string(b))
			}
		} else {
			fileLog = log.New(logFile, "", 0)
			if viper.GetString(config.LogFileMode) == "append" && fileInfo.Size() > 0 {
				_, _ = logFile.WriteString("\n\n==== New run at ====\n")
			}
		}
	}

	if viper.GetBool(config.LogToConsole) {
		consoleLog = log.New(os.Stdout, "", 0)
	}

	fmt.Println(text.Green("      Done."))
}

func rotateLogFile(filePath, logsFolder string) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return
	}

	if err := os.MkdirAll(logsFolder, 0755); err != nil {
		log.Printf("failed to create logs folder: %v", err)
		return
	}

	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	newName := filepath.Join(logsFolder, name+"_"+time.Now().Format("2006-01-02_15-04-05")+ext)
	if err := os.Rename(filePath, newName); err != nil {
		log.Printf("failed to rotate log file: %v", err)
	}
}

func Writer() io.Writer {
	var ws []io.Writer
	if consoleLog != nil {
		ws = append(ws, os.Stdout)
	}
	if logFile != nil {
		ws = append(ws, colors.NewANSIStripWriter(logFile))
	}
	if len(ws) == 0 {
		return io.Discard
	}
	return io.MultiWriter(ws...)
}

func FileWriter() io.Writer {
	if logFile == nil {
		return io.Discard
	}
	return logFile
}

func PanicFileWriter() io.Writer {
	if panicLogFile != nil {
		return panicLogFile
	}

	f, err := os.OpenFile("./panic.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return io.Discard
	}

	panicLogFile = f
	return panicLogFile
}

func GetLevel() Level { return currentLevel }

func SetLevel(l Level) { currentLevel = l }

func Debug(format string, args ...any) {
	if currentLevel < LevelDebug {
		return
	}
	write(text.Gray("DEBUG"), "DEBUG", fmt.Sprintf(format, args...))
}

func Info(format string, args ...any) {
	if currentLevel < LevelInfo {
		return
	}
	write(text.Green("INFO "), "INFO", fmt.Sprintf(format, args...))
}

func InfoAction(message string, action func() error) error {
	if currentLevel < LevelInfo {
		return action()
	}

	timestamp := ts()
	if consoleLog != nil {
		_, _ = fmt.Fprintf(os.Stdout, "%s %s %s", timestamp, text.Green("INFO "), message)
	}

	err := action()
	if err == nil {
		if consoleLog != nil {
			var padded = " "
			visibleWidth := utf8.RuneCountInString(timestamp) + 1 + utf8.RuneCountInString("INFO ") + 1 + utf8.RuneCountInString(message)
			if visibleWidth < actionDoneColumn {
				padded = strings.Repeat(" ", actionDoneColumn-visibleWidth)
			}
			_, _ = fmt.Fprintf(os.Stdout, "%s%s\n", padded, text.Green("Done."))
		}
		if fileLog != nil {
			fileLog.Printf("%s INFO %s Done.", timestamp, message)
		}
		if jsonFileLog != nil {
			b, _ := json.Marshal(jsonEntry{Time: timestamp, Level: "INFO", Message: message + " Done."})
			jsonFileLog.Println(string(b))
		}
		return nil
	}

	if consoleLog != nil {
		_, _ = fmt.Fprintln(os.Stdout)
	}
	if fileLog != nil {
		fileLog.Printf("%s INFO %s", timestamp, message)
	}
	if jsonFileLog != nil {
		b, _ := json.Marshal(jsonEntry{Time: timestamp, Level: "INFO", Message: message})
		jsonFileLog.Println(string(b))
	}

	return err
}

func InfoWithAction(message string, action func() error) error {
	return InfoAction(message, action)
}

func Warn(format string, args ...any) {
	if currentLevel < LevelWarn {
		return
	}
	write(text.Yellow("WARN "), "WARN", fmt.Sprintf(format, args...))
}

func Error(format string, args ...any) {
	write(text.Red("ERROR"), "ERROR", fmt.Sprintf(format, args...))
}

func ErrorWithID(ctx context.Context, format string, args ...any) {
	reqID, _ := ctx.Value("request_id").(string)
	Error("[%s] "+text.Red(format), append([]any{reqID}, args...)...)
}

func Fatal(v ...any) {
	write(text.Bold(text.Red("FATAL")), "FATAL", fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	write(text.Bold(text.Red("FATAL")), "FATAL", fmt.Sprintf(format, args...))
	os.Exit(1)
}

func write(coloredLevel, plainLevel, s string) {
	writeAt(ts(), coloredLevel, plainLevel, s)
}

func writeAt(stamp, coloredLevel, plainLevel, s string) {
	if consoleLog != nil {
		consoleLog.Printf("%s %s %s", stamp, coloredLevel, s)
	}
	if fileLog != nil {
		fileLog.Printf("%s %s %s", stamp, plainLevel, s)
	}
	if jsonFileLog != nil {
		b, _ := json.Marshal(jsonEntry{Time: stamp, Level: plainLevel, Message: s})
		jsonFileLog.Println(string(b))
	}
}

func ts() string { return time.Now().Format("2006/01/02 15:04:05") }

func Timestamp() string { return ts() }

func donePadding(timestamp, level, message string) string {
	visibleWidth := utf8.RuneCountInString(timestamp) + 1 + utf8.RuneCountInString(level) + 1 + utf8.RuneCountInString(message)
	if visibleWidth >= actionDoneColumn {
		return " "
	}
	return strings.Repeat(" ", actionDoneColumn-visibleWidth)
}

// GlobalLogger wraps package-level functions so they can be injected as a
// narrow Logger dependency — lets consumers (like pkg/postgres, pkg/redis)
// depend on a local interface instead of importing this package directly.
type GlobalLogger struct{}

func (GlobalLogger) Debug(format string, v ...any) { Debug(format, v...) }
func (GlobalLogger) Info(format string, v ...any)  { Info(format, v...) }
func (GlobalLogger) InfoAction(message string, action func() error) error {
	return InfoAction(message, action)
}
func (GlobalLogger) InfoWithAction(message string, action func() error) error {
	return InfoWithAction(message, action)
}
func (GlobalLogger) Warn(format string, v ...any)  { Warn(format, v...) }
func (GlobalLogger) Error(format string, v ...any) { Error(format, v...) }
