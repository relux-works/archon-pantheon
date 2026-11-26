// Package releasegrammar validates and compares release strings without process control.
package releasegrammar

import (
	"strconv"
	"strings"
)

// IsReleaseTriple reports whether s is a dotted release triple:
// three dot-separated runs of ASCII digits, "2.1.261". There is no
// leading-zero rule and no prerelease suffix: the pinned releases are
// exact triples, and anything else fails closed at the probe (never
// established) rather than matching a row it merely resembles. A future
// verified release with a suffix re-verifies this rule with its grammar
// version.
func IsReleaseTriple(s string) bool {
	_, _, _, ok := ParseReleaseTriple(s)
	return ok
}

// ParseReleaseTriple parses a dotted release triple ("1.5.18") into numeric
// major, minor, and patch components.
func ParseReleaseTriple(s string) (major, minor, patch int, ok bool) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	var nums [3]int
	for idx, part := range parts {
		if part == "" {
			return 0, 0, 0, false
		}
		for i := 0; i < len(part); i++ {
			if part[i] < '0' || part[i] > '9' {
				return 0, 0, 0, false
			}
		}
		v, err := strconv.Atoi(part)
		if err != nil {
			return 0, 0, 0, false
		}
		nums[idx] = v
	}
	return nums[0], nums[1], nums[2], true
}

// CompareReleaseTriple compares two dotted release triples numerically.
// Returns -1 if a < b, 0 if a == b, +1 if a > b, and ok=false if either input
// is not a valid release triple.
func CompareReleaseTriple(a, b string) (int, bool) {
	aMaj, aMin, aPat, okA := ParseReleaseTriple(a)
	bMaj, bMin, bPat, okB := ParseReleaseTriple(b)
	if !okA || !okB {
		return 0, false
	}
	switch {
	case aMaj != bMaj:
		if aMaj < bMaj {
			return -1, true
		}
		return 1, true
	case aMin != bMin:
		if aMin < bMin {
			return -1, true
		}
		return 1, true
	case aPat != bPat:
		if aPat < bPat {
			return -1, true
		}
		return 1, true
	default:
		return 0, true
	}
}
