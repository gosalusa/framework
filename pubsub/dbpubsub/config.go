package dbpubsub

import (
	"context"

	"gosalusa.com/database"
	"gosalusa.com/di"
	"gosalusa.com/pubsub"
)

func Register(ctx context.Context) {
	di.RegisterLazySingletonWith(ctx, func(u database.Update) (pubsub.PubSub, error) {
		return New(u), nil
	})
	pubsub.RegisterTopic(ctx)
}
