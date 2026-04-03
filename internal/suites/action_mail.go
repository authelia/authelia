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

type EmailMessagesResponse struct {
	Total         int            `json:"total"`
	Unread        int            `json:"unread"`
	Count         int            `json:"count"`
	MessagesCount int            `json:"messages_count"`
	Start         int            `json:"start"`
	Tags          []string       `json:"tags"`
	Messages      []EmailMessage `json:"messages"`
}

type EmailMessage struct {
	ID          string    `json:"ID"`
	MessageID   string    `json:"MessageID"`
	Read        bool      `json:"Read"`
	From        Address   `json:"From"`
	To          []Address `json:"To"`
	Cc          []Address `json:"Cc"`
	Bcc         []Address `json:"Bcc"`
	ReplyTo     []Address `json:"ReplyTo"`
	Subject     string    `json:"Subject"`
	Created     time.Time `json:"Created"`
	Tags        []string  `json:"Tags"`
	Size        int       `json:"Size"`
	Attachments int       `json:"Attachments"`
	Snippet     string    `json:"Snippet"`
}

type Address struct {
	Name    string `json:"Name"`
	Address string `json:"Address"`
}

func (m *EmailMessage) GetContentReader() (reader io.ReadCloser, err error) {
	client := NewHTTPClient()

	req, err := http.NewRequest(fasthttp.MethodGet, fmt.Sprintf("%s/view/%s.html", MailBaseURL, m.ID), nil)
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
	t.Helper()

	msg := doGetLastEmailMessageWithSubject(t, subject)

	reader, err := msg.GetContentReader()
	require.NoError(t, err)

	defer reader.Close()

	node, err = html.Parse(reader)
	require.NoError(t, err)

	return getHTMLNodeWithID(node, id)
}

func doGetOneTimeCodeFromLastMail(t *testing.T) string {
	t.Helper()

	element := doGetEmailNodeID(t, "[Authelia] Confirm your identity", "one-time-code")

	require.NotNil(t, element)
	require.NotNil(t, element.FirstChild)
	require.NotNil(t, element.LastChild)
	require.Equal(t, element.FirstChild, element.LastChild)

	return element.FirstChild.Data
}

//nolint:unused
func doGetOneTimeCodeLinkRevokeFromLastMail(t *testing.T) string {
	t.Helper()

	element := doGetEmailNodeID(t, "[Authelia] Confirm your identity", "link-revoke")

	require.NotNil(t, element)

	return doGetNodeAttribute(t, element, "href")
}

func doGetResetPasswordJWTLinkFromLastEmail(t *testing.T) string {
	t.Helper()

	element := doGetEmailNodeID(t, "[Authelia] Reset your password", "link")

	require.NotNil(t, element)

	return doGetNodeAttribute(t, element, "href")
}

//nolint:unused
func doGetResetPasswordJWTLinkRevokeFromLastEmail(t *testing.T) string {
	t.Helper()

	element := doGetEmailNodeID(t, "[Authelia] Reset your password", "link-revoke")

	require.NotNil(t, element)

	return doGetNodeAttribute(t, element, "href")
}

func doGetNodeAttribute(t *testing.T, node *html.Node, key string) string {
	t.Helper()

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
	t.Helper()

	messages := doGetEmailMessages(t)

	for i := len(messages) - 1; i >= 0; i-- {
		if subject == messages[i].Subject && !messages[i].Read {
			return messages[i]
		}
	}

	require.Fail(t, "Didn't find the message.")

	return message
}

func doGetEmailMessages(t *testing.T) []EmailMessage {
	t.Helper()

	var emr EmailMessagesResponse

	res := doHTTPGetQuery(t, fmt.Sprintf("%s/api/v1/messages", MailBaseURL))

	err := json.Unmarshal(res, &emr)

	require.NoError(t, err)

	return emr.Messages
}
