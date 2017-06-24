// +build darwin
// +build !windows

package tcplisten

var newSocketCloexec = newSocketCloexecOld
