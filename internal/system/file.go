package system

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// OpenFile is used to open file, if directory is not exists, it will create it.
func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	dir := filepath.Dir(name)
	if dir != "" {
		err := os.MkdirAll(dir, 0750)
		if err != nil {
			return nil, err
		}
	}
	return os.OpenFile(name, flag, perm) // #nosec
}

// WriteFile is used to write file and call synchronize, it used to write small file.
func WriteFile(filename string, data []byte) error {
	file, err := OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if e := file.Sync(); err == nil {
		err = e
	}
	if e := file.Close(); err == nil {
		err = e
	}
	return err
}

// CopyFile is used to copy file from source path to destination path.
func CopyFile(dst, src string) error {
	same, err := IsSamePath(dst, src)
	if err != nil {
		return err
	}
	if same {
		return nil
	}
	srcFile, err := os.Open(src) // #nosec
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()
	fi, err := srcFile.Stat()
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("\"%s\" is a directory", src)
	}
	dstFile, err := OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()
	fi, err = dstFile.Stat()
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("\"%s\" is a directory", dst)
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return dstFile.Sync()
}

// MoveFile is used to move file from source path to destination path.
// It can move file to the different volume(not use os.Rename).
func MoveFile(dst, src string) error {
	same, err := IsSamePath(dst, src)
	if err != nil {
		return err
	}
	if same {
		return nil
	}
	err = CopyFile(dst, src)
	if err != nil {
		return err
	}
	return os.Remove(src)
}

// IsSamePath is used to check paths are same(absolute).
func IsSamePath(path ...string) (bool, error) {
	l := len(path)
	if l < 2 {
		return false, errors.New("must input more than one path")
	}
	abs, err := filepath.Abs(path[0])
	if err != nil {
		return false, err
	}
	for i := 1; i < l; i++ {
		a, err := filepath.Abs(path[i])
		if err != nil {
			return false, err
		}
		if a != abs {
			return false, nil
		}
	}
	return true, nil
}

// IsFilePath is used to check the target path is a file.
func IsFilePath(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return !fi.IsDir(), nil
}

// IsDirPath is used to check the target path is a directory.
func IsDirPath(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

// IsPathExist is used to check the target path is exist.
func IsPathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// IsPathNotExist is used to check the target path is not exist.
func IsPathNotExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return false, nil
	}
	if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
}
