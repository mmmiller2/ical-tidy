package icaltidy

import "strings"

// DuplicateUID describes a UID that is shared by more than one VEVENT.
type DuplicateUID struct {
	UID   string
	Count int
}

// FindDuplicateUIDs returns the UIDs that appear on more than one VEVENT, in
// the order each UID was first seen. Events are matched on UID together with
// RECURRENCE-ID, because an override of a single occurrence of a recurring
// event legitimately reuses the UID of its master event. Only UID lines
// directly inside a VEVENT count; a UID in some other component is ignored,
// and so is an event with no UID at all.
func FindDuplicateUIDs(input string) []DuplicateUID {
	type key struct{ uid, recurrenceID string }

	counts := map[key]int{}
	var order []key

	inEvent := false
	var uid, recurrenceID string
	haveUID := false

	for _, line := range UnfoldLines(input) {
		name, value := splitContentLine(line)
		switch {
		case name == "BEGIN" && strings.EqualFold(value, "VEVENT"):
			inEvent = true
			uid, recurrenceID, haveUID = "", "", false
		case name == "END" && strings.EqualFold(value, "VEVENT"):
			if inEvent && haveUID {
				k := key{uid, recurrenceID}
				if counts[k] == 0 {
					order = append(order, k)
				}
				counts[k]++
			}
			inEvent = false
		case inEvent && name == "UID" && !haveUID:
			uid, haveUID = strings.TrimSpace(value), true
		case inEvent && name == "RECURRENCE-ID":
			recurrenceID = strings.TrimSpace(value)
		}
	}

	var dups []DuplicateUID
	for _, k := range order {
		if counts[k] > 1 {
			dups = append(dups, DuplicateUID{UID: k.uid, Count: counts[k]})
		}
	}
	return dups
}

// splitContentLine returns the uppercased property name and the raw value of
// a content line, skipping over any parameters. The value is empty if the
// line has no top-level colon.
func splitContentLine(line string) (name, value string) {
	colon := firstTopLevelByte(line, ":")
	if colon < 0 {
		return "", ""
	}
	name = line[:colon]
	if semi := firstTopLevelByte(name, ";"); semi >= 0 {
		name = name[:semi]
	}
	return strings.ToUpper(name), line[colon+1:]
}
