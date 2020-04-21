package notification

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// FileNotifier a notifier to send emails to SMTP servers.
type FileNotifier struct {
	path string
}

// NewFileNotifier create an FileNotifier writing the notification into a file.
func NewFileNotifier(configuration schema.FileSystemNotifierConfiguration) *FileNotifier {
	return &FileNotifier{
		path: configuration.Filename,
	}
}

// StartupCheck checks the file provider can write to the specified file
func (n *FileNotifier) StartupCheck() (ok bool, err error) {
	ok = true
	dir := filepath.Dir(n.path)
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, fileNotifierMode); err != nil {
				ok = false
				return
			}
			if err = ioutil.WriteFile(n.path, []byte(""), fileNotifierMode); err != nil {
				ok = false
				return
			}
		} else {
			ok = false
			return
		}
	} else if _, err = os.Stat(n.path); err != nil {
		if os.IsNotExist(err) {
			if err = ioutil.WriteFile(n.path, []byte(""), fileNotifierMode); err != nil {
				ok = false
				return
			}
		} else {
			ok = false
			return
		}
	}
	err = nil
	return
}

// Send send a identity verification link to a user.
func (n *FileNotifier) Send(recipient, subject, body string) error {
	content := fmt.Sprintf("Date: %s\nRecipient: %s\nSubject: %s\nBody: %s", time.Now(), recipient, subject, body)

	err := ioutil.WriteFile(n.path, []byte(content), fileNotifierMode)

	if err != nil {
		return err
	}
	return nil
}
