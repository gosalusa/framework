package option_test

import (
	"encoding/json"
	"fmt"

	"gosalusa.com/option"
)

func ExampleOption() {
	type user struct {
		Name option.Option[string] `json:"name"`
		Age  option.Option[int]    `json:"age"`
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

func ExampleOption_Map() {
	age := option.Some(21).Map(func(age int) string {
		return fmt.Sprintf("%d years old", age)
	})
	fmt.Println(age.OrElse("unknown age"))

	missing := option.Option[int]{}.Map(func(int) string {
		return "never mapped"
	})
	fmt.Println(missing.OrElse("unknown age"))
	// Output:
	// 21 years old
	// unknown age
}
