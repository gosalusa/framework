package models

import (
	"context"

	"gosalusa.com/database/builder"
	"gosalusa.com/database/model"
	"gosalusa.com/database/model/modeldi"
	"gosalusa.com/static/template/app/providers"
)

//go:generate spice generate:migration
type Foo struct {
	model.BaseModel

	ID int `json:"id" db:"id,primary,autoincrement"`
}

func init() {
	providers.Add(modeldi.Register[*Foo])
}

func FooQuery(ctx context.Context) *builder.ModelBuilder[*Foo] {
	return builder.From[*Foo]().WithContext(ctx)
}
