package colors

import (
	"io"
	"regexp"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func stripANSI(text string) string {
	return ansiRegexp.ReplaceAllString(text, "")
}

type ansiStripWriter struct {
	target io.Writer
}

func (w ansiStripWriter) Write(p []byte) (int, error) {
	if w.target == nil {
		return len(p), nil
	}

	clean := stripANSI(string(p))
	if _, err := io.WriteString(w.target, clean); err != nil {
		return 0, err
	}

	return len(p), nil
}

func NewANSIStripWriter(target io.Writer) io.Writer {
	if target == nil {
		return io.Discard
	}
	return ansiStripWriter{target: target}
}
