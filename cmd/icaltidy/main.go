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
	checkUIDs := flag.Bool("check-uids", false, "report UIDs shared by more than one VEVENT on stderr and exit with status 2")
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

	// Checked after the output is written so a duplicate does not stop the
	// file from being tidied; only the exit status changes.
	if *checkUIDs {
		dups := icaltidy.FindDuplicateUIDs(string(input))
		for _, d := range dups {
			fmt.Fprintf(os.Stderr, "icaltidy: duplicate UID %q on %d events\n", d.UID, d.Count)
		}
		if len(dups) > 0 {
			os.Exit(2)
		}
	}
}
