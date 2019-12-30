package notification

import (
	"fmt"
	"io/ioutil"
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

// Send send a identity verification link to a user.
func (n *FileNotifier) Send(recipient, subject, body string) error {
	content := fmt.Sprintf("Date: %s\nRecipient: %s\nSubject: %s\nBody: %s", time.Now(), recipient, subject, body)

	err := ioutil.WriteFile(n.path, []byte(content), 0755)

	if err != nil {
		return err
	}
	return nil
}
