// +build windows

package gmx

import (
	"fmt"
	"net"
)

// on windows we can't use a unix socket. so we use TCP sockets and let the user keep track of what port number is being used by which process

func localSocket() (net.Listener, error) {
	listener, err := net.ListenTCP("tcp", localSocketAddr())
	if err == nil {
		_, ok := listener.Addr().(net.Addr)
		if ok {
			fmt.Println("Listening at ", listener.Addr().String())
		}
	}
	return listener, err
}

func localSocketAddr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0,
	}
}
