package logging

import (
	"fmt"
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
		return fmt.Errorf("error opening log file: file is already open")
	}

	var file *os.File

	if file, err = os.OpenFile(FormatFilePath(f.name, time.Now()), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600); err != nil {
		return fmt.Errorf("error opening log file: %w", err)
	}

	f.file = file

	return nil
}

func (f *File) Reopen() (err error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if f.file == nil {
		return fmt.Errorf("error reopening log file: file isn't open")
	}

	if err = f.file.Close(); err != nil {
		return fmt.Errorf("error reopning log file: error closing current log file: %w", err)
	}

	f.file = nil

	var file *os.File

	if file, err = os.OpenFile(FormatFilePath(f.name, time.Now()), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return fmt.Errorf("error reopning log file: error opening new log file: %w", err)
	}

	f.file = file

	return nil
}

// Close the file.
func (f *File) Close() (err error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if f.file == nil {
		return fmt.Errorf("error closing log file: file isn't open")
	}

	if err = f.file.Close(); err != nil {
		return fmt.Errorf("error closing log file: error closing current log file: %w", err)
	}

	f.file = nil

	return nil
}

// Write to the file.
func (f *File) Write(b []byte) (n int, err error) {
	f.mu.Lock()

	defer f.mu.Unlock()

	if f.file == nil {
		return 0, fmt.Errorf("error writing log file: file is not open")
	}

	return f.file.Write(b)
}
