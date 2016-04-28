// +build !windows

package gmx

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func localSocket() (net.Listener, error) {
	addr := localSocketAddr()
	os.Remove(addr.Name) // clear out any leftover use of our filename (it has our PID in it, so I feel no pity)
	return net.ListenUnix("unix", addr)
}

func localSocketAddr() *net.UnixAddr {
	return &net.UnixAddr{
		Name: filepath.Join(os.TempDir(), fmt.Sprintf(".gmx.%d.%d", os.Getpid(), GMX_VERSION)),
		Net:  "unix",
	}
}
