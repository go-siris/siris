// +build windows

package tcplisten

import (
	"fmt"
	"syscall"
)

func newSocketCloexec(domain, typ, proto int) (int, error) {
	fd, err := syscall.Socket(domain, typ, proto)
	if err == nil {
		return int(fd), nil
	}

	if err == syscall.EPROTONOSUPPORT || err == syscall.EINVAL {
		return newSocketCloexecOld(domain, typ, proto)
	}

	return -1, fmt.Errorf("cannot create listening unblocked socket: %s", err)
}
