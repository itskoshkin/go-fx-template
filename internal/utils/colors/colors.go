package colors

import (
	"os"
)

const (
	regularWhite  = "\033[37m"
	regularBlack  = "\033[30m"
	regularRed    = "\033[31m"
	regularYellow = "\033[33m"
	regularGreen  = "\033[32m"
	regularCyan   = "\033[36m"
	regularBlue   = "\033[34m"
	regularPurple = "\033[35m"

	highlightedWhite  = "\033[97m"
	highlightedBlack  = "\033[90m"
	highlightedRed    = "\033[91m"
	highlightedYellow = "\033[93m"
	highlightedGreen  = "\033[92m"
	highlightedCyan   = "\033[96m"
	highlightedBlue   = "\033[94m"
	highlightedPurple = "\033[95m"

	constWhite     = highlightedWhite
	constLightGray = regularWhite
	constDarkGray  = highlightedBlack
	constBlack     = regularBlack
	constRed       = highlightedRed
	constOrange    = regularRed
	constYellow    = highlightedYellow
	constGreen     = highlightedGreen
	constDarkGreen = regularGreen
	constCyan      = regularCyan
	constSky       = highlightedCyan
	constBlue      = highlightedBlue
	constDarkBlue  = regularBlue
	constMagenta   = highlightedPurple
	constPurple    = regularPurple

	constBold          = "\033[1m"
	constFaint         = "\033[2m"
	constItalic        = "\033[3m"
	constUnderline     = "\033[4m"
	constBlink         = "\033[5m"
	constBackground    = "\033[7m"
	constHidden        = "\033[8m"
	constStrikethrough = "\033[9m"
	constReset         = "\033[0m"

	Nl  = "\n"            // New line
	Pl  = "\033[F"        // Previous line
	Cpl = "\033[A\033[2K" // Clear previous line
)

const (
	rgbRed      = "\033[38;2;255;0;0m"
	rgbOrange   = "\033[38;2;255;127;0m"
	rgbGreen    = "\033[38;2;0;255;0m"
	rgbSkyBlue  = "\033[38;2;0;191;255m"
	rgbDarkBlue = "\033[38;2;0;0;255m"
	rgbViolet   = "\033[38;2;148;0;211m"
	rgbPink     = "\033[38;2;255;105;180m"
	rgbBrown    = "\033[38;2;139;69;19m"
	rgbGold     = "\033[38;2;255;215;0m"
	rgbMint     = "\033[38;2;152;251;152m"
)

// IsRGBSupported reports whether truecolor/24-bit ANSI colors are likely supported
// noinspection SpellCheckingInspection
func IsRGBSupported() bool {
	val := os.Getenv("COLORTERM")
	return val == "truecolor" || val == "24bit"
}

// Colors — return raw ANSI escape code for use with fmt.Sprintf("%s...", colors.Red())

func Red() string {
	if IsRGBSupported() {
		return rgbRed
	}
	return constRed
}

func Orange() string {
	if IsRGBSupported() {
		return rgbOrange
	}
	return constYellow
}

func Yellow() string { return constYellow }
func Green() string {
	if IsRGBSupported() {
		return rgbGreen
	}
	return constGreen
}
func Sky() string       { return constSky }
func Blue() string      { return constBlue }
func Purple() string    { return constPurple }
func Magenta() string   { return constMagenta }
func White() string     { return constWhite }
func Gray() string      { return constLightGray }
func Black() string     { return constBlack }
func DarkGreen() string { return constDarkGreen }
func Cyan() string      { return constCyan }

// Styles

func Bold() string          { return constBold }
func Faint() string         { return constFaint }
func Italic() string        { return constItalic }
func Underline() string     { return constUnderline }
func Blink() string         { return constBlink }
func Background() string    { return constBackground }
func Hidden() string        { return constHidden }
func Strikethrough() string { return constStrikethrough }

// Reset

func Reset() string { return constReset }
