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

// NormalizePropertyName uppercases the property name and every parameter
// name on a content line, leaving parameter values and the property value
// untouched. Property and parameter names are case-insensitive per RFC
// 5545, but calendar files in the wild mix cases inconsistently, which
// makes them annoying to diff and grep.
//
// A quoted parameter value (the only kind allowed to contain ':', ';', or
// ',') is tracked so its contents are never mistaken for a delimiter, e.g.
// the embedded ':' in `DELEGATED-TO="mailto:jane@example.com"` does not end
// the parameter list early. If the line has no ':' or ';', or starts with
// one, it is returned as-is.
func NormalizePropertyName(line string) string {
	nameEnd := firstTopLevelByte(line, ":;")
	if nameEnd <= 0 {
		return line
	}

	var out strings.Builder
	out.WriteString(strings.ToUpper(line[:nameEnd]))
	rest := line[nameEnd:]

	for len(rest) > 0 && rest[0] == ';' {
		out.WriteByte(';')
		rest = rest[1:]

		paramEnd := firstTopLevelByte(rest, "=:;")
		if paramEnd < 0 {
			out.WriteString(rest)
			return out.String()
		}
		out.WriteString(strings.ToUpper(rest[:paramEnd]))
		if rest[paramEnd] != '=' {
			rest = rest[paramEnd:]
			continue
		}
		out.WriteByte('=')
		rest = rest[paramEnd+1:]

		valueEnd := firstTopLevelByte(rest, ";:")
		if valueEnd < 0 {
			out.WriteString(rest)
			return out.String()
		}
		out.WriteString(rest[:valueEnd])
		rest = rest[valueEnd:]
	}
	out.WriteString(rest)
	return out.String()
}

// firstTopLevelByte returns the index of the first byte in s that also
// appears in chars and falls outside a double-quoted span, or -1 if there
// is none. Quotes cannot nest or escape in RFC 5545 parameter values, so a
// plain toggle on '"' is enough to track them.
func firstTopLevelByte(s string, chars string) int {
	inQuotes := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inQuotes = !inQuotes
			continue
		}
		if !inQuotes && strings.IndexByte(chars, c) >= 0 {
			return i
		}
	}
	return -1
}
