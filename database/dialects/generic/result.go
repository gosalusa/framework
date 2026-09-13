package generic

import (
	"fmt"
	"strings"

	"gosalusa.com/database/dialects"
)

type rawQueryBuilder struct {
	query    strings.Builder
	bindings []any
	err      error
}

func newRawQueryBuilder() *rawQueryBuilder {
	return &rawQueryBuilder{
		query:    strings.Builder{},
		bindings: []any{},
	}
}

func (b *rawQueryBuilder) Add(r dialects.RawQuery, err error) *rawQueryBuilder {
	if b.err != nil || r.SQL == "" {
		return b
	}
	if err != nil {
		b.err = err
		return b
	} else {
		if b.query.Len() > 0 {
			b.query.WriteString(" ")
		}
		b.query.WriteString(r.SQL)
		b.bindings = append(b.bindings, r.Bindings...)
	}
	return b
}

func (b *rawQueryBuilder) AddStringf(s string, a ...any) *rawQueryBuilder {
	return b.AddString(fmt.Sprintf(s, a...))
}
func (b *rawQueryBuilder) AddString(s string) *rawQueryBuilder {
	return b.Add(dialects.RawQuery{
		SQL: s,
	}, nil)
}

func (b *rawQueryBuilder) AddStringNoSpace(s string) *rawQueryBuilder {
	b.query.WriteString(s)
	return b
}

func (b *rawQueryBuilder) Build() (dialects.RawQuery, error) {
	if b.err != nil {
		return dialects.RawQuery{}, b.err
	}
	return dialects.RawQuery{
		SQL:      b.query.String(),
		Bindings: b.bindings,
	}, nil
}

func joinRawQueries(results []dialects.RawQuery, sep string) dialects.RawQuery {
	if len(results) == 0 {
		return dialects.RawQuery{
			SQL:      "",
			Bindings: []any{},
		}
	}
	if len(results) == 1 {
		return results[0]
	}

	queryLen := len(sep) * (len(results) - 1)
	bindingCount := 0

	for _, r := range results {
		queryLen += len(r.SQL)
		bindingCount += len(r.Bindings)
	}

	query := strings.Builder{}
	query.Grow(queryLen)
	bindings := make([]any, 0, bindingCount)

	query.WriteString(results[0].SQL)
	bindings = append(bindings, results[0].Bindings...)

	for _, r := range results[1:] {
		query.WriteString(sep)
		query.WriteString(r.SQL)
		bindings = append(bindings, r.Bindings...)
	}

	return dialects.RawQuery{
		SQL:      query.String(),
		Bindings: bindings,
	}
}

func mapJoinRawQueries[T any](arr []T, sep string, fn func(v T) (dialects.RawQuery, error)) (dialects.RawQuery, error) {
	results := make([]dialects.RawQuery, len(arr))
	var err error
	for i, v := range arr {
		results[i], err = fn(v)
		if err != nil {
			return dialects.RawQuery{}, err
		}
	}
	return joinRawQueries(results, sep), nil
}
