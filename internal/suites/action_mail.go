package suites

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
)

type EmailMessage struct {
	ID         int       `json:"id"`
	Sender     string    `json:"sender"`
	Recipients []string  `json:"recipients"`
	Subject    string    `json:"subject"`
	Size       string    `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
}

func (m *EmailMessage) GetContentReader() (reader io.ReadCloser, err error) {
	client := NewHTTPClient()

	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/messages/%d.html", MailBaseURL, m.ID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(fasthttp.HeaderAccept, "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (m *EmailMessage) GetContent() (content []byte, err error) {
	reader, err := m.GetContentReader()

	defer func() {
		_ = reader.Close()
	}()

	content, _ = io.ReadAll(reader)

	return content, nil
}

func getHTMLNodeAttr(n *html.Node, key string) (value string, ok bool) {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val, true
		}
	}

	return "", false
}

func getHTMLNodeHasID(n *html.Node, id string) bool {
	if n.Type == html.ElementNode {
		value, ok := getHTMLNodeAttr(n, "id")
		if ok && value == id {
			return true
		}
	}

	return false
}

func getHTMLNodeWithID(n *html.Node, id string) (found *html.Node) {
	if getHTMLNodeHasID(n, id) {
		return n
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		found = getHTMLNodeWithID(c, id)
		if found != nil {
			return found
		}
	}

	return nil
}

func doGetEmailNodeID(t *testing.T, subject, id string) (node *html.Node) {
	msg := doGetLastEmailMessageWithSubject(t, subject)

	reader, err := msg.GetContentReader()
	require.NoError(t, err)

	defer reader.Close()

	node, err = html.Parse(reader)
	require.NoError(t, err)

	return getHTMLNodeWithID(node, id)
}

func doGetOneTimeCodeFromLastMail(t *testing.T) string {
	element := doGetEmailNodeID(t, "[Authelia] Confirm your identity", "one-time-code")

	require.NotNil(t, element)
	require.NotNil(t, element.FirstChild)
	require.NotNil(t, element.LastChild)
	require.Equal(t, element.FirstChild, element.LastChild)

	return element.FirstChild.Data
}

//nolint:unused
func doGetOneTimeCodeLinkRevokeFromLastMail(t *testing.T) string {
	element := doGetEmailNodeID(t, "[Authelia] Confirm your identity", "link-revoke")

	require.NotNil(t, element)

	return doGetNodeAttribute(t, element, "href")
}

func doGetResetPasswordJWTLinkFromLastEmail(t *testing.T) string {
	element := doGetEmailNodeID(t, "[Authelia] Reset your password", "link")

	require.NotNil(t, element)

	return doGetNodeAttribute(t, element, "href")
}

//nolint:unused
func doGetResetPasswordJWTLinkRevokeFromLastEmail(t *testing.T) string {
	element := doGetEmailNodeID(t, "[Authelia] Reset your password", "link-revoke")

	require.NotNil(t, element)

	return doGetNodeAttribute(t, element, "href")
}

func doGetNodeAttribute(t *testing.T, node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key != key {
			continue
		}

		return attr.Val
	}

	require.Fail(t, fmt.Sprintf("Attribute '%s' Not Found On Node", key))

	return ""
}

func doGetLastEmailMessageWithSubject(t *testing.T, subject string) (message EmailMessage) {
	messages := doGetEmailMessages(t)

	for i := len(messages) - 1; i >= 0; i-- {
		if subject == messages[i].Subject {
			return messages[i]
		}
	}

	require.Fail(t, "Didn't find the message.")

	return message
}

func doGetEmailMessages(t *testing.T) (messages []EmailMessage) {
	messages = make([]EmailMessage, 0)

	res := doHTTPGetQuery(t, fmt.Sprintf("%s/messages", MailBaseURL))

	err := json.Unmarshal(res, &messages)

	require.NoError(t, err)

	return messages
}
