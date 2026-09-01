// Command icaltidy reads iCalendar content from stdin and writes the
// normalized form to stdout.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mmmiller2/ical-tidy"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "icaltidy: reading stdin:", err)
		os.Exit(1)
	}

	if _, err := os.Stdout.WriteString(icaltidy.Format(string(input))); err != nil {
		fmt.Fprintln(os.Stderr, "icaltidy: writing stdout:", err)
		os.Exit(1)
	}
}
