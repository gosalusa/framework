package nulls

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNullJSON(t *testing.T) {
	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	t.Run("Marshal", func(t *testing.T) {
		i := 5
		tests := []struct {
			name         string
			value        any
			expectedJSON string
			expectErr    bool
		}{
			{"Valid Int", Null[int]{V: 123, Valid: true}, `123`, false},
			{"Invalid Int", Null[int]{}, `null`, false},
			{"Valid String", Null[string]{V: "hello world", Valid: true}, `"hello world"`, false},
			{"Invalid String", Null[string]{}, `null`, false},
			{"Valid Bool True", Null[bool]{V: true, Valid: true}, `true`, false},
			{"Valid Bool False", Null[bool]{V: false, Valid: true}, `false`, false},
			{"Invalid Bool", Null[bool]{}, `null`, false},
			{"Valid Float", Null[float64]{V: 3.14, Valid: true}, `3.14`, false},
			{"Invalid Float", Null[float64]{}, `null`, false},
			{"Valid Struct", Null[testStruct]{V: testStruct{Name: "Bob", Age: 30}, Valid: true}, `{"name":"Bob","age":30}`, false},
			{"Invalid Struct", Null[testStruct]{}, `null`, false},
			{"Valid Time", Null[time.Time]{V: time.Date(2023, 10, 26, 12, 0, 0, 0, time.UTC), Valid: true}, `"2023-10-26T12:00:00Z"`, false},
			{"Invalid Time", Null[time.Time]{}, `null`, false},
			{"Valid Slice", Null[[]int]{V: []int{1, 2, 3}, Valid: true}, `[1,2,3]`, false},
			{"Invalid Slice", Null[[]int]{}, `null`, false},
			{"Valid Pointer", Null[*int]{V: &i, Valid: true}, `5`, false},
			{"Invalid Pointer", Null[*int]{}, `null`, false},
			{"Valid Nil Pointer", Null[*int]{Valid: true}, `null`, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				b, err := json.Marshal(tt.value)
				if tt.expectErr {
					assert.Error(t, err, "json.Marshal")
					return
				}
				assert.NoError(t, err, "json.Marshal")
				assert.Equal(t, tt.expectedJSON, string(b), "json.Marshal")
			})
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		i := 10
		tests := []struct {
			name        string
			jsonInput   string
			targetPtr   any
			expectedVal any
			expectErr   bool
		}{
			{"Valid Int", `456`, new(Null[int]), Null[int]{V: 456, Valid: true}, false},
			{"Null Int", `null`, new(Null[int]), Null[int]{}, false},
			{"Invalid Int", `"abc"`, new(Null[int]), Null[int]{}, true},
			{"Valid String", `"world"`, new(Null[string]), Null[string]{V: "world", Valid: true}, false},
			{"Null String", `null`, new(Null[string]), Null[string]{}, false},
			{"Empty String", `""`, new(Null[string]), Null[string]{V: "", Valid: true}, false},
			{"Invalid String", `123`, new(Null[string]), Null[string]{}, true},
			{"Valid Bool True", `true`, new(Null[bool]), Null[bool]{V: true, Valid: true}, false},
			{"Valid Bool False", `false`, new(Null[bool]), Null[bool]{V: false, Valid: true}, false},
			{"Null Bool", `null`, new(Null[bool]), Null[bool]{}, false},
			{"Invalid Bool", `"true"`, new(Null[bool]), Null[bool]{}, true},
			{"Valid Float", `9.81`, new(Null[float64]), Null[float64]{V: 9.81, Valid: true}, false},
			{"Null Float", `null`, new(Null[float64]), Null[float64]{}, false},
			{"Valid Struct", `{"name":"Alice","age":25}`, new(Null[testStruct]), Null[testStruct]{V: testStruct{Name: "Alice", Age: 25}, Valid: true}, false},
			{"Null Struct", `null`, new(Null[testStruct]), Null[testStruct]{}, false},
			{"Invalid Struct Field", `{"name":"Alice","age":"twenty-five"}`, new(Null[testStruct]), nil, true},
			{"Valid Time", `"2024-01-15T10:30:00Z"`, new(Null[time.Time]), Null[time.Time]{V: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), Valid: true}, false},
			{"Null Time", `null`, new(Null[time.Time]), Null[time.Time]{}, false},
			{"Invalid Time", `"not-a-time"`, new(Null[time.Time]), nil, true},
			{"Valid Slice", `[4, 5, 6]`, new(Null[[]int]), Null[[]int]{V: []int{4, 5, 6}, Valid: true}, false},
			{"Null Slice", `null`, new(Null[[]int]), Null[[]int]{}, false},
			{"Valid Pointer", `10`, new(Null[*int]), Null[*int]{V: &i, Valid: true}, false},
			{"Null Pointer", `null`, new(Null[*int]), Null[*int]{}, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := json.Unmarshal([]byte(tt.jsonInput), tt.targetPtr)

				if tt.expectErr {
					assert.Error(t, err, "json.Unmarshal")
					return
				}
				assert.NoError(t, err, "json.Unmarshal")

				targetVal := reflect.ValueOf(tt.targetPtr).Elem().Interface()
				if !reflect.DeepEqual(targetVal, tt.expectedVal) {
					if ntTarget, ok := targetVal.(Null[time.Time]); ok {
						if ntExpected, ok := tt.expectedVal.(Null[time.Time]); ok {
							if ntTarget.Valid != ntExpected.Valid || (ntTarget.Valid && !ntTarget.V.Equal(ntExpected.V)) {
								t.Errorf("json.Unmarshal: expected %#v, got %#v", tt.expectedVal, targetVal)
							}
							return
						}
					}
					if ntTarget, ok := targetVal.(Null[*int]); ok {
						if ntExpected, ok := tt.expectedVal.(Null[*int]); ok {
							if ntTarget.Valid != ntExpected.Valid {
								t.Errorf("json.Unmarshal: pointer validity mismatch: expected %#v, got %#v", tt.expectedVal, targetVal)
								return
							}
							if ntTarget.Valid {
								if (ntTarget.V == nil && ntExpected.V != nil) || (ntTarget.V != nil && ntExpected.V == nil) || (ntTarget.V != nil && ntExpected.V != nil && *ntTarget.V != *ntExpected.V) {
									t.Errorf("json.Unmarshal: pointer value mismatch: expected %#v, got %#v", tt.expectedVal, targetVal)
									return
								}
							}
							return
						}
					}
					t.Errorf("json.Unmarshal: expected %#v, got %#v", tt.expectedVal, targetVal)
				}
			})
		}
	})
}

