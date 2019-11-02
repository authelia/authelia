package suites

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/stretchr/testify/assert"
)

type message struct {
	ID int `json:"id"`
}

func doGetLinkFromLastMail(s *SeleniumSuite) string {
	res := doHTTPGetQuery(s, fmt.Sprintf("%s/messages", MailBaseURL))
	messages := make([]message, 0)
	err := json.Unmarshal(res, &messages)
	assert.NoError(s.T(), err)
	assert.Greater(s.T(), len(messages), 0)

	messageID := messages[len(messages)-1].ID

	res = doHTTPGetQuery(s, fmt.Sprintf("%s/messages/%d.html", MailBaseURL, messageID))

	re := regexp.MustCompile(`<a href="(.+)" class="button">.*<\/a>`)
	matches := re.FindStringSubmatch(string(res))

	if len(matches) != 2 {
		log.Fatal("Number of match for link in email is not equal to one")
	}
	return matches[1]
}
