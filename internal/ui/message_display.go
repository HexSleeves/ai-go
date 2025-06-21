package ui

import (
	"strings"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/log"
)

// MessagePanel handles the display of game messages
type MessagePanel struct {
	*Panel
	scrollOffset int // How many lines scrolled up from bottom
	maxMessages  int // Maximum number of messages to keep in memory
}

// NewMessagePanel creates a new message panel
func NewMessagePanel() *MessagePanel {
	panel := NewPanel(
		config.MessageLogX,
		config.MessageLogY,
		config.MessageLogWidth,
		config.MessageLogHeight,
		"Messages",
		true,
	)

	return &MessagePanel{
		Panel:        panel,
		scrollOffset: 0,
		maxMessages:  100, // Keep last 100 messages
	}
}

// Render draws the message panel with recent messages
func (mp *MessagePanel) Render(grid gruid.Grid, messageLog *log.MessageLog) {
	// Clear and draw border
	mp.Clear(grid)
	mp.DrawBorder(grid)

	// Get content area
	contentX, contentY, contentWidth, contentHeight := mp.GetContentArea()

	if messageLog == nil || len(messageLog.Messages) == 0 {
		// Show "No messages" if empty
		style := gruid.Style{Fg: ColorUIText, Bg: ColorUIBackground}
		mp.drawTextAt(grid, "No messages yet...", contentX, contentY, style)
		return
	}

	// Get wrapped message lines
	wrappedLines := mp.getWrappedMessageLines(messageLog, contentWidth)

	// Calculate which lines to display based on scroll offset
	totalLines := len(wrappedLines)
	startLine := max(totalLines-contentHeight-mp.scrollOffset, 0)

	// Draw messages from bottom up
	currentY := contentY + contentHeight - 1
	for i := startLine + contentHeight - 1; i >= startLine && currentY >= contentY; i-- {
		if i < len(wrappedLines) {
			line := wrappedLines[i]
			mp.drawTextAt(grid, line.Text, contentX, currentY,
				gruid.Style{Fg: line.Color, Bg: ColorUIBackground})
		}
		currentY--
	}

	// Draw scroll indicator if not at bottom
	if mp.scrollOffset > 0 {
		mp.drawScrollIndicator(grid, contentX+contentWidth-1, contentY)
	}
}

// MessageLine represents a wrapped line of text with color
type MessageLine struct {
	Text  string
	Color gruid.Color
}

// getWrappedMessageLines converts messages to wrapped lines for display
func (mp *MessagePanel) getWrappedMessageLines(messageLog *log.MessageLog, width int) []MessageLine {
	var lines []MessageLine

	// Process messages in order (oldest first)
	for _, msg := range messageLog.Messages {
		wrappedText := mp.wrapMessageText(msg.Text, width)
		for _, line := range wrappedText {
			lines = append(lines, MessageLine{
				Text:  line,
				Color: msg.Color,
			})
		}
	}

	return lines
}

// wrapMessageText wraps a message to fit within the specified width
func (mp *MessagePanel) wrapMessageText(text string, width int) []string {
	if width <= 0 {
		return []string{}
	}

	// Handle empty text
	if strings.TrimSpace(text) == "" {
		return []string{""}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// Check if adding this word would exceed width
		proposedLength := currentLine.Len()
		if proposedLength > 0 {
			proposedLength++ // for space
		}
		proposedLength += len(word)

		if proposedLength > width && currentLine.Len() > 0 {
			// Start new line
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		// Add word to current line
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	// Add final line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// drawTextAt draws text at a specific position
func (mp *MessagePanel) drawTextAt(grid gruid.Grid, text string, x, y int, style gruid.Style) {
	for i, r := range text {
		if x+i >= mp.X+mp.Width-1 { // Don't draw over border
			break
		}
		if x+i < grid.Size().X && y < grid.Size().Y {
			grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{Rune: r, Style: style})
		}
	}
}

// drawScrollIndicator shows that there are more messages above
func (mp *MessagePanel) drawScrollIndicator(grid gruid.Grid, x, y int) {
	style := gruid.Style{Fg: ColorUIHighlight, Bg: ColorUIBackground}
	if x < grid.Size().X && y < grid.Size().Y {
		grid.Set(gruid.Point{X: x, Y: y}, gruid.Cell{Rune: '▲', Style: style})
	}
}

// ScrollUp scrolls the message log up (shows older messages)
func (mp *MessagePanel) ScrollUp(messageLog *log.MessageLog) {
	if messageLog == nil {
		return
	}

	_, _, contentWidth, contentHeight := mp.GetContentArea()
	wrappedLines := mp.getWrappedMessageLines(messageLog, contentWidth)

	maxScroll := len(wrappedLines) - contentHeight
	if maxScroll > 0 {
		mp.scrollOffset++
		if mp.scrollOffset > maxScroll {
			mp.scrollOffset = maxScroll
		}
	}
}

// ScrollDown scrolls the message log down (shows newer messages)
func (mp *MessagePanel) ScrollDown() {
	if mp.scrollOffset > 0 {
		mp.scrollOffset--
	}
}

// ScrollToBottom scrolls to the most recent messages
func (mp *MessagePanel) ScrollToBottom() {
	mp.scrollOffset = 0
}

// IsAtBottom returns true if showing the most recent messages
func (mp *MessagePanel) IsAtBottom() bool {
	return mp.scrollOffset == 0
}

// GetScrollOffset returns the current scroll offset
func (mp *MessagePanel) GetScrollOffset() int {
	return mp.scrollOffset
}