func TestNullSQL(t *testing.T) {
	t.Run("Scan", func(t *testing.T) {
		t.Run("Int From Int64", func(t *testing.T) {
			var ni Null[int]
			val := int64(987)
			err := ni.Scan(val)
			assert.NoError(t, err, "Scan int64")
			assert.True(t, ni.Valid, "Scan(%v): Expected Valid=true", val)
			assert.Equal(t, int(val), ni.V, "Scan(%v)", val)
		})

		t.Run("String From String", func(t *testing.T) {
			var ns Null[string]
			val := "db string"
			err := ns.Scan(val)
			assert.NoError(t, err, "Scan string")
			assert.True(t, ns.Valid, "Scan(%q)", val)
			assert.Equal(t, val, ns.V, "Scan(%q)", val)
		})

		t.Run("String From Bytes", func(t *testing.T) {
			var ns Null[string]
			val := []byte("db bytes")
			err := ns.Scan(val)
			assert.NoError(t, err, "Scan string from bytes")
			assert.True(t, ns.Valid, "Scan(%v)", val)
			assert.Equal(t, "db bytes", ns.V, "Scan(%v)", val)
		})

		t.Run("Time From Time", func(t *testing.T) {
			var nt Null[time.Time]
			val := time.Date(2025, 4, 22, 10, 0, 0, 0, time.UTC)
			err := nt.Scan(val)
			assert.NoError(t, err, "Scan time")
			assert.True(t, nt.Valid)
			assert.True(t, nt.V.Equal(val), "Scan(%v)", val)
		})

		t.Run("Any From Nil", func(t *testing.T) {
			ni := Null[int]{V: 100, Valid: true}
			err := ni.Scan(nil)
			assert.NoError(t, err, "Scan nil")
			assert.False(t, ni.Valid, "Scan(nil): expected Valid=false")
			assert.Zero(t, ni.V, "Scan(nil): expected zero V")

			ns := Null[string]{V: "initial", Valid: true}
			err = ns.Scan(nil)
			assert.NoError(t, err, "Scan nil")
			assert.False(t, ns.Valid, "Scan(nil): expected Valid=false")
			assert.Zero(t, ns.V, "Scan(nil): expected zero V")
		})

		t.Run("Int From Unsupported", func(t *testing.T) {
			var ni Null[int]
			err := ni.Scan(struct{}{})
			assert.Error(t, err, "Scan unsupported type")
		})
	})

	t.Run("Value", func(t *testing.T) {
		t.Run("Valid Int", func(t *testing.T) {
			n := Null[int]{V: 55, Valid: true}
			v, err := n.Value()
			assert.NoError(t, err, "Value valid int")
			if dv, ok := v.(int64); !ok || dv != 55 {
				if val, ok := v.(int); !ok || val != 55 {
					t.Errorf("Value(): expected driver.Value=%d, got %T(%v)", 55, v, v)
				}
			}
		})

		t.Run("Invalid Int", func(t *testing.T) {
			var n Null[int]
			v, err := n.Value()
			assert.NoError(t, err, "Value invalid int")
			assert.Nil(t, v, "Value(): expected driver.Value=nil")
		})

		t.Run("Valid String", func(t *testing.T) {
			n := Null[string]{V: "sql value", Valid: true}
			v, err := n.Value()
			assert.NoError(t, err, "Value valid string")
			if dv, ok := v.(string); !ok || dv != "sql value" {
				t.Errorf("Value(): expected driver.Value=%q, got %T(%v)", "sql value", v, v)
			}
		})

		t.Run("Invalid String", func(t *testing.T) {
			var n Null[string]
			v, err := n.Value()
			assert.NoError(t, err, "Value invalid string")
			assert.Nil(t, v, "Value(): expected driver.Value=nil")
		})

		t.Run("Valid Time", func(t *testing.T) {
			tval := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
			n := Null[time.Time]{V: tval, Valid: true}
			v, err := n.Value()
			assert.NoError(t, err, "Value valid time")
			if dv, ok := v.(time.Time); !ok || !dv.Equal(tval) {
				t.Errorf("Value(): expected driver.Value=%v, got %T(%v)", tval, v, v)
			}
		})

		t.Run("Invalid Time", func(t *testing.T) {
			var n Null[time.Time]
			v, err := n.Value()
			assert.NoError(t, err, "Value invalid time")
			assert.Nil(t, v, "Value(): expected driver.Value=nil")
		})

		t.Run("Invalid Struct", func(t *testing.T) {
			type testStruct struct{ V int }
			var n Null[testStruct]
			v, err := n.Value()
			assert.NoError(t, err, "Value invalid struct")
			assert.Nil(t, v, "Value(): expected driver.Value=nil")
		})
	})
}
