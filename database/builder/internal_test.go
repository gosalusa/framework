package builder

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database/model"
	"gosalusa.com/internal/relationship"
)

type barModel struct {
	model.BaseModel
	ID              int `db:"id,primary,autoincrement"`
	InternalModelID int `db:"internal_model_id"`
}

func (barModel) Table() string { return "bar_models" }

type internalModel struct {
	model.BaseModel
	ID   int                   `db:"id,primary,autoincrement"`
	Name string                `db:"name"`
	Bar  *HasOne[*barModel]    `json:"bar"`
	Bars *HasMany[*barModel]   `json:"bars"`
	Foo  *BelongsTo[*barModel] `json:"foo"`
}

func (internalModel) Table() string { return "internal_models" }

type compoundBarModel struct {
	model.BaseModel
	ID  int `db:"id,primary"`
	ID2 int `db:"id2,primary"`
}

func (compoundBarModel) Table() string { return "compound_bar_models" }

type scoperModel struct {
	model.BaseModel
	ID int `db:"id,primary,autoincrement"`
}

func (scoperModel) Table() string { return "scoper_models" }

func (scoperModel) Scopes() []*Scope { return []*Scope{{Name: "global-scope"}} }

type stringerInt int

func (s stringerInt) String() string { return fmt.Sprintf("s%d", s) }

type textInt int

func (t textInt) MarshalText() ([]byte, error) { return []byte(fmt.Sprintf("t%d", t)), nil }

type badTextInt int

func (t *badTextInt) MarshalText() ([]byte, error) { return nil, errors.New("marshal error") }

func TestStringify(t *testing.T) {
	assert.Nil(t, stringify((*stringerInt)(nil)))
	assert.Equal(t, "s5", stringify(stringerInt(5)))
	assert.Equal(t, "t7", stringify(textInt(7)))
	bad := badTextInt(1)
	assert.Equal(t, &bad, stringify(&bad))
	assert.Equal(t, 9, stringify(9))
}

func TestRelatedMap(t *testing.T) {
	rm := newRelatedMap[*barModel]()
	assert.Empty(t, rm.Get(5))
	assert.Equal(t, []*barModel{}, rm.Multi(5, false))
	assert.Nil(t, rm.Single(5, false))

	b := &barModel{ID: 1, InternalModelID: 5}
	rm.Add(5, b)
	assert.Equal(t, []*barModel{b}, rm.Get(5))
	assert.Equal(t, b, rm.Single(5, true))
	assert.Equal(t, []*barModel{b}, rm.Multi(5, true))

	rm.Add(5, &barModel{ID: 2, InternalModelID: 5})
	assert.Len(t, rm.Get(5), 2)

	rm.Add((*barModel)(nil), b)

	assert.Nil(t, rm.Single(999, true))
	assert.Empty(t, rm.Multi(999, true))
}

func TestHasOneOrMany(t *testing.T) {
	m := &internalModel{ID: 5}
	err := relationship.InitializeRelationships(m)
	assert.NoError(t, err)

	r := hasOneOrMany[*barModel]{parent: m, relatedKey: "internal_model_id", parentKey: "id"}
	v, ok := r.parentKeyValue()
	assert.True(t, ok)
	assert.Equal(t, 5, v)

	_, _ = r.relatedKeyValue()

	q := r.Query()
	assert.Equal(t, "bar_models", q.GetTable())

	sub := r.Subquery()
	assert.NotNil(t, sub)

	bad := hasOneOrMany[*barModel]{parent: m, relatedKey: "internal_model_id", parentKey: "missing"}
	assert.Panics(t, func() { bad.Query() })
}

func TestGetValue(t *testing.T) {
	m := &internalModel{ID: 5}
	v, ok := getValue(reflect.ValueOf(m), "ID")
	assert.True(t, ok)
	assert.Equal(t, 5, v.Interface().(int))

	_, ok = getValue(reflect.ValueOf(m), "missing")
	assert.False(t, ok)

	_, ok = getValue(reflect.ValueOf(5), "ID")
	assert.False(t, ok)
}

