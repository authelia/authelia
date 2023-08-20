package notification

const (
	fileNotifierMode   = 0600
	fileNotifierHeader = "Date: %s\nRecipient: %s\nSubject: %s\n"
)

const (
	posixNewLine = "\n"
)

var (
	posixDoubleNewLine = []byte(posixNewLine + posixNewLine)
)
