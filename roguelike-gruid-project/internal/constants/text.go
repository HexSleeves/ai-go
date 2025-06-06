package constants

// CommonTextChars is a pre-computed set of characters commonly used in UI text
// Using a map for O(1) lookup performance instead of recreating it each time
var CommonTextChars = map[rune]bool{
	// Lowercase letters
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true, 'g': true, 'h': true,
	'i': true, 'j': true, 'k': true, 'l': true, 'm': true, 'n': true, 'o': true, 'p': true,
	'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true, 'w': true, 'x': true,
	'y': true, 'z': true,

	// Uppercase letters
	'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true, 'G': true, 'H': true,
	'I': true, 'J': true, 'K': true, 'L': true, 'M': true, 'N': true, 'O': true, 'P': true,
	'Q': true, 'R': true, 'S': true, 'T': true, 'U': true, 'V': true, 'W': true, 'X': true,
	'Y': true, 'Z': true,

	// Digits
	'0': true, '1': true, '2': true, '3': true, '4': true, '5': true, '6': true, '7': true,
	'8': true, '9': true,

	// Common punctuation and symbols used in UI text
	' ': true, '.': true, ',': true, ':': true, ';': true, '!': true, '?': true, '-': true,
	'_': true, '(': true, ')': true, '[': true, ']': true, '{': true, '}': true, '/': true,
	'\\': true, '|': true, '=': true, '+': true, '*': true, '&': true, '%': true, '$': true,
	'"': true, '\'': true, '`': true, '~': true, '^': true, '<': true, '>': true,
}

// isCommonTextChar checks if a rune is commonly used in UI text
// Separated into its own function for better readability and potential reuse
func IsCommonTextChar(r rune) bool {
	return CommonTextChars[r]
}
