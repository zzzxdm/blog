package submissions

import "testing"

func TestCanReviewSubmission(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{status: StatusSubmitted, want: true},
		{status: StatusReturned, want: true},
		{status: StatusDraft, want: false},
		{status: StatusRejected, want: false},
		{status: StatusPublished, want: false},
	}

	for _, tt := range cases {
		t.Run(tt.status, func(t *testing.T) {
			if got := canReviewSubmission(tt.status); got != tt.want {
				t.Fatalf("canReviewSubmission(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
