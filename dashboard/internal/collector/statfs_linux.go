package collector

import "syscall"

type statfsResult struct {
	Total uint64
	Used  uint64
	Free  uint64
}

func statfs(path string, result *statfsResult) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return err
	}
	result.Total = stat.Blocks * uint64(stat.Bsize)
	result.Free = stat.Bavail * uint64(stat.Bsize)
	result.Used = result.Total - result.Free
	return nil
}
