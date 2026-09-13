package event

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gosalusa.com/di"
	"gosalusa.com/pubsub"
)

func TestNewDispatch(t *testing.T) {
	t.Run("enqueue", func(t *testing.T) {
		topic := newFakeTopic()
		d := NewDispatch(topic)

		err := d(context.Background(), &TestEvent1{Foo: "hi"})
		assert.NoError(t, err)

		data := topic.enqueuedData()
		assert.Len(t, data, 1)

		e, err := decodeEvent(data[0], map[EventType]reflect.Type{
			(&TestEvent1{}).Type(): reflect.TypeOf(&TestEvent1{}),
		})
		assert.NoError(t, err)
		assert.Equal(t, &TestEvent1{Foo: "hi"}, e)
	})

	t.Run("enqueue error", func(t *testing.T) {
		topic := newFakeTopic()
		topic.enqueueErr = errBoom
		d := NewDispatch(topic)

		err := d(context.Background(), &TestEvent1{Foo: "hi"})
		assert.ErrorIs(t, err, errBoom)
	})

	t.Run("encode error", func(t *testing.T) {
		d := NewDispatch(newFakeTopic())

		err := d(context.Background(), &badFuncEvent{})
		assert.Error(t, err)
	})
}

func TestRegister(t *testing.T) {
	ctx := di.TestDependencyProviderContext()
	ps := newFakePubSub()
	di.RegisterSingleton(ctx, func() pubsub.PubSub { return ps })

	Register(ctx)

	d, err := di.Resolve[Dispatch](ctx)
	assert.NoError(t, err)
	assert.NotNil(t, d)

	err = d(ctx, &TestEvent1{Foo: "reg"})
	assert.NoError(t, err)

	data := ps.topic.enqueuedData()
	assert.Len(t, data, 1)

	e, err := decodeEvent(data[0], map[EventType]reflect.Type{
		(&TestEvent1{}).Type(): reflect.TypeOf(&TestEvent1{}),
	})
	assert.NoError(t, err)
	assert.Equal(t, &TestEvent1{Foo: "reg"}, e)
}
