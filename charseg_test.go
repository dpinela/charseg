package charseg

import (
	"reflect"
	"testing"
)

type testCase struct {
	in string
	out []string
}

var specialTestCases = []testCase{
	{in: "\u00E1", out: []string{"\u00E1"}},
	{in: "a\u0301", out: []string{"a\u0301"}},
	{in: "€3", out: []string{"€", "3"}},
}

func segment(s string) []string {
	var segments []string
	for len(s) > 0 {
		i := NextBoundary(s, true)
		segments = append(segments, s[:i])
		s = s[i:]
	}
	return segments
}

func TestSegmentation(t *testing.T) {
	testWithCases(t, specialTestCases)
	testWithCases(t, unicodeTestCases)
}

func testWithCases(t *testing.T, cases []testCase) {
	for _, tt := range cases {
		if got := segment(tt.in); !reflect.DeepEqual(got, tt.out) {
			t.Errorf("segment(%+q) = %+q; want %+q", tt.in, got, tt.out)
		}
	}
}