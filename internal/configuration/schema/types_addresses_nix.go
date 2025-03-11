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

	var create func() (ln net.Listener, err error)

	if a.fd != nil {
		create = func() (ln net.Listener, err error) {
			fd := os.NewFile(uintptr(*a.fd), strconv.FormatUint(*a.fd, 10))

			defer func() {
				_ = fd.Close()
			}()

			return net.FileListener(fd)
		}
	} else {
		create = func() (ln net.Listener, err error) {
			return net.Listen(a.Network(), a.NetworkAddress())
		}
	}

	return a.listenerWithUmask(create)
}

func (a *Address) listenerWithUmask(create func() (net.Listener, error)) (ln net.Listener, err error) {
	if a.umask == -1 {
		return create()
	}

	umask := syscall.Umask(a.umask)

	defer func() {
		_ = syscall.Umask(umask)
	}()

	return create()
}