func TestGetRelation(t *testing.T) {
	m := &internalModel{ID: 5}
	err := relationship.InitializeRelationships(m)
	assert.NoError(t, err)

	r, ok := getRelation(reflect.ValueOf(m), "Bar")
	assert.True(t, ok)
	assert.NotNil(t, r)

	_, ok = getRelation(reflect.ValueOf(m), "missing")
	assert.False(t, ok)

	_, ok = getRelation(reflect.ValueOf(m), "ID")
	assert.False(t, ok)
}

func TestOfType(t *testing.T) {
	relations := []Relationship{
		&HasOne[*barModel]{},
		&HasMany[*barModel]{},
	}
	count := 0
	for r := range ofType[*HasOne[*barModel]](relations) {
		_ = r
		count++
	}
	assert.Equal(t, 1, count)

	i := 0
	for range ofType[*HasMany[*barModel]](relations) {
		i++
		if i == 1 {
			break
		}
	}
}

func TestKeyNameHelpers(t *testing.T) {
	field := reflect.StructField{Name: "Bar", Type: reflect.TypeOf(&HasOne[*barModel]{}), Tag: reflect.StructTag("")}

	k, err := foreignKeyName(field, "foreign", barModel{})
	assert.NoError(t, err)
	assert.Equal(t, "bar_model_id", k)

	_, err = foreignKeyName(field, "foreign", compoundBarModel{})
	assert.Error(t, err)

	_, err = primaryKeyName(field, "local", compoundBarModel{})
	assert.Error(t, err)
}

func TestBelongsToInitialize(t *testing.T) {
	r := &BelongsTo[*barModel]{}
	field := reflect.StructField{Name: "Foo", Type: reflect.TypeOf(&BelongsTo[*barModel]{}), Tag: reflect.StructTag("")}
	err := r.Initialize(&internalModel{}, field)
	assert.NoError(t, err)
	assert.Equal(t, "bar_model_id", r.getParentKey())
	assert.Equal(t, "id", r.getRelatedKey())

	r2 := &BelongsTo[*compoundBarModel]{}
	field2 := reflect.StructField{Name: "Foo", Type: reflect.TypeOf(&BelongsTo[*compoundBarModel]{}), Tag: reflect.StructTag("")}
	err = r2.Initialize(&internalModel{}, field2)
	assert.Error(t, err)
}

func TestHasOneOrManyInitialize(t *testing.T) {
	field := reflect.StructField{Name: "Bar", Type: reflect.TypeOf(&HasOne[*barModel]{}), Tag: reflect.StructTag("")}

	r := &HasOne[*barModel]{}
	err := r.Initialize(&internalModel{}, field)
	assert.NoError(t, err)
	assert.Equal(t, "id", r.getParentKey())
	assert.Equal(t, "internal_model_id", r.getRelatedKey())

	rm := &HasMany[*barModel]{}
	err = rm.Initialize(&internalModel{}, field)
	assert.NoError(t, err)
	assert.Equal(t, "id", rm.getParentKey())
	assert.Equal(t, "internal_model_id", rm.getRelatedKey())

	err = r.Initialize(&compoundBarModel{}, field)
	assert.Error(t, err)
}

func TestWhereHasFallback(t *testing.T) {
	b := NewBuilder()
	b.wheres.withParent(&struct{}{})
	b.wheres.whereHas("Missing", func(q *Builder) *Builder {
		return q.Where("id", "=", 1)
	}, false)
	assert.Len(t, b.wheres.Build(), 1)
}

func TestQueryError(t *testing.T) {
	e := &QueryError{err: errors.New("boom"), query: "SELECT 1"}
	assert.Equal(t, "SELECT 1: boom", e.Error())
	assert.Equal(t, "boom", errors.Unwrap(e).Error())
}

func TestScopesInternal(t *testing.T) {
	s := newScopes().withParent(&scoperModel{})
	assert.Len(t, s.allScopes(), 1)

	s2 := newScopes().withParent(&internalModel{})
	assert.Empty(t, s2.allScopes())
}
