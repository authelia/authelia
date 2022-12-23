package suites

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type message struct {
	ID int `json:"id"`
}

func doGetLinkFromLastMail(t *testing.T) string {
	res := doHTTPGetQuery(t, fmt.Sprintf("%s/messages", MailBaseURL))
	messages := make([]message, 0)
	err := json.Unmarshal(res, &messages)
	assert.NoError(t, err)
	assert.Greater(t, len(messages), 0)

	messageID := messages[len(messages)-1].ID

	res = doHTTPGetQuery(t, fmt.Sprintf("%s/messages/%d.html", MailBaseURL, messageID))

	re := regexp.MustCompile(`<a href="(.+)" class="button">.*<\/a>`)
	matches := re.FindStringSubmatch(string(res))

	require.Len(t, matches, 2, "Number of match for link in email is not equal to one")

	return matches[1]
}
