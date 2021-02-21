package main

import (
	"regexp"
	"testing"

	"golang.org/x/text/transform"
)

func TestMatcherTransform(t *testing.T) {
	tests := map[string]string{
		"Günter Çedille":     "Gunter Cedille",
		"Jan Mączyński":      "Jan Maczynski",
		"University of Łódź": "University of Lodz",
		"Ø":                  "O",
		"Ó":                  "O",
		"ä ñ ö ü ÿ":          "a n o u y",
		"Henry Ⅷ":            "Henry VIII",
	}
	for v, exp := range tests {
		re := regexp.MustCompile(v)
		m := newMatcher(re)
		transformed, _, _ := transform.String(*m.tr, v)
		assertEq(t, transformed, exp)
	}
}
