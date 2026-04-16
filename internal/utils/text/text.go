package text

import (
	"go-fx-template/internal/utils/colors"
)

// Red wraps text in red color codes.
func Red(text string) string { return colors.Red() + text + colors.Reset() }

// Orange wraps text in orange/yellow color codes.
func Orange(text string) string { return colors.Orange() + text + colors.Reset() }

// Yellow wraps text in yellow color codes.
func Yellow(text string) string { return colors.Yellow() + text + colors.Reset() }

// Green wraps text in green color codes.
func Green(text string) string { return colors.Green() + text + colors.Reset() }

// Cyan wraps text in cyan color codes.
func Cyan(text string) string { return colors.Cyan() + text + colors.Reset() }

// Sky wraps text in sky blue color codes.
func Sky(text string) string { return colors.Sky() + text + colors.Reset() }

// Blue wraps text in blue color codes.
func Blue(text string) string { return colors.Blue() + text + colors.Reset() }

// Purple wraps text in purple color codes.
func Purple(text string) string { return colors.Purple() + text + colors.Reset() }

// Magenta wraps text in magenta color codes.
func Magenta(text string) string { return colors.Magenta() + text + colors.Reset() }

// White wraps text in white color codes.
func White(text string) string { return colors.White() + text + colors.Reset() }

// Gray wraps text in gray color codes.
func Gray(text string) string { return colors.Gray() + text + colors.Reset() }

// Black wraps text in black color codes.
func Black(text string) string { return colors.Black() + text + colors.Reset() }

// Bold wraps text in bold style codes.
func Bold(text string) string { return colors.Bold() + text + colors.Reset() }

// Italic wraps text in italic style codes.
func Italic(text string) string { return colors.Italic() + text + colors.Reset() }

// Underline wraps text in underline style codes.
func Underline(text string) string { return colors.Underline() + text + colors.Reset() }

// Strikethrough wraps text in strikethrough style codes.
func Strikethrough(text string) string { return colors.Strikethrough() + text + colors.Reset() }

// Background wraps text in background/reverse style codes.
func Background(text string) string { return colors.Background() + text + colors.Reset() }
