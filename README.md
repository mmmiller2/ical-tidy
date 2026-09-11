# ical-tidy

Every calendar app that touches a `.ics` file leaves its own fingerprints on
it: Outlook exports with bare `\n` line endings, Google Calendar folds lines
at a slightly different width than the spec recommends, some tool leaves
trailing spaces on every line, property names show up as `Begin:VEvent`
instead of `BEGIN:VEVENT`. None of it breaks parsers that are lenient about
it, but it makes the files miserable to diff, grep, or hand-edit, and it
trips up parsers that follow RFC 5545 strictly.

`ical-tidy` takes that kind of file and normalizes it: consistent CRLF line
endings, content lines properly unfolded and refolded at 75 octets, no
trailing whitespace, no blank lines, and property/parameter names
uppercased. It does not validate that the calendar is well-formed - it is a
formatter, not a linter.

## Library usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/mmmiller2/ical-tidy"
)

func main() {
	raw, err := os.ReadFile("messy.ics")
	if err != nil {
		panic(err)
	}
	clean := icaltidy.Format(string(raw))
	fmt.Print(clean)
}
```

Every exported function takes plain strings in and returns plain strings
(or `[]string`) out - no file I/O, no globals - so they are straightforward
to unit test:

```go
icaltidy.NormalizePropertyName("summary:Team sync") // "SUMMARY:Team sync"
icaltidy.UnfoldLines("SUMMARY:one\r\n two")          // []string{"SUMMARY:onetwo"}
```

## Command line usage

```sh
go run ./cmd/icaltidy < messy.ics > clean.ics
```

Pass `-fold-width` to fold at a width other than the RFC 5545 recommended 75
octets, e.g. for a client that expects a narrower fold:

```sh
go run ./cmd/icaltidy -fold-width 60 < messy.ics > clean.ics
```

## Example

Input:

```
begin:vcalendar
version:2.0
begin:vevent
summary:Weekly sync
end:vevent
end:vcalendar
```

Output:

```
BEGIN:VCALENDAR\r\n
VERSION:2.0\r\n
BEGIN:VEVENT\r\n
SUMMARY:Weekly sync\r\n
END:VEVENT\r\n
END:VCALENDAR\r\n
```

(shown with explicit `\r\n` here since most terminals render carriage
returns invisibly)

## Status

Early skeleton. Line folding/unfolding, whitespace cleanup, and property
and parameter name casing work and are tested, including parameter values
that are quoted and contain a `:` or `;` of their own, and preserving the
original case of `X-` prefixed experimental property names. The CLI fold
width is configurable; the library defaults to the RFC 5545 recommended 75
octets everywhere else.

## License

MIT, see [LICENSE](LICENSE).
