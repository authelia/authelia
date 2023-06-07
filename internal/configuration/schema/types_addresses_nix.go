//go:build linux || freebsd || darwin || netbsd || solaris

package schema

import (
	"fmt"
	"net"
	"syscall"
)

// Listener creates and returns a net.Listener.
func (a *Address) Listener() (ln net.Listener, err error) {
	if a.url == nil {
		return nil, fmt.Errorf("address url is nil")
	}

	if a.socket && a.umask != -1 {
		umask := syscall.Umask(a.umask)

		ln, err = net.Listen(a.Network(), a.NetworkAddress())

		_ = syscall.Umask(umask)

		return ln, err
	}

	return net.Listen(a.Network(), a.NetworkAddress())
}
