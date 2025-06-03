package ui

import (
	"fmt"
	"time"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/log"
)

// FullMessageScreen handles the full-screen message log display
type FullMessageScreen struct {
	*Panel
	scrollOffset int
	maxMessages  int
}

// NewFullMessageScreen creates a new full message screen
func NewFullMessageScreen() *FullMessageScreen {
	panel := NewPanel(
		0, 0,
		config.DungeonWidth,
		config.DungeonHeight,
		"Message History",
		true,
	)

	return &FullMessageScreen{
		Panel:        panel,
		scrollOffset: 0,
		maxMessages:  200, // Keep more messages for full screen
	}
}

// Render draws the full message screen with extended message history
func (fms *FullMessageScreen) Render(grid gruid.Grid, messageLog *log.MessageLog) {
	// Clear and draw border
	fms.Clear(grid)
	fms.DrawBorder(grid)

	// Get content area
	contentX, contentY, contentWidth, contentHeight := fms.GetContentArea()

	if messageLog == nil || len(messageLog.Messages) == 0 {
		// Show "No messages" if empty
		fms.drawText(grid, "No messages yet...", contentX, contentY, ColorUIText)
		fms.drawInstructions(grid)
		return
	}

	// Get wrapped message lines with timestamps
	wrappedLines := fms.getWrappedMessageLines(messageLog, contentWidth)

	// Calculate which lines to display based on scroll offset
	totalLines := len(wrappedLines)
	displayHeight := contentHeight - 1 // Reserve space for instructions

	startLine := max(totalLines-displayHeight-fms.scrollOffset, 0)
	endLine := min(startLine+displayHeight, totalLines)

	// Draw messages from top to bottom
	currentY := contentY
	for i := startLine; i < endLine && currentY < contentY+displayHeight; i++ {
		line := wrappedLines[i]
		fms.drawText(grid, line.Text, contentX, currentY, line.Color)
		currentY++
	}

	// Draw scroll indicators
	fms.drawScrollIndicators(grid, contentX, contentY, contentWidth, displayHeight, startLine, endLine, totalLines)

	// Instructions at bottom
	fms.drawInstructions(grid)
}

// MessageLineWithTime represents a wrapped line of text with color and timestamp info
type MessageLineWithTime struct {
	Text      string
	Color     gruid.Color
	IsNewMsg  bool // True if this is the start of a new message
	Timestamp string
}

// getWrappedMessageLines converts messages to wrapped lines for display with timestamps
func (fms *FullMessageScreen) getWrappedMessageLines(messageLog *log.MessageLog, width int) []MessageLineWithTime {
	var lines []MessageLineWithTime

	// Reserve space for timestamp prefix
	timestampWidth := 12 // "[HH:MM:SS] "
	textWidth := width - timestampWidth
	if textWidth < 20 {
		textWidth = width // Fallback if screen too narrow
		timestampWidth = 0
	}

	// Process messages in order (oldest first)
	for _, msg := range messageLog.Messages {
		// Format timestamp
		timestamp := ""
		if timestampWidth > 0 {
			timestamp = fms.formatTimestamp(msg.Timestamp)
		}

		// Wrap message text
		wrappedText := fms.wrapMessageText(msg.Text, textWidth)

		for i, line := range wrappedText {
			messageText := line
			if timestampWidth > 0 {
				if i == 0 {
					// First line gets timestamp
					messageText = fmt.Sprintf("%s %s", timestamp, line)
				} else {
					// Continuation lines get indentation
					messageText = fmt.Sprintf("%*s %s", timestampWidth-1, "", line)
				}
			}

			lines = append(lines, MessageLineWithTime{
				Text:      messageText,
				Color:     msg.Color,
				IsNewMsg:  i == 0,
				Timestamp: timestamp,
			})
		}
	}

	return lines
}

// formatTimestamp formats a timestamp for display
func (fms *FullMessageScreen) formatTimestamp(timestamp time.Time) string {
	return fmt.Sprintf("[%02d:%02d:%02d]",
		timestamp.Hour(), timestamp.Minute(), timestamp.Second())
}

