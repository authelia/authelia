package utils

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	// RFC3339Zero is the default value for time.Time.Unix().
	RFC3339Zero = int64(-62135596800)

	clean   = "clean"
	tagged  = "tagged"
	unknown = "unknown"
)

const (
	period = "."
	https  = "https"
	wss    = "wss"
)

// X.509 consts.
const (
	BlockTypeRSAPrivateKey      = "RSA PRIVATE KEY"
	BlockTypeRSAPublicKey       = "RSA PUBLIC KEY"
	BlockTypeECDSAPrivateKey    = "EC PRIVATE KEY"
	BlockTypePKCS8PrivateKey    = "PRIVATE KEY"
	BlockTypePKIXPublicKey      = "PUBLIC KEY"
	BlockTypeCertificate        = "CERTIFICATE"
	BlockTypeCertificateRequest = "CERTIFICATE REQUEST"
	BlockTypeX509CRL            = "X509 CRL"

	KeyAlgorithmRSA     = "RSA"
	KeyAlgorithmECDSA   = "ECDSA"
	KeyAlgorithmEd25519 = "ED25519"

	HashAlgorithmSHA1   = "SHA1"
	HashAlgorithmSHA256 = "SHA256"
	HashAlgorithmSHA384 = "SHA384"
	HashAlgorithmSHA512 = "SHA512"

	EllipticCurveP224 = "P224"
	EllipticCurveP256 = "P256"
	EllipticCurveP384 = "P384"
	EllipticCurveP521 = "P521"

	EllipticCurveAltP224 = "P-224"
	EllipticCurveAltP256 = "P-256"
	EllipticCurveAltP384 = "P-384"
	EllipticCurveAltP521 = "P-521"
)

const (
	// Hour is an int based representation of the time unit.
	Hour = time.Minute * 60

	// Day is an int based representation of the time unit.
	Day = Hour * 24

	// Week is an int based representation of the time unit.
	Week = Day * 7

	// Year is an int based representation of the time unit.
	Year = Day * 365

	// Month is an int based representation of the time unit.
	Month = Year / 12
)

var (
	// StandardTimeLayouts is the set of standard time layouts used with ParseTimeString.
	StandardTimeLayouts = []string{
		"Jan 2 15:04:05 2006",
		time.DateTime,
		time.RFC3339,
		time.RFC1123Z,
		time.RubyDate,
		time.ANSIC,
		time.DateOnly,
	}

	standardDurationUnits = []string{"ns", "us", "µs", "μs", "ms", "s", "m", "h"}

	reOnlyNumeric      = regexp.MustCompile(`^\d+$`)
	reDurationStandard = regexp.MustCompile(`(?P<Duration>[1-9]\d*?)(?P<Unit>[^\d\s]+)`)
	reNumeric          = regexp.MustCompile(`\d+`)
)

// Duration unit types.
const (
	DurationUnitDays   = "d"
	DurationUnitWeeks  = "w"
	DurationUnitMonths = "M"
	DurationUnitYears  = "y"
)

// Number of hours in particular measurements of time.
const (
	HoursInDay   = 24
	HoursInWeek  = HoursInDay * 7
	HoursInMonth = HoursInDay * 30
	HoursInYear  = HoursInDay * 365
)

const (
	// timeUnixEpochAsMicrosoftNTEpoch represents the unix epoch as a Microsoft NT Epoch.
	// The Microsoft NT Epoch is ticks since Jan 1, 1601 (1 tick is 100ns).
	timeUnixEpochAsMicrosoftNTEpoch uint64 = 116444736000000000
)

var htmlEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&#34;",
	"'", "&#39;",
)

// ErrTimeoutReached error thrown when a timeout is reached.
var ErrTimeoutReached = errors.New("timeout reached")

const (
	windows               = "windows"
	errFmtLinuxNotFound   = "%s %%s: no such file or directory"
	errFmtWindowsNotFound = "%s %%s: The system cannot find the %s specified."

	strStat         = "stat"
	strOpen         = "open"
	strFile         = "file"
	strPath         = "path"
	strIsDir        = "isdir"
	strPathNotFound = "pathnotfound"
	strFileNotFound = "filenotfound"
)
