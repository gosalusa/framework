package channelpubsub

import (
	"context"

	"gosalusa.com/di"
	"gosalusa.com/pubsub"
)

func Register(ctx context.Context) {
	di.RegisterLazySingleton(ctx, func() (pubsub.PubSub, error) {
		return New(), nil
	})
	pubsub.RegisterTopic(ctx)
}
