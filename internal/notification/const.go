package notification

const fileNotifierMode = 0600
const rfc5322DateTimeLayout = "Mon, 2 Jan 2006 15:04:05 -0700"

var (
	newline           = []byte("\r\n")
	newlineDouble     = []byte("\r\n\r\n")
	boundarySeparator = []byte("--")

	headerDate        = []byte("Date: ")
	headerFrom        = []byte("\r\nFrom: ")
	headerTo          = []byte("\r\nTo: ")
	headerSubject     = []byte("\r\nSubject: ")
	headerMIMEVersion = []byte("\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=")

	headerTextPlain       = []byte("\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\nContent-Transfer-Encoding: quoted-printable\r\nContent-Disposition: inline")
	headerContentTypeHTML = []byte("\r\nContent-Type: text/html; charset=\"UTF-8\"")
)
