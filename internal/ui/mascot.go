package ui

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

const mascotB64 = `Pnc8CiAgIOKWhCAgICAg4paEICAgICAgICDiloQgICAgIOKWhCAgICAgICAg4paEICAgICDiloQgICAKICDilogg4paA4paE4paE4paE4paAIOKWiCAgICAgIOKWiCDiloDiloTiloTiloTiloAg4paIICAgICAg4paIIOKWgOKWhOKWhOKWhOKWgCDiloggIAog4paIIOKWgOKWhCAgIOKWhOKWgCDiloggICAg4paIIOKWhOKWhCAgIOKWhOKWhCDiloggICAg4paIICAgICAgIOKWhCDiloggCuKWiCAg4paAICDiloQgIOKWgCAg4paIICDiloggICAgIOKWhCAgICAg4paIICDiloggIOKWhCAg4paEICAgICDilogKIOKWgCAgIOKWgCDiloAgICDiloAgICAg4paAICAg4paAIOKWgCAgIOKWgCAgICDiloAgICDiloAg4paAICAg4paAIAogICDiloQgICAgIOKWhCAgICAgICAg4paEICAgICDiloQgICAgICAgIOKWhCAgICAg4paEICAgCiAg4paIIOKWgOKWhOKWhOKWhOKWgCDiloggICAgICDilogg4paA4paE4paE4paE4paAIOKWiCAgICAgIOKWiCDiloDiloTiloTiloTiloAg4paIICAKIOKWiCDiloQg4paEIOKWhCDiloQg4paIICAgIOKWiCAg4paEICAgIOKWhCDiloggICAg4paIICDiloQgICDiloQgIOKWiCAK4paIICDiloTiloDiloQg4paE4paA4paEICDiloggIOKWiCAgIOKWgCDiloQgIOKWgCAg4paIICDiloggICDiloAg4paEIOKWgCAgIOKWiAog4paAICAgICAgICAg4paAICAgIOKWgCAgIOKWgCDiloAgICDiloAgICAg4paAICAg4paAIOKWgCAgIOKWgCA=`

func parseMascotFrames() [][]string {
	decoded, _ := base64.StdEncoding.DecodeString(mascotB64)
	lines := strings.Split(string(decoded), "\n")

	mascotFrames := make([][]string, 6)
	for i := 0; i < 6; i++ {
		mascotFrames[i] = make([]string, 5)
	}

	for row := 0; row < 2; row++ {
		for col := 0; col < 3; col++ {
			frameIdx := row*3 + col
			for lineIdx := 0; lineIdx < 5; lineIdx++ {
				actualLine := 1 + row*5 + lineIdx
				if actualLine >= len(lines) {
					continue
				}

				l := strings.TrimRight(lines[actualLine], "\r")
				runes := []rune(l)

				start := col * 15
				end := start + 15
				if start >= len(runes) {
					continue
				}
				if end > len(runes) {
					end = len(runes)
				}
				mascotFrames[frameIdx][lineIdx] = string(runes[start:end])
			}
		}
	}
	return mascotFrames
}

// UiMascot plays a short startup idle animation then lands on the main frame.
func UiMascot() {
	if !UiEnabled() {
		// Just print a static version if UI is not enabled (e.g. no ANSI)
		frames := parseMascotFrames()
		for _, line := range frames[0] {
			fmt.Println(line)
		}
		fmt.Println()
		return
	}

	frames := parseMascotFrames()

	// Initial print of 5 lines
	for _, line := range frames[0] {
		fmt.Println(UiPaint(ansiCyan, line))
	}

	// Play idle animation (about 2.5 seconds)
	// We blink or change expressions several times.
	idleSequence := []struct {
		frame    int
		duration time.Duration
	}{
		{0, 600 * time.Millisecond},
		{2, 300 * time.Millisecond}, // wink
		{0, 600 * time.Millisecond},
		{1, 400 * time.Millisecond}, // happy
		{0, 500 * time.Millisecond},
		{4, 400 * time.Millisecond}, // surprise
		{0, 200 * time.Millisecond},
	}

	for _, step := range idleSequence {
		time.Sleep(step.duration)
		fmt.Print("\033[5A") // Move cursor up 5 lines
		for _, line := range frames[step.frame] {
			// using ANSI clear to end of line just in case \033[K
			fmt.Println("\033[K" + UiPaint(ansiCyan, line))
		}
	}
	fmt.Println()
}
