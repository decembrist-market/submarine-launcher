package internal

import (
	"fmt"
	"strings"
)

const BarWidth = 40
const (
	GreenColor = "\x1b[38;2;53;203;54m"
	ResetColor = "\x1b[0m"
	WarnColor  = "\x1b[38;2;255;193;7m"
	ErrorColor = "\x1b[38;2;244;67;54m"
	Info       = "Info"
	Warn       = "Warn"
	Error      = "Error"
	Success    = "Success"
)

func drawProgress(downloaded, total float64) {
	percent := int(downloaded / total * 100)
	if percent > 100 {
		percent = 100
	}
	filled := int(float64(BarWidth) * float64(percent) / 100.0)
	bar := GreenColor + strings.Repeat("█", filled) + ResetColor + strings.Repeat("█", BarWidth-filled)
	fmt.Printf("\r%s %3d%%", bar, percent)
}

func ShowExitMessage(level, message string) {
	if message != "" {
		ShowStyledMessage(level, message)
	}
	fmt.Println("Нажмите Enter для выхода...")
	fmt.Scanln()
}

func CheckAnswer() bool {
	var answer string
	_, err := fmt.Scanln(&answer)
	if err != nil {
		return false
	}
	return answer == "y" || answer == "Y" || answer == "н" || answer == "Н"
}
