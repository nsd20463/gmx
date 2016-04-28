// +build !windows

package gmx

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"log"
)

func localSocket() (net.Listener, error) {
	return net.ListenUnix("unix", localSocketAddr())
}

func localSocketAddr() *net.UnixAddr {
	gmxFilePath := filepath.Join(os.TempDir(), fmt.Sprintf(".gmx.%d.%d", os.Getpid(), GMX_VERSION))
	if _, err := os.Stat(gmxFilePath); err == nil {
		log.Printf("File %s already exists, deleting it before creating another\n", gmxFilePath)
		err = os.Remove(gmxFilePath)
		log.Printf("Unable to delete file %s. Error: %q\n", gmxFilePath, err)
	}
	return &net.UnixAddr{
		gmxFilePath,
		"unix",
	}
}
