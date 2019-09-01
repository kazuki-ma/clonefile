// +build windows

package clonefile

import (
	"io"
	"ioutil"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	initialCloneRegionSize int64 = 1024 * 1024
	finalCloneRegionSize   int64 = 4 * 1024 // = minimum refs cluster size.
)

// FSCTL_DUPLICATE_EXTENTS_TO_FILE :
// Instructs the file system to copy a range of file bytes on behalf of an application.
//
// https://docs.microsoft.com/windows/win32/api/winioctl/ni-winioctl-fsctl_duplicate_extents_to_file
const FSCTL_DUPLICATE_EXTENTS_TO_FILE = 623428

// DUPLICATE_EXTENTS_DATA :
// Contains parameters for the FSCTL_DUPLICATE_EXTENTS control code that performs the Block Cloning operation.
//
// https://docs.microsoft.com/windows/win32/api/winioctl/ns-winioctl-duplicate_extents_data
type DUPLICATE_EXTENTS_DATA struct {
	FileHandle       windows.Handle
	SourceFileOffset int64
	TargetFileOffset int64
	ByteCount        int64
}

func ByPath(src, dst string) (success bool, err error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return
	}

	dstFile, err := os.Open(dst)
	if err != nil {
		return
	}

	return ByFD(srcFile, dstFile)
}

func ByFD(src, dst *os.File) (success bool, err error) {
	srcStat, err := src.Stat()
	if err != nil {
		return
	}
	fileSize := srcStat.Size()

	err = setFileSize(dst, srcStat.Size())
	if err != nil {
		return
	}

	i := int64(0)

	// Requirement
	// The source and destination regions must begin and end at a cluster boundary.
	// API limitation requires less than 4GiB.
	// So we try from initialCloneRegionSize and reduces clonesize to finalCloneRegionSize(4KiB).
	// see https://docs.microsoft.com/ja-jp/windows/win32/fileio/block-cloning
	for cloneRegionSize := initialCloneRegionSize; cloneRegionSize >= finalCloneRegionSize; cloneRegionSize /= 2 {
		for ; i < fileSize; i += cloneRegionSize {
			// Ovrerflow of copy size is safe and just ignored by windows.
			if err = callDuplicateExtentsToFile(src, dst, i, cloneRegionSize); err != nil {
				break
			}
		}

		if i >= fileSize {
			break // clone finished
		}
	}

	return err == nil, err
}

// CheckCloneFileSupported runs explicit test of clone file on supplied directory.
// This function creates some (src and dst) file in the directory and remove after test finished.
//
// If check failed (e.g. directory is read-only), returns err.
func CheckCloneFileSupported(dir string) (supported bool, err error) {
	src, err := ioutil.TempFile(dir, "src")
	if err != nil {
		return false, err
	}
	defer os.Remove(src.Name())

	dst, err := ioutil.TempFile(dir, "dst")
	if err != nil {
		return false, err
	}
	defer os.Remove(dst.Name())

	return ByFD(src, dst)
}

var null = (*byte)(unsafe.Pointer(nil))

// call FSCTL_DUPLICATE_EXTENTS_TO_FILE IOCTL
// see https://docs.microsoft.com/en-us/windows/win32/api/winioctl/ni-winioctl-fsctl_duplicate_extents_to_file
func callDuplicateExtentsToFile(src, dst *os.File, base int64, size int64) error {
	var bytesReturned uint32
	var overwrapped windows.Overlapped

	request := DUPLICATE_EXTENTS_DATA{
		FileHandle:       windows.Handle(src.Fd()),
		SourceFileOffset: base,
		TargetFileOffset: base,
		ByteCount:        size,
	}

	return windows.DeviceIoControl(
		windows.Handle(dst.Fd()),
		FSCTL_DUPLICATE_EXTENTS_TO_FILE,
		(*byte)(unsafe.Pointer(&request)),
		uint32(unsafe.Sizeof(request)),
		null,
		0,
		&bytesReturned,
		&overwrapped)
}

func setFileSize(target *os.File, size int64) error {
	_, err := target.Seek(size, io.SeekStart)
	if err != nil {
		return err
	}

	return windows.SetEndOfFile(windows.Handle(target.Fd()))
}
