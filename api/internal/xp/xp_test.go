package xp

import "testing"

func TestAward(t *testing.T) {
	tests := []struct {
		bucket Bucket
		photo  bool
		want   int
	}{
		{Under30m, false, 30},
		{M30to60, false, 40},
		{OneTo2h, true, 90},
		{HalfDay, false, 100},
		{FullDay, true, 170},
	}
	for _, tt := range tests {
		if got := Award(tt.bucket, tt.photo); got != tt.want {
			t.Errorf("Award(%s, %v) = %d, want %d", tt.bucket, tt.photo, got, tt.want)
		}
	}
}

func TestLevelForXp(t *testing.T) {
	tests := []struct {
		total int
		want  Level
	}{
		{0, Level{0, 0, 100}},
		{99, Level{0, 99, 100}},
		{100, Level{1, 0, 200}},
		{299, Level{1, 199, 200}},
		{300, Level{2, 0, 300}},
		{650, Level{3, 50, 400}},
	}
	for _, tt := range tests {
		if got := LevelForXp(tt.total); got != tt.want {
			t.Errorf("LevelForXp(%d) = %+v, want %+v", tt.total, got, tt.want)
		}
	}
}

func TestTitleClampsToLast(t *testing.T) {
	if got := Title(100); got != "Legend" {
		t.Fatalf("Title(100) = %q, want Legend", got)
	}
}
