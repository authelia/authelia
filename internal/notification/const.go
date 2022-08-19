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
	smtpEncodingBinary          = "binary"
	smtpEncoding7bit            = "7bit"
)

const (
	fmtSMTPGenericError = "error performing %s with the SMTP server: %w"
	fmtSMTPDialError    = "error dialing the SMTP server: %w"
)

var (
	rfc2822DoubleNewLine = []byte("\r\n\r\n")
)
