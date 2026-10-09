package wfs

import (
	"io/fs"
	"os"
	"path"
)

type LocalFS struct {
	root string
	dirFS
}
type dirFS interface {
	fs.StatFS
	fs.ReadFileFS
	fs.ReadDirFS
	fs.ReadLinkFS
}

var _ dirFS = (*LocalFS)(nil)
var _ WriteFileFS = (*LocalFS)(nil)
var _ MkdirFS = (*LocalFS)(nil)
var _ RemoveFS = (*LocalFS)(nil)
var _ OpenFileFS = (*LocalFS)(nil)

func NewLocalFS(root string) *LocalFS {
	return &LocalFS{
		root:  root,
		dirFS: os.DirFS(root).(dirFS),
	}
}

func (l *LocalFS) WriteFile(name string, b []byte) error {
	fullname := l.join(name)
	return l.wrapPathErr(os.WriteFile(fullname, b, 0644), name)
}

func (l *LocalFS) OpenFile(name string, flag int) (File, error) {
	fullname := l.join(name)
	f, err := os.OpenFile(fullname, flag, 0644)
	return f, l.wrapPathErr(err, name)
}

func (l *LocalFS) Mkdir(name string) error {
	fullname := l.join(name)
	return l.wrapPathErr(os.MkdirAll(fullname, 0775), name)
}

func (l *LocalFS) Remove(name string) error {
	fullname := l.join(name)
	return l.wrapPathErr(os.Remove(fullname), name)
}

func (l *LocalFS) Rename(oldname, newname string) error {
	return l.wrapLinkErr(os.Rename(l.join(oldname), l.join(newname)), oldname, newname)
}

// join returns the path for name in dir.
func (l *LocalFS) wrapPathErr(err error, name string) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*os.PathError); ok {
		e.Path = name
	}
	return err
}

// join returns the path for name in dir.
func (l *LocalFS) wrapLinkErr(err error, oldname, newname string) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*os.LinkError); ok {
		e.Old = oldname
		e.New = newname
	}
	return err
}
func (l *LocalFS) join(name string) string {
	return path.Join(l.root, name)
}
