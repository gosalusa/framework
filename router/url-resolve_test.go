package router_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/internal/test"
	"gosalusa.com/router"
)

func TestToAttrs(t *testing.T) {
	testCases := []struct {
		name        string
		params      []any
		expected    []*router.Attr
		expectError bool
	}{
		{
			name:   "raw",
			params: []any{"foo", "bar"},
			expected: []*router.Attr{
				{Key: "foo", Value: "bar"},
			},
		},
		{
			name: "attr",
			params: []any{
				&router.Attr{Key: "foo", Value: "bar"},
			},
			expected: []*router.Attr{
				{Key: "foo", Value: "bar"},
			},
		},
		{
			name: "mixed",
			params: []any{
				"a", "b",
				&router.Attr{Key: "foo", Value: "bar"},
			},
			expected: []*router.Attr{
				{Key: "a", Value: "b"},
				{Key: "foo", Value: "bar"},
			},
		},
		{
			name: "number",
			params: []any{
				"int", 1,
				"float", 1.5,
			},
			expected: []*router.Attr{
				{Key: "int", Value: "1"},
				{Key: "float", Value: "1.5"},
			},
		},
		{
			name: "model",
			params: []any{
				"foo", &test.Foo{ID: 7},
			},
			expected: []*router.Attr{
				{Key: "foo", Value: "7"},
			},
		},
		{
			name: "invalid",
			params: []any{
				"a",
				&router.Attr{Key: "foo", Value: "bar"},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attrs, err := router.ToAttrs(tc.params)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, attrs)
			}
		})
	}
}

func TestSalusaResolver_Resolve(t *testing.T) {
	testCases := []struct {
		name     string
		origin   string
		path     string
		params   []any
		expected string
	}{
		{
			name:     "raw",
			origin:   "https://example.com",
			path:     "test",
			params:   []any{},
			expected: "https://example.com/test",
		},
		{
			name:     "raw",
			origin:   "https://example.com",
			path:     "test",
			params:   []any{"foo", "bar"},
			expected: "https://example.com/test?foo=bar",
		},
		{
			name:     "raw",
			origin:   "https://example.com",
			path:     "test/{id}",
			params:   []any{"id", 1},
			expected: "https://example.com/test/1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := router.NewResolver(tc.origin, router.New())

			u := r.Resolve(tc.path, tc.params...)
			assert.Equal(t, tc.expected, u)
		})
	}
}
