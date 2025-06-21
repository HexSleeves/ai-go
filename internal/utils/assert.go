package utils

import "fmt"

func Assert(condition bool, message string) {
	if !condition {
		panic(message)
	}
}

// Assertf is a formatted version of Assert that allows for formatted messages.
func Assertf(condition bool, format string, args ...any) {
	if !condition {
		panic(fmt.Sprintf(format, args...))
	}
}
