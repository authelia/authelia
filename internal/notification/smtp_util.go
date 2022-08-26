package notification

import (
	"fmt"
	"io"
	"mime/multipart"
	"mime/quotedprintable"
	"net/textproto"

	"github.com/authelia/authelia/v4/internal/utils"
)

func smtpMIMEHeaders(ext8BITMIME bool, contentType string, data []byte) textproto.MIMEHeader {
	headers := textproto.MIMEHeader{
		headerContentType:        []string{fmt.Sprintf(smtpFmtContentType, contentType)},
		headerContentDisposition: []string{smtpFmtContentDispositionInline},
	}

	characteristics := NewMIMECharacteristics(data)

	if !ext8BITMIME || characteristics.LongLines || characteristics.Characters8BIT {
		headers.Set(headerContentTransferEncoding, smtpEncodingQuotedPrintable)
	} else {
		headers.Set(headerContentTransferEncoding, smtpEncoding8bit)
	}

	return headers
}

func multipartWrite(mwr *multipart.Writer, header textproto.MIMEHeader, data []byte) (err error) {
	var (
		wc io.WriteCloser
		wr io.Writer
	)

	if wr, err = mwr.CreatePart(header); err != nil {
		return err
	}

	switch header.Get(headerContentTransferEncoding) {
	case smtpEncodingQuotedPrintable:
		wc = quotedprintable.NewWriter(wr)
	case smtpEncoding8bit, "":
		wc = utils.NewWriteCloser(wr)
	default:
		return fmt.Errorf("unknown encoding: %s", header.Get(headerContentTransferEncoding))
	}

	if _, err = wc.Write(data); err != nil {
		_ = wc.Close()

		return err
	}

	_ = wc.Close()

	return nil
}

// NewMIMECharacteristics detects the SMTP MIMECharacteristics for the given data bytes.
func NewMIMECharacteristics(data []byte) MIMECharacteristics {
	characteristics := MIMECharacteristics{}

	cl := 0

	n := len(data)

	for i := 0; i < n; i++ {
		cl++

		if cl > 1000 {
			characteristics.LongLines = true
		}

		if data[i] == 10 {
			cl = 0

			if i == 0 || data[i-1] != 13 {
				characteristics.LineFeeds = true
			}
		}

		if data[i] >= 128 {
			characteristics.Characters8BIT = true
		}
	}

	return characteristics
}

// MIMECharacteristics represents specific MIME related characteristics.
type MIMECharacteristics struct {
	LongLines      bool
	LineFeeds      bool
	Characters8BIT bool
}
