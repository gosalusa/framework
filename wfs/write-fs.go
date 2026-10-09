package wfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
)

var ErrNotImplemented = errors.New("not implemented")

// Flags to OpenFile wrapping those of the underlying system. Not all
// flags may be implemented on a given system.
const (
	// Exactly one of O_RDONLY, O_WRONLY, or O_RDWR must be specified.
	O_RDONLY int = os.O_RDONLY // open the file read-only.
	O_WRONLY int = os.O_WRONLY // open the file write-only.
	O_RDWR   int = os.O_RDWR   // open the file read-write.
	// The remaining values may be or'ed in to control behavior.
	O_APPEND int = os.O_APPEND // append data to the file when writing.
	O_CREATE int = os.O_CREATE // create a new file if none exists.
	O_EXCL   int = os.O_EXCL   // used with O_CREATE, file must not exist.
	O_SYNC   int = os.O_SYNC   // open for synchronous I/O.
	O_TRUNC  int = os.O_TRUNC  // truncate regular writable file when opened.
)

type File interface {
	fs.File
	io.Writer
}

type WriteFileFS interface {
	fs.FS

	WriteFile(name string, b []byte) error
}

func WriteFile(fsys fs.FS, name string, b []byte) error {
	if fsys, ok := fsys.(WriteFileFS); ok {
		return fsys.WriteFile(name, b)
	}
	if fsys, ok := fsys.(OpenFileFS); ok {
		f, err := fsys.OpenFile(name, O_WRONLY|O_CREATE|O_TRUNC)
		if err != nil {
			return err
		}
		_, err = f.Write(b)
		return err
	}
	return &os.PathError{Op: "writefile", Path: name, Err: ErrNotImplemented}
}

type OpenFileFS interface {
	fs.FS

	OpenFile(name string, flag int) (File, error)
}

func OpenFile(fsys fs.FS, name string, flag int) (File, error) {
	if fsys, ok := fsys.(OpenFileFS); ok {
		return fsys.OpenFile(name, flag)
	}
	return nil, &os.PathError{Op: "openfile", Path: name, Err: ErrNotImplemented}
}

type MkdirFS interface {
	fs.FS

	Mkdir(name string) error
}

func Mkdir(fsys fs.FS, name string) error {
	if fsys, ok := fsys.(MkdirFS); ok {
		return fsys.Mkdir(name)
	}
	return &os.PathError{Op: "mkdir", Path: name, Err: ErrNotImplemented}
}

type RemoveFS interface {
	fs.FS

	Remove(name string) error
}

func Remove(fsys fs.FS, name string) error {
	if fsys, ok := fsys.(RemoveFS); ok {
		return fsys.Remove(name)
	}
	return &os.PathError{Op: "remove", Path: name, Err: ErrNotImplemented}
}

type RenameFS interface {
	fs.FS

	Rename(oldname, newname string) error
}

func Rename(fsys fs.FS, oldname, newname string) error {
	if fsys, ok := fsys.(RenameFS); ok {
		return fsys.Rename(oldname, newname)
	}
	return &os.LinkError{Op: "rename", Old: oldname, New: newname, Err: ErrNotImplemented}
}
