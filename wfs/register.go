package wfs

import (
	"context"
	"io/fs"

	"gosalusa.com/di"
)

func RegisterLocal(ctx context.Context, path string) {
	di.RegisterLazySingleton(ctx, func() (fs.FS, error) {
		return NewLocalFS(path), nil
	})
}
