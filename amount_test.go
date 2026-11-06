package util9s

import "testing"

func TestYuanToFen(t *testing.T) {
	cases := []struct {
		yuan float64
		fen  int64
	}{
		{0, 0},
		{1, 100},
		{58.4, 5840},
		{0.005, 1},  // Round half away → 0.01? 0.005*100=0.5 → Round → 1
		{0.004, 0},
		{-1.5, -150},
		{19.99, 1999},
	}
	for _, tc := range cases {
		if got := YuanToFen(tc.yuan); got != tc.fen {
			t.Fatalf("YuanToFen(%v)=%d want %d", tc.yuan, got, tc.fen)
		}
	}
}

func TestMustMatchPayAmount(t *testing.T) {
	if !MustMatchPayAmount(100, 100, 0) {
		t.Fatal("exact")
	}
	if MustMatchPayAmount(100, 102, 0) {
		t.Fatal("no tol")
	}
	if !MustMatchPayAmount(100, 101, 1) {
		t.Fatal("tol 1")
	}
	if !MatchPayAmountDefault(100, 101) {
		t.Fatal("default tol")
	}
	if MustMatchPayAmount(100, 103, 1) {
		t.Fatal("beyond tol")
	}
	m := Money(5840)
	if m.Fen() != 5840 || m.YuanString() != "58.40" {
		t.Fatalf("Money=%v %q", m.Fen(), m.YuanString())
	}
}
