package ui

import (
	"strings"
	"unicode/utf8"

	"codeberg.org/anaseto/gruid"
)

// Panel represents a UI panel with position and dimensions
type Panel struct {
	X, Y          int
	Width, Height int
	Title         string
	HasBorder     bool
}

// NewPanel creates a new panel with the specified parameters
func NewPanel(x, y, width, height int, title string, hasBorder bool) *Panel {
	return &Panel{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Title:     title,
		HasBorder: hasBorder,
	}
}

// Clear clears the panel area on the grid
func (p *Panel) Clear(grid gruid.Grid) {
	for y := p.Y; y < p.Y+p.Height; y++ {
		for x := p.X; x < p.X+p.Width; x++ {
			if x < grid.Size().X && y < grid.Size().Y {
				grid.Set(gruid.Point{X: x, Y: y}, gruid.Cell{Rune: ' ', Style: gruid.Style{Bg: ColorUIBackground}})
			}
		}
	}
}

// DrawBorder draws a border around the panel
func (p *Panel) DrawBorder(grid gruid.Grid) {
	if !p.HasBorder {
		return
	}

	borderStyle := gruid.Style{Fg: ColorUIBorder, Bg: ColorUIBackground}

	// Draw corners
	grid.Set(gruid.Point{X: p.X, Y: p.Y}, gruid.Cell{Rune: '┌', Style: borderStyle})
	grid.Set(gruid.Point{X: p.X + p.Width - 1, Y: p.Y}, gruid.Cell{Rune: '┐', Style: borderStyle})
	grid.Set(gruid.Point{X: p.X, Y: p.Y + p.Height - 1}, gruid.Cell{Rune: '└', Style: borderStyle})
	grid.Set(gruid.Point{X: p.X + p.Width - 1, Y: p.Y + p.Height - 1}, gruid.Cell{Rune: '┘', Style: borderStyle})

	// Draw horizontal borders
	for x := p.X + 1; x < p.X+p.Width-1; x++ {
		grid.Set(gruid.Point{X: x, Y: p.Y}, gruid.Cell{Rune: '─', Style: borderStyle})
		grid.Set(gruid.Point{X: x, Y: p.Y + p.Height - 1}, gruid.Cell{Rune: '─', Style: borderStyle})
	}

	// Draw vertical borders
	for y := p.Y + 1; y < p.Y+p.Height-1; y++ {
		grid.Set(gruid.Point{X: p.X, Y: y}, gruid.Cell{Rune: '│', Style: borderStyle})
		grid.Set(gruid.Point{X: p.X + p.Width - 1, Y: y}, gruid.Cell{Rune: '│', Style: borderStyle})
	}

	// Draw title if provided
	if p.Title != "" {
		p.DrawTitle(grid)
	}
}

// DrawTitle draws the panel title in the top border
func (p *Panel) DrawTitle(grid gruid.Grid) {
	if p.Title == "" || !p.HasBorder {
		return
	}

	titleStyle := gruid.Style{Fg: ColorUITitle, Bg: ColorUIBackground}
	title := " " + p.Title + " "

	// Truncate title if too long
	maxTitleWidth := p.Width - 4
	if utf8.RuneCountInString(title) > maxTitleWidth {
		title = title[:maxTitleWidth-3] + "..."
	}

	// Center the title
	startX := p.X + (p.Width-utf8.RuneCountInString(title))/2

	for i, r := range title {
		if startX+i < p.X+p.Width-1 {
			grid.Set(gruid.Point{X: startX + i, Y: p.Y}, gruid.Cell{Rune: r, Style: titleStyle})
		}
	}
}

// GetContentArea returns the area inside the panel (excluding borders)
func (p *Panel) GetContentArea() (x, y, width, height int) {
	if p.HasBorder {
		return p.X + 1, p.Y + 1, p.Width - 2, p.Height - 2
	}
	return p.X, p.Y, p.Width, p.Height
}

// DrawText draws text within the panel, handling word wrapping
func (p *Panel) DrawText(grid gruid.Grid, text string, style gruid.Style, startLine int) int {
	contentX, contentY, contentWidth, contentHeight := p.GetContentArea()

	if startLine >= contentHeight {
		return 0
	}

	lines := p.WrapText(text, contentWidth)
	linesDrawn := 0

	for i, line := range lines {
		if i < startLine {
			continue
		}

		lineY := contentY + i - startLine
		if lineY >= contentY+contentHeight {
			break
		}

		for j, r := range line {
			if j >= contentWidth {
				break
			}
			grid.Set(gruid.Point{X: contentX + j, Y: lineY}, gruid.Cell{Rune: r, Style: style})
		}
		linesDrawn++
	}

	return linesDrawn
}

// WrapText wraps text to fit within the specified width
func (p *Panel) WrapText(text string, width int) []string {
	if width <= 0 {
		return []string{}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// If adding this word would exceed the width, start a new line
		if currentLine.Len() > 0 && currentLine.Len()+1+utf8.RuneCountInString(word) > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		// Add word to current line
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	// Add the last line if it has content
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}

// DrawProgressBar draws a progress bar within the panel
func (p *Panel) DrawProgressBar(grid gruid.Grid, x, y, width int, current, max int, style gruid.Style) {
	if max <= 0 || width <= 0 {
		return
	}

	filled := (current * width) / max
	if filled > width {
		filled = width
	}

	for i := 0; i < width; i++ {
		var r rune
		if i < filled {
			r = '█'
		} else {
			r = '░'
		}

		if x+i < grid.Size().X && y < grid.Size().Y {
			grid.Set(gruid.Point{X: x + i, Y: y}, gruid.Cell{Rune: r, Style: style})
		}
	}
}

// IsPointInPanel checks if a point is within the panel bounds
func (p *Panel) IsPointInPanel(x, y int) bool {
	return x >= p.X && x < p.X+p.Width && y >= p.Y && y < p.Y+p.Height
}
