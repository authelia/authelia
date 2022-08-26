package notification

import (
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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

// StartupCheck implements the startup check provider interface.
func (n *FileNotifier) StartupCheck() (err error) {
	dir := filepath.Dir(n.path)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, fileNotifierMode); err != nil {
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
func (n *FileNotifier) Send(recipient mail.Address, subject string, bodyText, _ []byte) error {
	content := fmt.Sprintf("Date: %s\nRecipient: %s\nSubject: %s\nBody: %s", time.Now(), recipient, subject, bodyText)

	return os.WriteFile(n.path, []byte(content), fileNotifierMode)
}
