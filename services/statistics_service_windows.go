// +build windows

package services

import (
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"
)

// getDiskUsage 获取磁盘使用情况（Windows 实现）
func getDiskUsage(path string) (total, used uint64, err error) {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return 0, 0, err
	}

	// 获取驱动器根路径（如 C:\）
	volume := filepath.VolumeName(absPath)
	if volume == "" {
		volume = absPath[:3] // 默认取前3个字符 (C:\)
	} else {
		volume = volume + "\\"
	}

	// 转换为 UTF16
	volumePtr, err := windows.UTF16PtrFromString(volume)
	if err != nil {
		return 0, 0, err
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64

	// 调用 Windows API
	err = windows.GetDiskFreeSpaceEx(
		volumePtr,
		(*uint64)(unsafe.Pointer(&freeBytesAvailable)),
		(*uint64)(unsafe.Pointer(&totalBytes)),
		(*uint64)(unsafe.Pointer(&totalFreeBytes)),
	)
	if err != nil {
		return 0, 0, err
	}

	total = totalBytes
	used = totalBytes - totalFreeBytes

	return total, used, nil
}
