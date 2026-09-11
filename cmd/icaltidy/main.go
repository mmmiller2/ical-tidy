// Command icaltidy reads iCalendar content from stdin and writes the
// normalized form to stdout.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mmmiller2/ical-tidy"
)

func main() {
	width := flag.Int("fold-width", 75, "fold content lines at this many octets (RFC 5545 recommends 75)")
	flag.Parse()

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "icaltidy: reading stdin:", err)
		os.Exit(1)
	}

	if _, err := os.Stdout.WriteString(icaltidy.FormatWidth(string(input), *width)); err != nil {
		fmt.Fprintln(os.Stderr, "icaltidy: writing stdout:", err)
		os.Exit(1)
	}
}
