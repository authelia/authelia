//go:build linux || freebsd || darwin || netbsd || solaris

package schema

import (
	"fmt"
	"net"
	"os"
	"strconv"
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
	if a.url.Scheme == AddressSchemeFileDescriptor {
		fd := os.NewFile(uintptr(a.port), strconv.Itoa(a.port))
		defer fd.Close()
		return net.FileListener(fd)
	}

	return net.Listen(a.Network(), a.NetworkAddress())
}
