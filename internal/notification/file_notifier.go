package notification

import (
	"bufio"
	"context"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/templates"
)

// FileNotifier a notifier to send emails to SMTP servers.
type FileNotifier struct {
	path   string
	append bool
}

// NewFileNotifier create an FileNotifier writing the notification into a file.
func NewFileNotifier(configuration schema.NotifierFileSystem) *FileNotifier {
	return &FileNotifier{
		path: configuration.Filename,
	}
}

// StartupCheck implements the startup check provider interface.
func (n *FileNotifier) StartupCheck() (err error) {
	dir := filepath.Dir(n.path)
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, fileNotifierDirMode); err != nil {
				return err
			}
		} else {
			return err
		}
	} else if _, err = os.Stat(n.path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	return os.WriteFile(n.path, []byte(""), fileNotifierMode)
}

// Send send a identity verification link to a user.
func (n *FileNotifier) Send(_ context.Context, recipient mail.Address, subject string, et *templates.EmailTemplate, data any) (err error) {
	var f *os.File

	var flag int

	switch {
	case n.append:
		flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	default:
		flag = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	}

	if f, err = os.OpenFile(n.path, flag, fileNotifierMode); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer f.Close()

	w := bufio.NewWriter(f)

	if _, err = fmt.Fprintf(w, fileNotifierHeader, time.Now(), recipient, subject); err != nil {
		return fmt.Errorf("failed to write to the buffer: %w", err)
	}

	if err = et.Text.Execute(w, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if _, err = w.Write(posixDoubleNewLine); err != nil {
		return fmt.Errorf("failed to write to the buffer: %w", err)
	}

	if err = w.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	if err = f.Sync(); err != nil {
		return fmt.Errorf("failed to sync the file: %w", err)
	}

	return nil
}
