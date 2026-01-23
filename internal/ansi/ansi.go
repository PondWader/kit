package ansi

import "fmt"

// Reset codes
const (
	Reset = "\033[0m"
)

// Foreground colors
func Black(s string) string {
	return fmt.Sprintf("\033[30m%s\033[0m", s)
}

func Red(s string) string {
	return fmt.Sprintf("\033[31m%s\033[0m", s)
}

func Green(s string) string {
	return fmt.Sprintf("\033[32m%s\033[0m", s)
}

func Yellow(s string) string {
	return fmt.Sprintf("\033[33m%s\033[0m", s)
}

func Blue(s string) string {
	return fmt.Sprintf("\033[34m%s\033[0m", s)
}

func Magenta(s string) string {
	return fmt.Sprintf("\033[35m%s\033[0m", s)
}

func Cyan(s string) string {
	return fmt.Sprintf("\033[36m%s\033[0m", s)
}

func White(s string) string {
	return fmt.Sprintf("\033[37m%s\033[0m", s)
}

// Bright foreground colors
func BrightBlack(s string) string {
	return fmt.Sprintf("\033[90m%s\033[0m", s)
}

func BrightRed(s string) string {
	return fmt.Sprintf("\033[91m%s\033[0m", s)
}

func BrightGreen(s string) string {
	return fmt.Sprintf("\033[92m%s\033[0m", s)
}

func BrightYellow(s string) string {
	return fmt.Sprintf("\033[93m%s\033[0m", s)
}

func BrightBlue(s string) string {
	return fmt.Sprintf("\033[94m%s\033[0m", s)
}

func BrightMagenta(s string) string {
	return fmt.Sprintf("\033[95m%s\033[0m", s)
}

func BrightCyan(s string) string {
	return fmt.Sprintf("\033[96m%s\033[0m", s)
}

func BrightWhite(s string) string {
	return fmt.Sprintf("\033[97m%s\033[0m", s)
}

// Background colors
func BgBlack(s string) string {
	return fmt.Sprintf("\033[40m%s\033[0m", s)
}

func BgRed(s string) string {
	return fmt.Sprintf("\033[41m%s\033[0m", s)
}

func BgGreen(s string) string {
	return fmt.Sprintf("\033[42m%s\033[0m", s)
}

func BgYellow(s string) string {
	return fmt.Sprintf("\033[43m%s\033[0m", s)
}

func BgBlue(s string) string {
	return fmt.Sprintf("\033[44m%s\033[0m", s)
}

func BgMagenta(s string) string {
	return fmt.Sprintf("\033[45m%s\033[0m", s)
}

func BgCyan(s string) string {
	return fmt.Sprintf("\033[46m%s\033[0m", s)
}

func BgWhite(s string) string {
	return fmt.Sprintf("\033[47m%s\033[0m", s)
}

// Bright background colors
func BgBrightBlack(s string) string {
	return fmt.Sprintf("\033[100m%s\033[0m", s)
}

func BgBrightRed(s string) string {
	return fmt.Sprintf("\033[101m%s\033[0m", s)
}

func BgBrightGreen(s string) string {
	return fmt.Sprintf("\033[102m%s\033[0m", s)
}

func BgBrightYellow(s string) string {
	return fmt.Sprintf("\033[103m%s\033[0m", s)
}

func BgBrightBlue(s string) string {
	return fmt.Sprintf("\033[104m%s\033[0m", s)
}

func BgBrightMagenta(s string) string {
	return fmt.Sprintf("\033[105m%s\033[0m", s)
}

func BgBrightCyan(s string) string {
	return fmt.Sprintf("\033[106m%s\033[0m", s)
}

func BgBrightWhite(s string) string {
	return fmt.Sprintf("\033[107m%s\033[0m", s)
}

// Text styles
func Bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

func Dim(s string) string {
	return fmt.Sprintf("\033[2m%s\033[0m", s)
}

func Italic(s string) string {
	return fmt.Sprintf("\033[3m%s\033[0m", s)
}

func Underline(s string) string {
	return fmt.Sprintf("\033[4m%s\033[0m", s)
}

func Blink(s string) string {
	return fmt.Sprintf("\033[5m%s\033[0m", s)
}

func Reverse(s string) string {
	return fmt.Sprintf("\033[7m%s\033[0m", s)
}

func Hidden(s string) string {
	return fmt.Sprintf("\033[8m%s\033[0m", s)
}

func Strikethrough(s string) string {
	return fmt.Sprintf("\033[9m%s\033[0m", s)
}

// 256-color support
func Color256(colorCode int, s string) string {
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", colorCode, s)
}

func BgColor256(colorCode int, s string) string {
	return fmt.Sprintf("\033[48;5;%dm%s\033[0m", colorCode, s)
}

func IsMarker(r rune) bool {
	return r == '\x1b'
}

func IsTerminator(r rune) bool {
	return (r >= 0x40 && r <= 0x5a) || (r == 0x5e) || (r >= 0x60 && r <= 0x7e)
}
