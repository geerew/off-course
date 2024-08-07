package appFs

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/geerew/off-course/utils"
	"github.com/geerew/off-course/utils/types"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

var (
	loggerType = slog.Any("type", types.LogTypeFileSystem)
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AppFs represents a filesystem. It uses afero under the hood, which
// eases testing, as we can dynamically injection to pass a real fs or
// an in-mem fs
type AppFs struct {
	Fs     afero.Fs
	logger *slog.Logger
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// NewAppFs create a new filesystem
func NewAppFs(fs afero.Fs, logger *slog.Logger) *AppFs {
	return &AppFs{
		Fs:     fs,
		logger: logger,
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PathContents defines the fields populated during a path
// scan
type PathContents struct {
	Files       []fs.FileInfo
	Directories []fs.FileInfo
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// SetLogger sets the logger for the filesystem
func (appFs *AppFs) SetLogger(l *slog.Logger) {
	appFs.logger = l
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Open attempts to open a file with a given name from its underlying afero.Fs. It adapts the behavior of
// afero.Fs.Open to match the fs.FS interface from Go's standard library. This function returns an fs.File
// and an error. The returned file can be nil if the error is not nil. If the file does not exist, it
// returns an fs.PathError with fs.ErrNotExist. Other types of errors are returned as is.
func (appFs *AppFs) Open(name string) (fs.File, error) {
	file, err := appFs.Fs.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
		}
		return nil, err
	}
	return file, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ReadDir reads the contents of a path and builds a slice of files and
// directories
func (appFs AppFs) ReadDir(path string, sortResult bool) (*PathContents, error) {
	path = utils.NormalizeWindowsDrive(path)

	items, err := appFs.PathItems(path)
	if err != nil {
		return nil, err
	}

	// Sort the items
	if sortResult {
		sort.Strings(items)
	}

	// Build slice of directories and files
	directories := make([]fs.FileInfo, 0)
	files := make([]fs.FileInfo, 0)

	for _, file := range items {
		fullPath := filepath.Join(path, file)

		if fileStat, err := appFs.Fs.Stat(fullPath); err == nil {
			if fileStat.IsDir() {
				directories = append(directories, fileStat)
			} else {
				files = append(files, fileStat)
			}
		}
	}

	return &PathContents{Files: files, Directories: directories}, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// ReadDirFlat recursively reads a directory down to a certain depth, and returns
// a flat string slice of paths
func (appFs AppFs) ReadDirFlat(path string, depth int) ([]string, error) {
	return appFs.recursivelyReadDir(utils.NormalizeWindowsDrive(path), depth, 0)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// AvailableDrives returns a string slice of available drives on this system. For non-wsl
// systems `gopsutil` is used. For WSL systems, the string slice is generated manually
func (appFs AppFs) AvailableDrives() ([]string, error) {
	// Lookup system
	kernel, err := host.KernelVersion()
	if err != nil {
		appFs.logger.Error(
			"Failed to lookup kernel version",
			slog.String("error", err.Error()),
			loggerType,
		)
		return nil, fmt.Errorf("failed to lookup system information")
	}

	// WSL
	if strings.Contains(strings.ToLower(kernel), "wsl") {
		return appFs.wslDrives()
	}

	// Non-WSL
	return appFs.nonWslDrives()
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PathItems does the common work of opening a path and listing
// its contents
func (appFs AppFs) PathItems(path string) ([]string, error) {
	f, err := appFs.Fs.Open(path)
	if err != nil {
		appFs.logger.Error(
			"Unable to open path",
			slog.String("error", err.Error()),
			slog.String("path", path),
			loggerType,
		)
		return nil, fmt.Errorf("unable to open path")
	}

	// List the items at the path
	items, err := f.Readdirnames(-1)
	if err != nil {
		appFs.logger.Error(
			"Unable to read path",
			slog.String("error", err.Error()),
			slog.String("path", path),
			loggerType,
		)
		return nil, fmt.Errorf("unable to read path")
	}

	return items, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// recursivelyReadDir recursively reads a directory down to a certain depth. It calls itself
// utils the depth is reached, in which case a flat string slice of all found paths (files
// and directories) is returned
func (appFs AppFs) recursivelyReadDir(path string, maxDepth, currDepth int) ([]string, error) {
	// Default max depth to 1
	if maxDepth < 1 {
		maxDepth = 1
	}

	if currDepth == maxDepth {
		return nil, nil
	}

	res := []string{}

	items, err := appFs.PathItems(path)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		fullPath := filepath.Join(path, item)

		if fileStat, err := appFs.Fs.Stat(fullPath); err == nil {
			if fileStat.IsDir() {
				recursiveRes, err := appFs.recursivelyReadDir(fullPath, maxDepth, currDepth+1)
				if err != nil {
					return nil, err
				}

				res = append(res, recursiveRes...)
			} else {
				res = append(res, fullPath)
			}
		}
	}

	return res, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// nonWslDrives builds a list of available drives for non-wsl systems via `gopsutil`
func (appFs AppFs) nonWslDrives() ([]string, error) {
	var drives []string

	partitions, err := disk.Partitions(false)
	if err != nil {
		appFs.logger.Error(
			"Failed to list drives",
			slog.String("error", err.Error()),
			loggerType,
		)
		return nil, fmt.Errorf("failed to list drives")
	}

	for _, partition := range partitions {
		drives = append(drives, partition.Mountpoint)
	}

	return drives, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// wslDrives builds a list of available drives in WSL
func (appFs AppFs) wslDrives() ([]string, error) {
	drives := []string{"/"}

	items, err := appFs.ReadDir("/mnt", true)
	if err != nil {
		return nil, err
	}

	for _, directory := range items.Directories {
		if !strings.Contains(directory.Name(), "wsl") {
			drives = append(drives, filepath.Join("/mnt", directory.Name()))
		}
	}

	return drives, nil
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PartialHash is a function that receives a file path and a chunk size as arguments and
// returns a partial hash of the file, by reading the first, middle, and last
// chunks of the file, as well as two random chunks, and hashes them together
//
// It uses the SHA-256 hashing algorithm from the standard library to calculate the hash
func (appFs AppFs) PartialHash(filePath string, chunkSize int64) (string, error) {
	file, err := appFs.Fs.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	// Append file size to the hash
	fileSize := fileInfo.Size()
	binary.Write(hash, binary.LittleEndian, fileSize)

	// Function to read and hash a chunk at a given position
	readAndHashChunk := func(position int64) error {
		_, err := file.Seek(position, 0)
		if err != nil {
			return err
		}
		chunk := make([]byte, chunkSize)
		n, err := file.Read(chunk)
		if err != nil && err != io.EOF {
			return err
		}
		hash.Write(chunk[:n])
		return nil
	}

	// Read and hash the first chunk
	if err = readAndHashChunk(0); err != nil {
		return "", err
	}

	// Read and hash the middle chunk
	middlePosition := fileSize / 2
	if middlePosition < fileSize {
		if err = readAndHashChunk(middlePosition); err != nil {
			return "", err
		}
	}

	// Read and hash the last chunk
	lastPosition := fileSize - chunkSize
	if lastPosition < 0 {
		lastPosition = 0
	}
	if lastPosition < fileSize {
		if err = readAndHashChunk(lastPosition); err != nil {
			return "", err
		}
	}

	// Random chunks
	additionalPositions := []int64{fileSize / 4, 3 * fileSize / 4}
	for _, position := range additionalPositions {
		if position < fileSize {
			if err = readAndHashChunk(position); err != nil {
				return "", err
			}
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
