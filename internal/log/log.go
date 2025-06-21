package log

import (
	"fmt"
	"time"

	"codeberg.org/anaseto/gruid"
)

type logStyle int

const (
	logNormal logStyle = iota
	logCritic
	logNotable
	logDamage
	logSpecial
	logStatusEnd
	logError
	logConfirm
)

type logEntry struct {
	Text  string
	MText string
	Index int
	Tick  bool
	Style logStyle
	Dups  int
}

// Message represents a single message with associated color and timestamp.
type Message struct {
	Text      string
	Color     gruid.Color
	Timestamp time.Time
}

// MessageLog stores a list of game messages.
type MessageLog struct {
	Messages []Message
	// TODO: Consider adding a max size and pruning logic if needed.
}

// NewMessageLog creates a new empty MessageLog.
func NewMessageLog() *MessageLog {
	return &MessageLog{
		Messages: []Message{},
	}
}

// AddMessage adds a new message with the given text and color to the log.
func (ml *MessageLog) AddMessage(text string, color gruid.Color) {
	ml.Messages = append(ml.Messages, Message{
		Text:      text,
		Color:     color,
		Timestamp: time.Now(),
	})
	// TODO: Pruning logic if max size is implemented.
}

// AddMessagef adds a new formatted message with the given color to the log.
func (ml *MessageLog) AddMessagef(color gruid.Color, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	ml.AddMessage(text, color)
}

// AddMessageWithTimestamp adds a message with a specific timestamp (for loading saved messages)
func (ml *MessageLog) AddMessageWithTimestamp(text string, color gruid.Color, timestamp time.Time) {
	ml.Messages = append(ml.Messages, Message{
		Text:      text,
		Color:     color,
		Timestamp: timestamp,
	})
}

// // Optional: Method to retrieve messages (e.g., for UI rendering)
// func (ml *MessageLog) GetMessages(count int) []Message {
//  start := len(ml.Messages) - count
//  if start < 0 {
//      start = 0
//  }
//  return ml.Messages[start:]
// }
