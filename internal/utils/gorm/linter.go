package gormutils

import (
	"regexp"
	"strings"

	"go-fx-template/internal/utils/text"
)

var leadingSQLVerbRegex = regexp.MustCompile(`(?i)^\s*(SELECT|UPDATE|DELETE|INSERT|WITH)\b`)

func highlightSQL(sql string) string {
	return leadingSQLVerbRegex.ReplaceAllStringFunc(sql, func(s string) string {
		switch strings.ToUpper(strings.TrimSpace(s)) {
		case "INSERT":
			return text.Blue("INSERT")
		case "SELECT":
			return text.Green("SELECT")
		case "UPDATE":
			return text.Yellow("UPDATE")
		case "DELETE":
			return text.Red("DELETE")
		default:
			return s
		}
	})
}
