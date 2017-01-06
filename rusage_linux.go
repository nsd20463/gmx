package gmx

// syscall.Rusage instrumentation for linux

import "syscall"
import "time"

// constant copied from C header <linux/resource.h> to avoid requiring cgo
// (this is a kernel API, so it is going to be very very stable)
const RUSAGE_SELF = 0

func init() {
	// publish the total CPU time (userspace+system) used by this process, in seconds
	Publish("runtime.cpu.time", rusageCpu)
}

func rusageCpu() interface{} {
	var rusage syscall.Rusage
	err := syscall.Getrusage(RUSAGE_SELF, &rusage)
	if err != nil {
		return -1 // return an obviously wrong result
	}

	total := time.Second * time.Duration(rusage.Stime.Sec+rusage.Utime.Sec)
	total += time.Microsecond * time.Duration(rusage.Stime.Usec+rusage.Utime.Usec)

	return total.Seconds()
}
