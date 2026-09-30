package icaltidy

import (
	"reflect"
	"testing"
)

func TestFindDuplicateUIDs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []DuplicateUID
	}{
		{
			name:  "no events",
			input: "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n",
			want:  nil,
		},
		{
			name: "unique uids",
			input: "BEGIN:VEVENT\r\nUID:a\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nUID:b\r\nEND:VEVENT\r\n",
			want: nil,
		},
		{
			name: "same uid twice",
			input: "BEGIN:VEVENT\r\nUID:a\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nUID:a\r\nEND:VEVENT\r\n",
			want: []DuplicateUID{{UID: "a", Count: 2}},
		},
		{
			name: "lowercase names and bare newlines",
			input: "begin:vevent\nuid:a\nend:vevent\n" +
				"begin:vevent\nuid:a\nend:vevent\n" +
				"begin:vevent\nuid:a\nend:vevent\n",
			want: []DuplicateUID{{UID: "a", Count: 3}},
		},
		{
			name: "folded uid value",
			input: "BEGIN:VEVENT\r\nUID:abc\r\n def\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nUID:abcdef\r\nEND:VEVENT\r\n",
			want: []DuplicateUID{{UID: "abcdef", Count: 2}},
		},
		{
			name: "recurrence override is not a duplicate",
			input: "BEGIN:VEVENT\r\nUID:a\r\nRRULE:FREQ=DAILY\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nUID:a\r\nRECURRENCE-ID:20260101T090000Z\r\nEND:VEVENT\r\n",
			want: nil,
		},
		{
			name: "duplicate overrides of the same occurrence",
			input: "BEGIN:VEVENT\r\nUID:a\r\nRECURRENCE-ID;TZID=UTC:20260101T090000\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nUID:a\r\nRECURRENCE-ID;TZID=UTC:20260101T090000\r\nEND:VEVENT\r\n",
			want: []DuplicateUID{{UID: "a", Count: 2}},
		},
		{
			name: "uid outside a vevent is ignored",
			input: "BEGIN:VTODO\r\nUID:a\r\nEND:VTODO\r\n" +
				"BEGIN:VEVENT\r\nUID:a\r\nEND:VEVENT\r\n",
			want: nil,
		},
		{
			name: "events without a uid are ignored",
			input: "BEGIN:VEVENT\r\nSUMMARY:x\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nSUMMARY:y\r\nEND:VEVENT\r\n",
			want: nil,
		},
		{
			name: "reported in first-seen order",
			input: "BEGIN:VEVENT\r\nUID:b\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nUID:a\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nUID:a\r\nEND:VEVENT\r\n" +
				"BEGIN:VEVENT\r\nUID:b\r\nEND:VEVENT\r\n",
			want: []DuplicateUID{{UID: "b", Count: 2}, {UID: "a", Count: 2}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindDuplicateUIDs(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindDuplicateUIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}
