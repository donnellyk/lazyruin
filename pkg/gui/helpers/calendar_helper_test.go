package helpers

import "testing"

func TestDaysIn(t *testing.T) {
	tests := []struct {
		name  string
		year  int
		month int
		want  int
	}{
		{name: "January", year: 2026, month: 1, want: 31},
		{name: "February non-leap", year: 2025, month: 2, want: 28},
		{name: "February leap", year: 2024, month: 2, want: 29},
		{name: "April", year: 2026, month: 4, want: 30},
		{name: "December", year: 2026, month: 12, want: 31},
		{name: "September", year: 2026, month: 9, want: 30},
		{name: "February century non-leap", year: 1900, month: 2, want: 28},
		{name: "February century leap", year: 2000, month: 2, want: 29},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := daysIn(tc.year, tc.month)
			if got != tc.want {
				t.Errorf("daysIn(%d, %d) = %d, want %d", tc.year, tc.month, got, tc.want)
			}
		})
	}
}
