package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	primaryColor = lipgloss.Color("#10b981") // CWAT Green / Cyan
	secondaryClr = lipgloss.Color("#60A5FA") // Ocean blue theme
	mutedColor   = lipgloss.Color("#737373")
	blueColor    = lipgloss.Color("#60A5FA")

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().Foreground(primaryColor)
	mutedStyle  = lipgloss.NewStyle().Foreground(mutedColor)
	boldPrimary = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
)

const CWATRobot = `   ▄     ▄     
  █ ▀▄▄▄▀ █    
 █ ▀▄   ▄▀ █   
█  ▀  ▄  ▀  █  
 ▀   ▀ ▀   ▀   `

func GetTermWidth() int {
	fd := int(os.Stdout.Fd())
	width, _, err := term.GetSize(fd)
	if err != nil {
		return 100 // fallback
	}
	return width
}

func getCwdSafe() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}

func PrintDashboard(modelName string) {
	if !UiEnabled() {
		fmt.Println("Welcome back dev! (CWAT)")
		return
	}

	termW := GetTermWidth()
	dashWidth := termW - 4
	if dashWidth > 110 {
		dashWidth = 110
	}
	if dashWidth < 60 {
		dashWidth = 60
	}

	leftWidth := (dashWidth * 4) / 10
	rightWidth := dashWidth - leftWidth - 5
	if rightWidth < 10 {
		rightWidth = 10
	}

	// === LEFT COLUMN ===
	welcomeMsg := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Center).Render(boldPrimary.Render("Welcome back dev!"))

	mascotLines := strings.Split(CWATRobot, "\n")
	var centeredMascot []string
	for _, l := range mascotLines {
		renderedLine := lipgloss.NewStyle().Foreground(secondaryClr).Render(l)
		centeredMascot = append(centeredMascot, lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Center).Render(renderedLine))
	}
	mascotBlock := strings.Join(centeredMascot, "\n")

	infoLine := fmt.Sprintf("%s · CWAT Pro", modelName)
	cwd := getCwdSafe()
	if lipgloss.Width(cwd) > leftWidth-4 {
		cwd = "..." + cwd[len(cwd)-(leftWidth-7):]
	}

	infoBlock := lipgloss.NewStyle().Width(leftWidth).Align(lipgloss.Center).Foreground(mutedColor).Render(
		lipgloss.JoinVertical(lipgloss.Center,
			infoLine,
			"dev's Organization",
			cwd,
		),
	)

	leftContent := lipgloss.JoinVertical(lipgloss.Center, welcomeMsg, "", mascotBlock, "", infoBlock)

	// === RIGHT COLUMN ===
	tipsHeader := headerStyle.Render("Tips for getting started")
	tipsText := "Run /init to create a CWAT.md file with instructions for cwat"
	noteText := "Note: You have launched cwat in your home directory. For the best experience, laun..."

	rightTop := lipgloss.JoinVertical(lipgloss.Left,
		tipsHeader,
		tipsText,
		mutedStyle.Render(noteText),
	)

	activityHeader := headerStyle.Render("Recent activity")
	activityText := "1w ago  hello\n" + mutedStyle.Render("/resume for more")

	rightBot := lipgloss.JoinVertical(lipgloss.Left,
		activityHeader,
		activityText,
	)

	divider := mutedStyle.Render(strings.Repeat("-", rightWidth))
	rightContent := lipgloss.JoinVertical(lipgloss.Left, rightTop, divider, rightBot)

	// === JOIN & BORDER ===
	leftBox := lipgloss.NewStyle().Width(leftWidth).PaddingRight(2).Render(leftContent)
	rightBox := lipgloss.NewStyle().Width(rightWidth).BorderLeft(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(primaryColor).PaddingLeft(2).Render(rightContent)

	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftBox, rightBox)
	finalBox := boxStyle.Width(dashWidth).Render(layout)

	// Tag label ("CWAT Code v2.1.101")
	label := boldPrimary.Render(" CWAT Code ") + mutedStyle.Render("v2.1.101 ")

	// Create horizontal line extending before and after tag
	dashPrefix := lipgloss.NewStyle().Foreground(primaryColor).Render("───")
	dashSuffix := lipgloss.NewStyle().Foreground(primaryColor).Render(strings.Repeat("─", dashWidth-lipgloss.Width(label)-2))

	fmt.Println()
	fmt.Printf("%s%s%s\n", dashPrefix, label, dashSuffix)
	fmt.Println(finalBox)
	fmt.Println()
}

func PrintStatusLine() {
	termW := GetTermWidth()
	line := lipgloss.NewStyle().Foreground(mutedColor).Render(strings.Repeat("─", termW))
	fmt.Println(line)
}

func GetStatusBar(status string) string {
	termW := GetTermWidth()
	shortcuts := mutedStyle.Render("? for shortcuts")

	totalLen := lipgloss.Width(shortcuts) + lipgloss.Width(status)
	if totalLen+2 > termW {
		return status
	}

	padding := termW - totalLen
	spacer := strings.Repeat(" ", padding)

	return shortcuts + spacer + status
}

func PrintPrompt(label string) {
	fmt.Printf("%s\n\n", mutedStyle.Render("─────────────────────────────────────────────────────────────────────────────"))
}
