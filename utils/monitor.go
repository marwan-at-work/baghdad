package utils

import (
	"syscall"
	"time"
)

const (
	// B byte
	B = 1
	// KB kilobytes
	KB = 1024 * B
	// MB megabytes
	MB = 1024 * KB
	// GB gigabytes
	GB = 1024 * MB
)

type diskStatus struct {
	All  float64
	Used float64
	Free float64
}

func diskUsage(path string) (disk diskStatus) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return
	}

	all := fs.Blocks * uint64(fs.Bsize)
	free := fs.Bfree * uint64(fs.Bsize)
	used := disk.All - disk.Free

	disk.All = float64(all) / float64(GB)
	disk.Free = float64(free) / float64(GB)
	disk.Used = float64(used) / float64(GB)
	return
}

// Monitor takes a channel and monitors the root directory's disk usage.
// if the disk is less than 3Gigs free, it will send a message to the channel.
func Monitor(ch chan bool) {
	for {
		time.Sleep(time.Minute)
		disk := diskUsage("/")
		if disk.Free < 3.0 {
			ch <- true
		}
	}
}
