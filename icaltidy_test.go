package icaltidy

import "testing"

func TestUnfoldLines(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "no folding",
			input: "BEGIN:VCALENDAR\r\nEND:VCALENDAR",
			want:  []string{"BEGIN:VCALENDAR", "END:VCALENDAR"},
		},
		{
			name:  "crlf continuation",
			input: "SUMMARY:Long meeting\r\n title continues here\r\nEND:VEVENT",
			want:  []string{"SUMMARY:Long meeting title continues here", "END:VEVENT"},
		},
		{
			name:  "bare lf continuation with tab",
			input: "DESCRIPTION:part one\n\tpart two",
			want:  []string{"DESCRIPTION:part onepart two"},
		},
		{
			name:  "blank lines dropped",
			input: "BEGIN:VEVENT\r\n\r\nEND:VEVENT\r\n",
			want:  []string{"BEGIN:VEVENT", "END:VEVENT"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UnfoldLines(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d lines, want %d: %v", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestFoldLine(t *testing.T) {
	short := "SUMMARY:short"
	if got := FoldLine(short, 75); got != short {
		t.Errorf("short line should be unchanged, got %q", got)
	}

	long := "SUMMARY:" + repeat("x", 100)
	got := FoldLine(long, 75)

	for _, part := range splitCRLF(got) {
		if len(part) > 75 {
			t.Errorf("folded segment exceeds 75 octets: %q (%d)", part, len(part))
		}
	}

	unfolded := UnfoldLines(got)
	if len(unfolded) != 1 || unfolded[0] != long {
		t.Errorf("fold/unfold round trip mismatch: got %v", unfolded)
	}
}

func TestNormalizePropertyName(t *testing.T) {
	cases := map[string]string{
		"summary:Team sync":            "SUMMARY:Team sync",
		"Dtstart;TZID=America/NY:2026": "DTSTART;TZID=America/NY:2026",
		"BEGIN:VEVENT":                 "BEGIN:VEVENT",
		":no name":                     ":no name",
		"no-delimiter":                 "no-delimiter",
		"dtstart;tzid=America/NY:2026": "DTSTART;TZID=America/NY:2026",
		`attendee;cn="Doe, John";delegated-to="mailto:jane@example.com":mailto:john@example.com`: `ATTENDEE;CN="Doe, John";DELEGATED-TO="mailto:jane@example.com":mailto:john@example.com`,
		`x-prop;param="a;b":value`:    `x-prop;PARAM="a;b":value`,
		"X-WR-CALNAME:My Calendar":    "X-WR-CALNAME:My Calendar",
		"X-Wr-CalName:My Calendar":    "X-Wr-CalName:My Calendar",
		"x-apple-tzid;VALUE=text:foo": "x-apple-tzid;VALUE=text:foo",
	}
	for input, want := range cases {
		if got := NormalizePropertyName(input); got != want {
			t.Errorf("NormalizePropertyName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFormat(t *testing.T) {
	messy := "begin:vcalendar\n" +
		"version:2.0  \n" +
		"\n" +
		"begin:vevent\n" +
		"summary:Weekly planning meeting for the whole engineering organization\n" +
		"end:vevent\n" +
		"end:vcalendar"

	got := Format(messy)

	for _, part := range splitCRLF(got) {
		if part == "" {
			continue
		}
		if len(part) > 75 {
			t.Errorf("output line exceeds 75 octets: %q", part)
		}
	}

	unfolded := UnfoldLines(got)
	want := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"BEGIN:VEVENT",
		"SUMMARY:Weekly planning meeting for the whole engineering organization",
		"END:VEVENT",
		"END:VCALENDAR",
	}
	if len(unfolded) != len(want) {
		t.Fatalf("got %d lines, want %d: %v", len(unfolded), len(want), unfolded)
	}
	for i := range want {
		if unfolded[i] != want[i] {
			t.Errorf("line %d: got %q, want %q", i, unfolded[i], want[i])
		}
	}
}

func repeat(s string, n int) string {
	out := make([]byte, 0, n*len(s))
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}

func splitCRLF(s string) []string {
	var parts []string
	start := 0
	for i := 0; i+1 < len(s); i++ {
		if s[i] == '\r' && s[i+1] == '\n' {
			parts = append(parts, s[start:i])
			start = i + 2
			i++
		}
	}
	parts = append(parts, s[start:])
	return parts
}
