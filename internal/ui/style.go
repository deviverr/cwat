package ui

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiCyan   = "\033[36m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
)

func IsInteractiveTerminal() bool {
	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	stdoutInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stdinInfo.Mode()&os.ModeCharDevice) != 0 && (stdoutInfo.Mode()&os.ModeCharDevice) != 0
}

func UiEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb") {
		return false
	}
	return IsInteractiveTerminal()
}

func UiPaint(code, text string) string {
	if !UiEnabled() {
		return text
	}
	return code + text + ansiReset
}

func UiAccent(text string) string {
	return UiPaint(ansiCyan, text)
}

func UiBold(text string) string {
	return UiPaint(ansiBold, text)
}

func UiMuted(text string) string {
	return UiPaint(ansiDim, text)
}

func UiSuccess(text string) string {
	return UiPaint(ansiGreen, text)
}

func UiWarn(text string) string {
	return UiPaint(ansiYellow, text)
}

func UiError(text string) string {
	return UiPaint(ansiRed, text)
}

func UiCommand(text string) string {
	return UiPaint(ansiGreen, text)
}

func UiDivider() {
	fmt.Println(UiAccent("------------------------------------------------------------"))
}

func UiHero(title, subtitle string) {
	fmt.Println()
	UiDivider()
	fmt.Printf("%s %s\n", UiAccent("CWAT"), UiBold(title))
	if strings.TrimSpace(subtitle) != "" {
		fmt.Println(UiMuted(subtitle))
	}
	UiDivider()
}

func UiStep(current, total int, title, subtitle string) {
	fmt.Println()
	fmt.Printf("%s %s\n", UiAccent(fmt.Sprintf("[%d/%d]", current, total)), UiBold(title))
	if strings.TrimSpace(subtitle) != "" {
		fmt.Println(UiMuted(subtitle))
	}
}

func UiOption(key, title, subtitle string) {
	fmt.Printf("  %s %s\n", UiAccent(key), title)
	if strings.TrimSpace(subtitle) != "" {
		fmt.Printf("    %s\n", UiMuted(subtitle))
	}
}

func UiPrompt(label string) string {
	return fmt.Sprintf("%s %s", UiAccent("->"), label)
}

func UiCommandBlock(command string) {
	fmt.Printf("  %s\n", UiCommand(command))
}

type Spinner struct {
	stop chan struct{}
	o    int32
}

func StartSpinner(message string) *Spinner {
	s := &Spinner{stop: make(chan struct{})}
	frames := []string{"◐", "◓", "◑", "◒"}
	if !UiEnabled() {
		fmt.Println(message)
		return s
	}
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K")
				return
			case <-ticker.C:
				fmt.Printf("\r  %s %s", UiMuted(frames[i%4]), UiMuted(message))
				i++
			}
		}
	}()
	return s
}

func (s *Spinner) Stop() {
	if atomic.CompareAndSwapInt32(&s.o, 0, 1) {
		close(s.stop)
	}
}
