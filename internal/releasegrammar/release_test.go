package releasegrammar

import "testing"

func TestCompareReleaseTriple(t *testing.T) {
	cases := []struct {
		a, b string
		want int
		ok   bool
	}{
		{"1.5.16", "1.5.16", 0, true},
		{"1.5.18", "1.5.16", 1, true},
		{"1.5.15", "1.5.16", -1, true},
		{"1.10.0", "1.5.18", 1, true},
		{"2.0.0", "1.99.99", 1, true},
		{"1.5", "1.5.0", 0, false},
		{"1.5.18-beta", "1.5.18", 0, false},
	}
	for _, tc := range cases {
		got, ok := CompareReleaseTriple(tc.a, tc.b)
		if ok != tc.ok || got != tc.want {
			t.Errorf("CompareReleaseTriple(%q, %q) = (%d, %v), want (%d, %v)", tc.a, tc.b, got, ok, tc.want, tc.ok)
		}
	}
}
