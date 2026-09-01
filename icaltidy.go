// Package icaltidy normalizes iCalendar (RFC 5545) content lines that have
// been mangled by round trips through different calendar clients: mixed
// line endings, lines folded at the wrong width (or not folded at all),
// trailing whitespace, and lowercase property names.
package icaltidy

import (
	"strings"
	"unicode/utf8"
)

// maxOctets is the RFC 5545 recommended maximum line length, including the
// single leading whitespace octet on continuation lines.
const maxOctets = 75

// Format takes raw iCalendar content and returns a version with normalized
// line endings (CRLF), no blank lines, uppercase property names, and lines
// folded to the RFC 5545 recommended width. It does not validate that the
// input is a well-formed calendar; it only tidies the text.
func Format(input string) string {
	var out strings.Builder
	for _, line := range UnfoldLines(input) {
		line = strings.TrimRight(line, " \t")
		if line == "" {
			continue
		}
		line = NormalizePropertyName(line)
		out.WriteString(FoldLine(line, maxOctets))
		out.WriteString("\r\n")
	}
	return out.String()
}

// UnfoldLines splits raw calendar content into logical content lines,
// rejoining any continuation lines. Per RFC 5545, a line break immediately
// followed by a space or tab is a fold, not a real line break, and the
// leading whitespace is removed when rejoining. Line endings may be CRLF,
// bare LF, or bare CR; all are treated the same way. Blank raw lines are
// dropped rather than treated as empty logical lines.
func UnfoldLines(input string) []string {
	normalized := strings.ReplaceAll(input, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	var logical []string
	for _, raw := range strings.Split(normalized, "\n") {
		if raw == "" {
			continue
		}
		if len(logical) > 0 && (raw[0] == ' ' || raw[0] == '\t') {
			logical[len(logical)-1] += raw[1:]
			continue
		}
		logical = append(logical, raw)
	}
	return logical
}

// FoldLine folds a single logical line into RFC 5545 form: whenever adding
// the next rune would push the current output line past limit octets, a
// CRLF followed by a single space is inserted before it. limit is clamped
// up to 75 if given as 1 or less, since a fold needs room for the leading
// continuation space plus at least one content octet. Folding is done on
// UTF-8 octet boundaries so multi-byte runes are never split.
func FoldLine(line string, limit int) string {
	if limit <= 1 {
		limit = maxOctets
	}
	if len(line) <= limit {
		return line
	}

	var out strings.Builder
	count := 0
	for _, r := range line {
		rl := utf8.RuneLen(r)
		if count > 0 && count+rl > limit {
			out.WriteString("\r\n ")
			count = 1
		}
		out.WriteRune(r)
		count += rl
	}
	return out.String()
}

// NormalizePropertyName uppercases the property (and parameter block) name
// at the start of a content line, i.e. everything before the first ':' or
// ';', and leaves the rest of the line untouched. Property and parameter
// names are case-insensitive per RFC 5545, but calendar files in the wild
// mix cases inconsistently, which makes them annoying to diff and grep.
// If the line has no ':' or ';', or starts with one, it is returned as-is.
func NormalizePropertyName(line string) string {
	idx := strings.IndexAny(line, ":;")
	if idx <= 0 {
		return line
	}
	return strings.ToUpper(line[:idx]) + line[idx:]
}
