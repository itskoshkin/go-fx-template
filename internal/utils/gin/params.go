package ginutils

import (
	"github.com/gin-gonic/gin"
)

func extractIPAndReferer(param gin.LogFormatterParams) (ipAddr string, referer string) {
	ipAddr = param.ClientIP
	if realIP := param.Request.Header.Get("X-Real-IP"); realIP != "" {
		ipAddr = realIP
	}

	referer = param.Request.Referer()
	if referer != "" {
		referer = " | Referer \"" + referer + "\""
	}

	return
}
