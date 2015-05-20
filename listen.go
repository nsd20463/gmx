// +build !windows

package gmx

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func localSocket() (net.Listener, error) {
	return net.ListenUnix("unix", localSocketAddr())
}

func localSocketAddr() *net.UnixAddr {
	return &net.UnixAddr{
		filepath.Join(os.TempDir(), fmt.Sprintf(".gmx.%d.%d", os.Getpid(), GMX_VERSION)),
		"unix",
	}
}
