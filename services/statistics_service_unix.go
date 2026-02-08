// +build !windows

package services

import (
	"syscall"
)

// getDiskUsage 获取磁盘使用情况（Unix/Linux 实现）
func getDiskUsage(path string) (total, used uint64, err error) {
	var stat syscall.Statfs_t
	err = syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	// 计算总空间和已使用空间
	total = stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used = total - free

	return total, used, nil
}
