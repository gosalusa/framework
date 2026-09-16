package match_test

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
	"gosalusa.com/testing/match"
)

type TestCase struct {
	Name    string
	Matcher match.Matcher
	Value   any
	match.Result
}

var ptrA = new(1)
var ptrB = new(1)

const dontCheck = "🙈"

func TestSame(t *testing.T) {
	tcs := []TestCase{
		{
			Name:    "Same pass",
			Matcher: match.Same(ptrA),
			Value:   ptrA,
			Result:  match.Result{Matches: true},
		},
		{
			Name:    "Same fail",
			Matcher: match.Same(ptrA),
			Value:   ptrB,
			Result:  match.Result{Matches: false, Description: dontCheck},
		},
		{
			Name:    "Equal pass",
			Matcher: match.Equal(1),
			Value:   1,
			Result:  match.Result{Matches: true},
		},
		{
			Name:    "Equal fail",
			Matcher: match.Equal(1),
			Value:   2,
			Result:  match.Result{Matches: false, Description: "Not equal: \nexpected: 1\nactual  : 2"},
		},
		{
			Name:    "Len str pass",
			Matcher: match.Len(1),
			Value:   "a",
			Result:  match.Result{Matches: true},
		},
		{
			Name:    "Len str fail",
			Matcher: match.Len(1),
			Value:   "ab",
			Result:  match.Result{Matches: false, Description: "\"ab\" should have 1 item(s), but has 2"},
		},
		{
			Name:    "Len slice pass",
			Matcher: match.Len(1),
			Value:   []int{1},
			Result:  match.Result{Matches: true},
		},
		{
			Name:    "Len slice fail",
			Matcher: match.Len(1),
			Value:   []int{1, 2},
			Result:  match.Result{Matches: false, Description: "\"[1 2]\" should have 1 item(s), but has 2"},
		},
		{
			Name:    "All pass",
			Matcher: match.All(match.Len(1), match.Equal("a")),
			Value:   "a",
			Result:  match.Result{Matches: true},
		},
		{
			Name:    "All fail",
			Matcher: match.All(match.Len(1), match.Equal("b")),
			Value:   "a",
			Result:  match.Result{Matches: false, Description: "Not equal: \nexpected: \"b\"\nactual  : \"a\""},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			r := tc.Matcher.Matches(tc.Value)
			assert.Equal(t, tc.Result.Matches, r.Matches)
			if tc.Result.Description != dontCheck {
				assert.Equal(t, tc.Result.Description, r.Description)
			}
		})
	}
}
