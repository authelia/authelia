package notification

const fileNotifierMode = 0600
const rfc5322DateTimeLayout = "Mon, 2 Jan 2006 15:04:05 -0700"

var (
	newline           = []byte("\n")
	newlineDouble     = []byte("\n\n")
	boundarySeparator = []byte("--")

	headerDate        = []byte("Date: ")
	headerFrom        = []byte("\nFrom: ")
	headerTo          = []byte("\nTo: ")
	headerSubject     = []byte("\nSubject: ")
	headerMIMEVersion = []byte("\nMIME-Version: 1.0\nContent-Type: multipart/alternative; boundary=")

	headerTextPlain       = []byte("\nContent-Type: text/plain; charset=\"UTF-8\"\nContent-Transfer-Encoding: quoted-printable\nContent-Disposition: inline")
	headerContentTypeHTML = []byte("\nContent-Type: text/html; charset=\"UTF-8\"")
)
