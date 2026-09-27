package nulls_test

import (
	"encoding/json"
	"fmt"

	"gosalusa.com/nulls"
)

func ExampleNull() {
	type user struct {
		Name nulls.Null[string] `json:"name"`
		Age  nulls.Null[int]    `json:"age"`
	}

	var u user
	_ = json.Unmarshal([]byte(`{"name":"Salusa","age":null}`), &u)

	data, _ := json.Marshal(u)
	fmt.Println(string(data))

	fmt.Println(u.Name.OrElse("<unknown>"), u.Age.OrElse(-1))
	// Output:
	// {"name":"Salusa","age":null}
	// Salusa -1
}

func ExampleNull_Map() {
	age := nulls.New(21).Map(func(age int) string {
		return fmt.Sprintf("%d years old", age)
	})
	fmt.Println(age.OrElse("unknown age"))

	missing := nulls.Null[int]{}.Map(func(int) string {
		return "never mapped"
	})
	fmt.Println(missing.OrElse("unknown age"))
	// Output:
	// 21 years old
	// unknown age
}
