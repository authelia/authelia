package notification

const (
	fileNotifierMode   = 0666
	fileNotifierHeader = "Date: %s\nRecipient: %s\nSubject: %s\n"
)

const (
	posixNewLine = "\n"
)

var (
	posixDoubleNewLine = []byte(posixNewLine + posixNewLine)
)
