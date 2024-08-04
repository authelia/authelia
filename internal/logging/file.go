package logging

import (
	"os"
	"sync"
	"time"
)

func NewFile(name string) *File {
	return &File{
		mu:   sync.Mutex{},
		name: name,
	}
}

type File struct {
	mu   sync.Mutex
	file *os.File
	name string
}

// Open or reopen the file.
func (f *File) Open() (err error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if f.file != nil {
		_ = f.file.Close()
		f.file = nil
	}

	file, err := os.OpenFile(FormatFilePath(f.name, time.Now()), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	f.file = file

	return nil
}

// Close the file.
func (f *File) Close() (err error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	return f.file.Close()
}

// Write to the file.
func (f *File) Write(b []byte) (n int, err error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	return f.file.Write(b)
}
