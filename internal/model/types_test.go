package model

import (
	"fmt"
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	args := fosite.Arguments{"abc", "123"}

	x := StringSlicePipeDelimited(args)

	fmt.Println(x)
}

func TestDatabaseModelTypeIP(t *testing.T) {
	ip := IP{}

	value, err := ip.Value()
	assert.Nil(t, value)
	assert.EqualError(t, err, "cannot value model type 'model.IP' with value nil to driver.Value")

	err = ip.Scan("192.168.2.0")
	assert.NoError(t, err)

	assert.True(t, ip.IP.IsPrivate())
	assert.False(t, ip.IP.IsLoopback())
	assert.Equal(t, "192.168.2.0", ip.IP.String())

	value, err = ip.Value()
	assert.NoError(t, err)
	assert.Equal(t, "192.168.2.0", value)

	err = ip.Scan([]byte("127.0.0.0"))
	assert.NoError(t, err)

	assert.False(t, ip.IP.IsPrivate())
	assert.True(t, ip.IP.IsLoopback())
	assert.Equal(t, "127.0.0.0", ip.IP.String())

	err = ip.Scan(1)

	assert.EqualError(t, err, "cannot scan model type '*model.IP' from type 'int' with value '1'")

	err = ip.Scan(nil)
	assert.EqualError(t, err, "cannot scan model type '*model.IP' from value nil: type doesn't support nil values")
}

func TestDatabaseModelTypeNullIP(t *testing.T) {
	ip := NullIP{}

	value, err := ip.Value()
	assert.Nil(t, value)
	assert.NoError(t, err)

	err = ip.Scan("192.168.2.0")
	assert.NoError(t, err)

	assert.True(t, ip.IP.IsPrivate())
	assert.False(t, ip.IP.IsLoopback())
	assert.Equal(t, "192.168.2.0", ip.IP.String())

	value, err = ip.Value()
	assert.NoError(t, err)
	assert.Equal(t, "192.168.2.0", value)

	err = ip.Scan([]byte("127.0.0.0"))
	assert.NoError(t, err)

	assert.False(t, ip.IP.IsPrivate())
	assert.True(t, ip.IP.IsLoopback())
	assert.Equal(t, "127.0.0.0", ip.IP.String())

	err = ip.Scan(1)

	assert.EqualError(t, err, "cannot scan model type '*model.NullIP' from type 'int' with value '1'")

	err = ip.Scan(nil)
	assert.NoError(t, err)
}

func TestDatabaseModelTypeBase64(t *testing.T) {
	b64 := Base64{}

	value, err := b64.Value()
	assert.Equal(t, "", value)
	assert.NoError(t, err)
	assert.Nil(t, b64.Bytes())

	err = b64.Scan(nil)
	assert.EqualError(t, err, "cannot scan model type '*model.Base64' from value nil: type doesn't support nil values")

	err = b64.Scan("###")
	assert.EqualError(t, err, "cannot scan model type '*model.Base64' from type 'string' with value '###': illegal base64 data at input byte 0")

	err = b64.Scan(1)
	assert.EqualError(t, err, "cannot scan model type '*model.Base64' from type 'int' with value '1'")

	err = b64.Scan("YXV0aGVsaWE=")
	assert.NoError(t, err)

	assert.Equal(t, []byte("authelia"), b64.Bytes())
	assert.Equal(t, "YXV0aGVsaWE=", b64.String())

	err = b64.Scan([]byte("c2VjdXJpdHk="))
	assert.NoError(t, err)

	assert.Equal(t, []byte("security"), b64.Bytes())
	assert.Equal(t, "c2VjdXJpdHk=", b64.String())

	err = b64.Scan([]byte("###"))
	assert.NoError(t, err)

	assert.Equal(t, []byte("###"), b64.Bytes())
	assert.Equal(t, "IyMj", b64.String())
}
