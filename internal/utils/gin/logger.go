package ginutils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"go-fx-template/internal/config"
	"go-fx-template/internal/logger"
	"go-fx-template/internal/utils/colors"
	"go-fx-template/internal/utils/text"
)

const (
	pathWidth    = 50
	pathMaxWidth = 80
)

func LoggingMiddlewares() []gin.HandlerFunc {
	var middlewares []gin.HandlerFunc
	// Access logs writers
	if logger.GetLevel() >= logger.LevelInfo {
		if viper.GetString(config.LogFormat) == "json" {
			if out := joinWriters(appendWriterIfEnabled(viper.GetBool(config.LogToConsole), os.Stdout), logger.FileWriter()); out != nil {
				middlewares = append(middlewares, customJSONLogger(out))
			}
		} else {
			if viper.GetBool(config.LogToConsole) {
				middlewares = append(middlewares, customTextLogger(os.Stdout))
			}
			if fileOut := logger.FileWriter(); fileOut != io.Discard {
				middlewares = append(middlewares, customTextLogger(colors.NewANSIStripWriter(fileOut)))
			}
		}
	}
	// Recovery writers
	if viper.GetString(config.LogFormat) == "json" {
		if out := joinWriters(appendWriterIfEnabled(viper.GetBool(config.LogToConsole), os.Stdout), logger.PanicFileWriter()); out != nil {
			middlewares = append(middlewares, customJSONRecovery(out))
		}
	} else {
		if viper.GetBool(config.LogToConsole) {
			middlewares = append(middlewares, gin.RecoveryWithWriter(os.Stdout))
		}
		if fileOut := logger.FileWriter(); fileOut != io.Discard {
			middlewares = append(middlewares, gin.RecoveryWithWriter(colors.NewANSIStripWriter(fileOut)))
		}
	}
	return middlewares
}

func customTextLogger(out io.Writer) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			ipAddr, referer := extractIPAndReferer(param)
			requestID, _ := param.Keys["request_id"].(string)
			path := param.Path
			if utf8.RuneCountInString(path) > pathMaxWidth {
				path = string([]rune(path)[:pathMaxWidth-3]) + "..."
			}
			if utf8.RuneCountInString(path) > pathWidth {
				return twoLinedAccessLog(param, path)
			}
			return fmt.Sprintf("%s %s [%s]  %7s %-50s | %s | %12v | %-15s | %s%s\n",
				param.TimeStamp.Format("2006/01/02 15:04:05"),
				text.Purple("GIN  "),
				text.Gray(requestID),
				param.Method,
				path,
				mapHTTPCodeToColor(param.StatusCode, strconv.Itoa(param.StatusCode)),
				param.Latency,
				ipAddr,
				formatUserAgent(param.Request.UserAgent()),
				referer,
			)
		},
		Output: out,
	})
}

func customJSONLogger(out io.Writer) gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			b, _ := json.Marshal(map[string]any{
				"time":      param.TimeStamp.Format(time.RFC3339),
				"method":    param.Method,
				"path":      param.Path,
				"status":    param.StatusCode,
				"latency":   param.Latency.String(),
				"client_ip": param.ClientIP,
				"ua":        param.Request.UserAgent(),
			})
			return string(b) + "\n"
		},
		Output: out,
	})
}

func customJSONRecovery(out io.Writer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				b, _ := json.Marshal(map[string]any{
					"time":       time.Now().Format(time.RFC3339),
					"level":      "ERROR",
					"message":    "panic recovered",
					"method":     ctx.Request.Method,
					"path":       ctx.Request.URL.RequestURI(),
					"client_ip":  ctx.ClientIP(),
					"user_agent": ctx.Request.UserAgent(),
					"panic":      fmt.Sprint(rec),
					"stack":      string(debug.Stack()),
				})
				_, _ = out.Write(append(b, '\n'))
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		ctx.Next()
	}
}

func appendWriterIfEnabled(enabled bool, w io.Writer) io.Writer {
	if !enabled {
		return nil
	}
	return w
}

func joinWriters(writers ...io.Writer) io.Writer {
	filtered := make([]io.Writer, 0, len(writers))
	for _, w := range writers {
		if w != nil && w != io.Discard {
			filtered = append(filtered, w)
		}
	}
	if len(filtered) == 0 {
		return nil
	}
	return io.MultiWriter(filtered...)
}

func twoLinedAccessLog(param gin.LogFormatterParams, path string) string {
	ipAddr, referer := extractIPAndReferer(param)
	requestID, _ := param.Keys["request_id"].(string)
	coloredPrefix := fmt.Sprintf("%s %s [%s] %7s ",
		param.TimeStamp.Format("2006/01/02 15:04:05"),
		text.Purple("GIN  "),
		text.Gray(requestID),
		param.Method)
	plainPrefix := fmt.Sprintf("%s GIN   | %7s ",
		param.TimeStamp.Format("2006/01/02 15:04:05"),
		param.Method,
	)
	visiblePrefixLen := utf8.RuneCountInString(plainPrefix)
	firstLine := coloredPrefix + path + "\n"
	secondLine := fmt.Sprintf("%s| %s | %12v | %-15s | %s%s\n",
		strings.Repeat(" ", visiblePrefixLen+pathWidth+1),
		mapHTTPCodeToColor(param.StatusCode, strconv.Itoa(param.StatusCode)),
		param.Latency,
		ipAddr,
		formatUserAgent(param.Request.UserAgent()),
		referer,
	)
	return firstLine + secondLine
}

func mapHTTPCodeToColor(code int, message string) string {
	switch {
	case code >= 200 && code < 300:
		return text.Green(message)
	case code >= 400 && code < 500:
		return text.Yellow(message)
	case code >= 500:
		return text.Red(message)
	default:
		return message
	}
}
