package suites

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/matryer/is"
)

type message struct {
	ID int `json:"id"`
}

func doGetLinkFromLastMail(t *testing.T) string {
	is := is.New(t)
	res := doHTTPGetQuery(t, fmt.Sprintf("%s/messages", MailBaseURL))
	messages := make([]message, 0)
	err := json.Unmarshal(res, &messages)
	is.NoErr(err)
	is.True(len(messages) > 0)

	messageID := messages[len(messages)-1].ID

	res = doHTTPGetQuery(t, fmt.Sprintf("%s/messages/%d.html", MailBaseURL, messageID))

	re := regexp.MustCompile(`<a href="(.+)" class="button">.*</a>`)
	matches := re.FindStringSubmatch(string(res))
	is.True(len(matches) == 2)

	return matches[1]
}
