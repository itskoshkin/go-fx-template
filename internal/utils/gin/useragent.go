package ginutils

import (
	"strings"
)

func formatUserAgent(ua string) string {
	if ua == "" {
		return ""
	}

	platform := ""
	rest := ua
	const moz = "Mozilla/5.0 ("
	if i := strings.Index(ua, moz); i != -1 {
		start := i + len(moz)
		if j := strings.Index(ua[start:], ")"); j != -1 {
			platform = ua[start : start+j]
			rest = ua[start+j+1:]
		}
	} else if i := strings.Index(ua, "("); i != -1 {
		start := i + 1
		if j := strings.Index(ua[start:], ")"); j != -1 {
			platform = ua[start : start+j]
			rest = ua[start+j+1:]
		}
	}

	keep := make([]string, 0, 4)
	for _, t := range strings.Fields(rest) {
		pt := strings.Trim(t, " ;,()")
		if pt == "" {
			continue
		}
		if strings.HasPrefix(pt, "AppleWebKit/") ||
			strings.EqualFold(pt, "KHTML") ||
			strings.EqualFold(pt, "like") ||
			strings.HasPrefix(pt, "Gecko") ||
			strings.HasPrefix(pt, "Version/") ||
			strings.HasPrefix(pt, "Mobile/") {
			continue
		}
		if strings.HasPrefix(pt, "Chrome/") || strings.HasPrefix(pt, "CriOS/") || strings.HasPrefix(pt, "Chromium/") ||
			strings.HasPrefix(pt, "Edg/") || strings.HasPrefix(pt, "EdgiOS/") || strings.HasPrefix(pt, "EdgA/") ||
			strings.HasPrefix(pt, "OPR/") || strings.HasPrefix(pt, "Opera/") ||
			strings.HasPrefix(pt, "Safari/") ||
			strings.HasPrefix(pt, "Firefox/") || strings.HasPrefix(pt, "FxiOS/") ||
			strings.HasPrefix(pt, "YaBrowser/") || strings.HasPrefix(pt, "Vivaldi/") ||
			strings.HasPrefix(pt, "SamsungBrowser/") || strings.HasPrefix(pt, "DuckDuckGo/") {
			keep = append(keep, pt)
		}
	}

	if platform == "" && len(keep) == 0 {
		return ua
	}
	if platform != "" && len(keep) > 0 {
		return platform + " " + strings.Join(keep, " ")
	}
	if platform != "" {
		return platform
	}
	return strings.Join(keep, " ")
}
