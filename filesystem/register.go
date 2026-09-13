package filesystem

import (
	"context"
	"io/fs"
	"os"

	"gosalusa.com/di"
)

type Config interface {
	FS() fs.FS
}

type LocalFS struct {
	Root string
}

func NewLocalFS(root string) *LocalFS {
	return &LocalFS{
		Root: root,
	}
}

func (l *LocalFS) FS() fs.FS {
	return os.DirFS(l.Root)
}

func Register(ctx context.Context, cfg Config) {
	di.RegisterLazySingleton(ctx, func() (fs.FS, error) {
		return cfg.FS(), nil
	})
}
