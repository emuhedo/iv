package io

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// RegisterProtocol registers a file system protocol, such as "zip".
// Protocols are recognized by mount with the syntax: "zip://".
// They are registered by external packages supporting special file systems.
func RegisterProtocol(name string, p Protocol) {
	protocols[name] = p
}

// FileSystem is the interface for a file system provider.
// A directory returns the names of it's content in the reader, with directories ending in a slash.
type FileSystem interface {
	Open(string) (io.ReadCloser, error)
	String() string
}

// fs stores the leading part of the path which is cut from file names.
type fs string

func (o fs) String() string {
	return string(o)
}

func (o fs) Open(name string) (io.ReadCloser, error) {
	f, err := os.Open(o.path(name))
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}
	if fi.IsDir() == false {
		return f, nil
	}

	defer f.Close()
	dir, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(dir))
	for i, d := range dir {
		names[i] = d.Name()
		if d.IsDir() {
			names[i] += "/"
		}
	}
	return ioutil.NopCloser(strings.NewReader(strings.Join(names, "\n"))), nil
}

func (o fs) path(name string) string {
	return filepath.Join(string(o), filepath.FromSlash(name))
}

// Mtab is the mounting table.
// Currently this is shared over all apl instances. This could be changed.
var mtab struct {
	sync.Mutex
	tab []mpoint
}

// Mpoint defines a mount point.
type mpoint struct {
	mpt string
	src FileSystem
}

// Open opens a file or directory from the filesystem.
func Open(name string) (io.ReadCloser, error) {
	if fs, err := lookup(name); err != nil {
		return nil, err
	} else {
		return fs.Open(name)
	}
}

func lookup(name string) (FileSystem, error) {
	mtab.Lock()
	defer mtab.Unlock()

	n := len(mtab.tab)
	if n == 0 {
		return nil, &os.PathError{
			Op:   "open",
			Path: name,
			Err:  fmt.Errorf("no filesystem is mounted"),
		}
	}
	// Files may shadow each other.
	// The last mounted file system is tested first.
	for i := n - 1; i >= 0; i-- {
		t := mtab.tab[i]
		if strings.HasPrefix(name, t.mpt) {
			return t.src, nil
		}
	}
	return nil, &os.PathError{
		Op:   "open",
		Path: name,
		Err:  fmt.Errorf("not found"),
	}
}

// Mount adds a FileSystem to mtab under the given name.
func Mount(mpt string, fs FileSystem) error {
	mtab.Lock()
	defer mtab.Unlock()

	if strings.HasPrefix(mpt, "/") == false {
		return fmt.Errorf("io mount: mount point must start with /: %s", mpt)
	}

	for _, t := range mtab.tab {
		if t.mpt == mpt {
			return fmt.Errorf("mount point already used: %s", mpt)
		}
	}
	mtab.tab = append(mtab.tab, mpoint{mpt, fs})
	return nil
}

// Umount removes the moint point.
func Umount(mpt string) {
	mtab.Lock()
	defer mtab.Unlock()

	n := -1
	for i, t := range mtab.tab {
		if t.mpt == mpt {
			n = i
		}
	}
	if n < 0 {
		return
	}
	mtab.tab = append(mtab.tab[:n], mtab.tab[n+1:]...)
}

var protocols map[string]Protocol

type Protocol interface {
	FileSystem(string) (FileSystem, error)
}
