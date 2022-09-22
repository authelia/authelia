package notification

const (
	fileNotifierMode = 0600
)

const (
	smtpAUTHMechanismPlain = "PLAIN"
	smtpAUTHMechanismLogin = "LOGIN"

	smtpPortSUBMISSIONS = 465

	smtpCommandDATA     = "DATA"
	smtpCommandHELLO    = "EHLO/HELO"
	smtpCommandSTARTTLS = "STARTTLS"
	smtpCommandAUTH     = "AUTH"
	smtpCommandMAIL     = "MAIL"
	smtpCommandRCPT     = "RCPT"

	smtpEncodingQuotedPrintable = "quoted-printable"
	smtpEncoding8bit            = "8bit"

	smtpContentTypeTextPlain        = "text/plain"
	smtpContentTypeTextHTML         = "text/html"
	smtpFmtContentType              = `%s; charset="UTF-8"`
	smtpFmtContentDispositionInline = "inline"

	smtpExtSTARTTLS = smtpCommandSTARTTLS
	smtpExt8BITMIME = "8BITMIME"
)

const (
	headerContentType             = "Content-Type"
	headerContentDisposition      = "Content-Disposition"
	headerContentTransferEncoding = "Content-Transfer-Encoding"
)

const (
	fmtSMTPGenericError = "error performing %s with the SMTP server: %w"
	fmtSMTPDialError    = "error dialing the SMTP server: %w"
)

var (
	rfc2822DoubleNewLine = []byte("\r\n\r\n")
)
