//go:build !linux && !freebsd && !darwin && !netbsd && !solaris

package schema

import (
	"fmt"
	"net"
)

// Listener creates and returns a net.Listener.
func (a *Address) Listener() (ln net.Listener, err error) {
	if a.url == nil {
		return nil, fmt.Errorf("address url is nil")
	}

	return net.Listen(a.Network(), a.NetworkAddress())
}
