package hooks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/database"
	"gosalusa.com/database/hooks"
)

type recordingModel struct {
	beforeSave int
	afterSave  int
	afterLoad  int
	saveErr    error
	loadErr    error
	EmbeddedHook
}

func (m *recordingModel) BeforeSave(ctx context.Context, tx database.DB) error {
	m.beforeSave++
	return m.saveErr
}
func (m *recordingModel) AfterSave(ctx context.Context, tx database.DB) error {
	m.afterSave++
	return m.saveErr
}
func (m *recordingModel) AfterLoad(ctx context.Context, tx database.DB) error {
	m.afterLoad++
	return m.loadErr
}

type EmbeddedHook struct {
	before int
	after  int
	loaded int
}

func (e *EmbeddedHook) BeforeSave(ctx context.Context, tx database.DB) error {
	e.before++
	return nil
}
func (e *EmbeddedHook) AfterSave(ctx context.Context, tx database.DB) error {
	e.after++
	return nil
}
func (e *EmbeddedHook) AfterLoad(ctx context.Context, tx database.DB) error {
	e.loaded++
	return nil
}

type plainModel struct{}

func TestBeforeSave(t *testing.T) {
	ctx := context.Background()

	t.Run("model implements BeforeSaver", func(t *testing.T) {
		m := &recordingModel{}
		err := hooks.BeforeSave(ctx, nil, m)
		assert.NoError(t, err)
		assert.Equal(t, 1, m.beforeSave)
		assert.Equal(t, 1, m.EmbeddedHook.before, "embedded fields should be recursed into")
	})

	t.Run("plain model", func(t *testing.T) {
		err := hooks.BeforeSave(ctx, nil, &plainModel{})
		assert.NoError(t, err)
	})

	t.Run("error propagates", func(t *testing.T) {
		m := &recordingModel{saveErr: errors.New("boom")}
		err := hooks.BeforeSave(ctx, nil, m)
		assert.EqualError(t, err, "boom")
	})

	t.Run("pointer to struct", func(t *testing.T) {
		p := &plainModel{}
		err := hooks.BeforeSave(ctx, nil, &p)
		assert.NoError(t, err)
	})

	t.Run("slice", func(t *testing.T) {
		items := []*recordingModel{{}, {}}
		err := hooks.BeforeSave(ctx, nil, items)
		assert.NoError(t, err)
		assert.Equal(t, 1, items[0].beforeSave)
		assert.Equal(t, 1, items[1].beforeSave)
	})

	t.Run("slice element error", func(t *testing.T) {
		items := []*recordingModel{{saveErr: errors.New("slice boom")}}
		err := hooks.BeforeSave(ctx, nil, items)
		assert.EqualError(t, err, "slice boom")
	})

	t.Run("primitive", func(t *testing.T) {
		i := 5
		err := hooks.BeforeSave(ctx, nil, &i)
		assert.NoError(t, err)
	})
}

func TestAfterSave(t *testing.T) {
	ctx := context.Background()

	m := &recordingModel{}
	err := hooks.AfterSave(ctx, nil, m)
	assert.NoError(t, err)
	assert.Equal(t, 1, m.afterSave)
	assert.Equal(t, 1, m.EmbeddedHook.after)

	err = hooks.AfterSave(ctx, nil, &plainModel{})
	assert.NoError(t, err)

	failing := &recordingModel{saveErr: errors.New("after boom")}
	err = hooks.AfterSave(ctx, nil, failing)
	assert.EqualError(t, err, "after boom")
}

func TestAfterLoad(t *testing.T) {
	ctx := context.Background()

	m := &recordingModel{}
	err := hooks.AfterLoad(ctx, nil, m)
	assert.NoError(t, err)
	assert.Equal(t, 1, m.afterLoad)
	assert.Equal(t, 1, m.EmbeddedHook.loaded)

	err = hooks.AfterLoad(ctx, nil, &plainModel{})
	assert.NoError(t, err)

	failing := &recordingModel{loadErr: errors.New("load boom")}
	err = hooks.AfterLoad(ctx, nil, failing)
	assert.EqualError(t, err, "load boom")
}
