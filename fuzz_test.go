package icaltidy

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzFoldUnfold checks that any content line, once folded at an arbitrary
// width, comes back unchanged through UnfoldLines. Invalid UTF-8 is skipped:
// FoldLine's octet-counting loop assumes valid runes and its behavior on
// malformed input isn't a claim this library makes. Lines containing \r or
// \n are skipped too, since those are the raw line delimiters UnfoldLines
// splits on, not characters a single logical line can contain.
func FuzzFoldUnfold(f *testing.F) {
	f.Add("SUMMARY:short line", 75)
	f.Add("SUMMARY:"+repeat("x", 200), 75)
	f.Add("DESCRIPTION:"+repeat("é", 40), 12)
	f.Add("SUMMARY:one two three", 1)
	f.Add(" leading space then a lot of filler text to force a fold", 20)

	f.Fuzz(func(t *testing.T, line string, width int) {
		if line == "" || strings.ContainsAny(line, "\r\n") || !utf8.ValidString(line) {
			t.Skip()
		}

		folded := FoldLine(line, width)
		got := UnfoldLines(folded)
		if len(got) != 1 || got[0] != line {
			t.Fatalf("round trip broke for width %d: got %v, want [%q]", width, got, line)
		}
	})
}
