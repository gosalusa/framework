package dbtest

import (
	"gosalusa.com/database"
	"gosalusa.com/database/model"
)

type Factory[T model.Model] func(tx database.DB) T

func NewFactory[T model.Model](cb func(tx database.DB) T) Factory[T] {
	return Factory[T](cb)
}

type CountFactory[T model.Model] struct {
	factory Factory[T]
	count   int
}

// func name[T any]() string {
// 	var m T
// 	return reflect.TypeOf(m).String()
// }

func (f Factory[T]) Count(count int) *CountFactory[T] {
	return &CountFactory[T]{
		factory: f,
		count:   count,
	}
}
func (f Factory[T]) State(s func(T)) Factory[T] {
	return func(tx database.DB) T {
		model := f(tx)
		s(model)
		return model
	}
}

func (f Factory[T]) Create(tx database.DB) T {
	m := f(tx)
	model.MustSave(tx, m)
	// err := relationship.InitializeRelationships(m)
	// if err != nil {
	// 	panic(err)
	// }
	return m
}

func (f *CountFactory[T]) Create(tx database.DB) []T {
	models := make([]T, f.count)
	for i := 0; i < f.count; i++ {
		m := f.factory(tx)
		err := model.Save(tx, m)
		if err != nil {
			panic(err)
		}
		models[i] = m
	}
	return models
}
