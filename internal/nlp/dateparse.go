package nlp

import (
	"time"

	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/common"
	"github.com/olebedev/when/rules/en"
)

// DateResult holds the result of parsing a natural language date from text.
type DateResult struct {
	Time  time.Time
	Text  string // The matched text fragment (e.g. "on Friday")
	Index int    // Start position of the match in the original string
}

var parser *when.Parser

func init() {
	parser = when.New(nil)
	parser.Add(en.All...)
	parser.Add(common.All...)
}

// ExtractDate parses text for natural language date/time references.
// Returns nil if no date is found.
func ExtractDate(text string, ref time.Time) *DateResult {
	r, err := parser.Parse(text, ref)
	if err != nil || r == nil {
		return nil
	}
	return &DateResult{
		Time:  r.Time,
		Text:  r.Text,
		Index: r.Index,
	}
}

// StripDate removes the matched date text from the original string,
// cleaning up extra whitespace. Useful for extracting a clean title.
func StripDate(text string, result *DateResult) string {
	if result == nil {
		return text
	}
	before := text[:result.Index]
	after := text[result.Index+len(result.Text):]

	// Clean up: trim spaces and common prepositions left dangling.
	cleaned := trimJoinParts(before, after)
	return cleaned
}

func trimJoinParts(before, after string) string {
	// Trim trailing/leading spaces.
	b := []byte(before)
	a := []byte(after)

	// Trim trailing space from before.
	for len(b) > 0 && b[len(b)-1] == ' ' {
		b = b[:len(b)-1]
	}
	// Trim leading space from after.
	for len(a) > 0 && a[0] == ' ' {
		a = a[1:]
	}

	if len(b) == 0 {
		return string(a)
	}
	if len(a) == 0 {
		return string(b)
	}
	return string(b) + " " + string(a)
}
