package match

import (
	"cmp"
	"fmt"

	"github.com/go-openapi/testify/v2/assert"
)

type Matcher interface {
	Matches(actual any) Result
}

type Result struct {
	Matches     bool
	Description string
}

type MatcherFunc func(actual any) Result

func (m MatcherFunc) Matches(actual any) Result {
	return m(actual)
}

func Same[T any](expected *T) Matcher {
	return MatcherFunc(func(actual any) Result {
		if actual == expected {
			return Result{Matches: true}
		}
		return Result{
			Matches: false,
			Description: fmt.Sprintf("Not same: \n"+
				"expected: %[2]s (%[1]T)(%[1]p)\n"+
				"actual  : %[4]s (%[3]T)(%[3]p)", expected, truncatingFormat("%#v", expected), actual, truncatingFormat("%#v", actual)),
		}
	})
}

func Equal(expected any) Matcher {
	return MatcherFunc(func(actual any) Result {
		if assert.ObjectsAreEqual(expected, actual) {
			return Result{Matches: true}
		}
		expectedStr, actualStr := formatUnequalValues(expected, actual)
		return Result{
			Matches: false,
			Description: fmt.Sprintf("Not equal: \n"+
				"expected: %s\n"+
				"actual  : %s",
				expectedStr,
				actualStr),
		}
	})
}

func Len(length int) Matcher {
	return MatcherFunc(func(actual any) Result {
		actualLen, ok := getLen(actual)
		if ok && actualLen != length {
			return Result{
				Matches:     false,
				Description: fmt.Sprintf("%q should have %d item(s), but has %d", truncatingFormat("%v", actual), length, actualLen),
			}
		}

		return Result{Matches: true}
	})
}

func Greater[T cmp.Ordered](value T) Matcher {
	return MatcherFunc(func(actual any) Result {
		if actual, ok := actual.(T); ok {
			if actual > value {
				return Result{Matches: true}
			}
		}

		valueStr, actualStr := formatUnequalValues(value, actual)
		return Result{
			Matches:     false,
			Description: fmt.Sprintf("%s is not greater than %s", valueStr, actualStr),
		}
	})
}

func GreaterOrEqual[T cmp.Ordered](value T) Matcher {
	return MatcherFunc(func(actual any) Result {
		if actual, ok := actual.(T); ok {
			if actual >= value {
				return Result{Matches: true}
			}
		}

		valueStr, actualStr := formatUnequalValues(value, actual)
		return Result{
			Matches:     false,
			Description: fmt.Sprintf("%s is not greater than or equal to %s", valueStr, actualStr),
		}
	})
}

func Less[T cmp.Ordered](value T) Matcher {
	return MatcherFunc(func(actual any) Result {
		if actual, ok := actual.(T); ok {
			if actual < value {
				return Result{Matches: true}
			}
		}

		valueStr, actualStr := formatUnequalValues(value, actual)
		return Result{
			Matches:     false,
			Description: fmt.Sprintf("%s is not less than %s", valueStr, actualStr),
		}
	})
}

func LessOrEqual[T cmp.Ordered](value T) Matcher {
	return MatcherFunc(func(actual any) Result {
		if actual, ok := actual.(T); ok {
			if actual <= value {
				return Result{Matches: true}
			}
		}

		valueStr, actualStr := formatUnequalValues(value, actual)
		return Result{
			Matches:     false,
			Description: fmt.Sprintf("%s is not less than or equal to %s", valueStr, actualStr),
		}
	})
}

func All(matchers ...Matcher) Matcher {
	return MatcherFunc(func(actual any) Result {
		for _, m := range matchers {
			r := m.Matches(actual)
			if !r.Matches {
				return r
			}
		}
		return Result{Matches: true}
	})
}