// wrapMessageText wraps a message to fit within the specified width
func (fms *FullMessageScreen) wrapMessageText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	// Handle empty text
	if len(text) == 0 {
		return []string{""}
	}

	words := []string{}
	currentWord := ""

	// Split into words, preserving spaces
	for _, r := range text {
		if r == ' ' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
			words = append(words, " ")
		} else {
			currentWord += string(r)
		}
	}
	if currentWord != "" {
		words = append(words, currentWord)
	}

	var lines []string
	var currentLine string

	for _, word := range words {
		if len(currentLine)+len(word) <= width {
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}

			// Handle very long words
			if len(word) > width {
				for len(word) > width {
					lines = append(lines, word[:width])
					word = word[width:]
				}
				currentLine = word
			} else {
				currentLine = word
			}
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		lines = []string{""}
	}

	return lines
}

// drawScrollIndicators shows scroll position and availability
func (fms *FullMessageScreen) drawScrollIndicators(grid gruid.Grid, x, y, width, height, startLine, endLine, totalLines int) {
	// Scroll up indicator
	if startLine > 0 {
		indicator := fmt.Sprintf("▲ %d more above", startLine)
		fms.drawText(grid, indicator, x, y, ColorUIHighlight)
	}

	// Scroll down indicator
	if endLine < totalLines {
		remaining := totalLines - endLine
		indicator := fmt.Sprintf("▼ %d more below", remaining)
		fms.drawText(grid, indicator, x, y+height-1, ColorUIHighlight)
	}

	// Scroll position indicator (right side)
	if totalLines > height {
		scrollPercent := (startLine * 100) / (totalLines - height)
		if scrollPercent > 100 {
			scrollPercent = 100
		}
		positionText := fmt.Sprintf("%d%%", scrollPercent)
		fms.drawText(grid, positionText, x+width-len(positionText), y, ColorUIHighlight)
	}
}

// drawInstructions renders control instructions at the bottom
func (fms *FullMessageScreen) drawInstructions(grid gruid.Grid) {
	instructionY := fms.Y + fms.Height - 2
	instructions := "↑↓/jk/PgUp/PgDn: Scroll | Home: Top | End: Bottom | [ESC]/q: Close"

	// Center the instructions
	startX := fms.X + (fms.Width-len(instructions))/2
	if startX < fms.X+1 {
		startX = fms.X + 1
	}

	fms.drawText(grid, instructions, startX, instructionY, ColorUIHighlight)
}

// ScrollUp scrolls the message log up (shows older messages)
func (fms *FullMessageScreen) ScrollUp(messageLog *log.MessageLog, lines int) {
	if messageLog == nil {
		return
	}

	_, _, contentWidth, contentHeight := fms.GetContentArea()
	wrappedLines := fms.getWrappedMessageLines(messageLog, contentWidth)

	maxScroll := len(wrappedLines) - (contentHeight - 1)
	if maxScroll > 0 {
		fms.scrollOffset += lines
		if fms.scrollOffset > maxScroll {
			fms.scrollOffset = maxScroll
		}
	}
}

// ScrollDown scrolls the message log down (shows newer messages)
func (fms *FullMessageScreen) ScrollDown(lines int) {
	fms.scrollOffset -= lines
	if fms.scrollOffset < 0 {
		fms.scrollOffset = 0
	}
}

// ScrollToTop scrolls to the oldest messages
func (fms *FullMessageScreen) ScrollToTop(messageLog *log.MessageLog) {
	if messageLog == nil {
		return
	}

	_, _, contentWidth, contentHeight := fms.GetContentArea()
	wrappedLines := fms.getWrappedMessageLines(messageLog, contentWidth)

	maxScroll := len(wrappedLines) - (contentHeight - 1)
	if maxScroll > 0 {
		fms.scrollOffset = maxScroll
	}
}

// ScrollToBottom scrolls to the most recent messages
func (fms *FullMessageScreen) ScrollToBottom() {
	fms.scrollOffset = 0
}

// IsAtBottom returns true if showing the most recent messages
func (fms *FullMessageScreen) IsAtBottom() bool {
	return fms.scrollOffset == 0
}

// GetScrollOffset returns the current scroll offset
func (fms *FullMessageScreen) GetScrollOffset() int {
	return fms.scrollOffset
}

// drawText draws text at the specified position
func (fms *FullMessageScreen) drawText(grid gruid.Grid, text string, x, y int, color gruid.Color) {
	style := gruid.Style{Fg: color, Bg: ColorUIBackground}

	for i, r := range text {
		if x+i >= fms.X+fms.Width-1 { // Don't draw over border
			break
		}
		if x+i < grid.Size().X && y < grid.Size().Y {
			grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{Rune: r, Style: style})
		}
	}
}
