package comments

import "testing"

func TestApprovedCommentDelta(t *testing.T) {
	cases := []struct {
		name     string
		previous string
		next     string
		want     int
	}{
		{name: "pending to approved", previous: "pending", next: "approved", want: 1},
		{name: "approved to rejected", previous: "approved", next: "rejected", want: -1},
		{name: "approved to approved", previous: "approved", next: "approved", want: 0},
		{name: "pending to spam", previous: "pending", next: "spam", want: 0},
		{name: "trim and normalize", previous: " APPROVED ", next: "deleted", want: -1},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := approvedCommentDelta(tt.previous, tt.next); got != tt.want {
				t.Fatalf("approvedCommentDelta(%q, %q) = %d, want %d", tt.previous, tt.next, got, tt.want)
			}
		})
	}
}
