package filemgr

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"

	"project/internal/task"
	"project/internal/xpanic"
)

// name about task
const (
	TaskNameCopy   = "copy"
	TaskNameMove   = "move"
	TaskNameDelete = "delete"
	TaskNameZip    = "zip"
	TaskNameUnZip  = "unzip"
)

// ErrCtrl is used to tell Move or Copy function how to control the same file,
// directory, or copy, move error. src and dst is the absolute file path.
// err and fileStat in stats maybe nil.
type ErrCtrl func(ctx context.Context, typ uint8, err error, stats *SrcDstStat) uint8

// errors about ErrCtrl
const (
	_                    uint8 = iota
	ErrCtrlSameFile            // two same name file
	ErrCtrlSameFileDir         // same src file name with dst directory
	ErrCtrlSameDirFile         // same src directory name with dst file name
	ErrCtrlCollectFailed       // appear error in collect path information
	ErrCtrlCopyFailed          // appear error in copy task
	ErrCtrlMoveFailed          // appear error in move task
	ErrCtrlDeleteFailed        // appear error in delete task
	ErrCtrlZipFailed           // appear error in zip task
	ErrCtrlUnZipFailed         // appear error in unzip task
)

// operation code about ErrCtrl
const (
	ErrCtrlOpInvalid uint8 = iota
	ErrCtrlOpReplace       // replace same name file
	ErrCtrlOpRetry         // try to copy or move again
	ErrCtrlOpSkip          // skip same name file, directory or copy
	ErrCtrlOpCancel        // cancel whole copy or move operation
)

var (
	// ReplaceAll is used to replace all src file to dst file.
	ReplaceAll = func(_ context.Context, typ uint8, _ error, _ *SrcDstStat) uint8 {
		if typ == ErrCtrlSameFile {
			return ErrCtrlOpReplace
		}
		return ErrCtrlOpSkip
	}

	// SkipAll is used to skip all existed file or other error.
	SkipAll = func(context.Context, uint8, error, *SrcDstStat) uint8 {
		return ErrCtrlOpSkip
	}

	// Cancel is used to cancel current task if appear some error.
	// tests about filemgr will use it.
	Cancel = func(_ context.Context, typ uint8, err error, _ *SrcDstStat) uint8 {
		fmt.Println(typ, err)
		return ErrCtrlOpCancel
	}
)

// for calculate task progress
var (
	zeroFloat = big.NewFloat(0)
	oneFloat  = big.NewFloat(1)
)

// stat is used to get file stat, if err is NotExist, it will return nil
// error and os.FileInfo, usually it used to check destination path.
func stat(name string) (os.FileInfo, error) {
	stat, err := os.Stat(name)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return stat, nil
}

// validatePaths is used to make sure all paths is in the same directory,
// if passed, it will sort paths and return base path.
func validatePaths(paths []string) (string, error) {
	pathsLen := len(paths)
	pathsMap := make(map[string]struct{}, pathsLen)
	var basePath string
	for i := 0; i < pathsLen; i++ {
		if paths[i] == "" {
			return "", errors.New("appear empty path in source paths")
		}
		// make sure all source path is absolute
		absPath, err := filepath.Abs(paths[i])
		if err != nil {
			return "", errors.Wrap(err, "failed to get absolute file path")
		}
		paths[i] = absPath
		if i == 0 {
			pathsMap[absPath] = struct{}{}
			basePath = filepath.Dir(absPath)
			continue
		}
		// if appear root path, that must only one in paths
		if isRoot(absPath) {
			return "", errors.Errorf("appear root path \"%s\"", absPath)
		}
		// check file path is already exists
		if _, ok := pathsMap[absPath]; ok {
			return "", errors.Errorf("appear the same path \"%s\"", absPath)
		}
		// compare directory is same as the first path
		dir := filepath.Dir(absPath)
		if dir != basePath {
			const format = "\"%s\" and \"%s\" are not in the same directory"
			return "", errors.Errorf(format, absPath, paths[0])
		}
		pathsMap[absPath] = struct{}{}
	}
	sort.Strings(paths)
	return basePath, nil
}

// isRoot is used to check path is root directory.
func isRoot(path string) bool {
	switch path {
	case "/", "\\": // Unix
		return true
	case filepath.VolumeName(path) + "\\": // Windows: C:\
		return true
	case filepath.VolumeName(path): // UNC: \\host\share
		return true
	}
	return false
}

type file struct {
	path  string // absolute
	stat  os.FileInfo
	files []*file
}

type fileStat struct {
	path string // absolute
	stat os.FileInfo
}

// zipFiles is used to sort files by name in zip file.
type zipFiles []*zip.File

func (zf zipFiles) Len() int           { return len(zf) }
func (zf zipFiles) Less(i, j int) bool { return zf[i].Name < zf[j].Name }
func (zf zipFiles) Swap(i, j int)      { zf[i], zf[j] = zf[j], zf[i] }

// noticePs contains three parameters about notice function.
type noticePs struct {
	ctx     context.Context
	task    *task.Task
	errCtrl ErrCtrl
}

// SrcDstStat contains absolute path and file stat about src and dst.
type SrcDstStat struct {
	SrcAbs  string // "E:\file.dat" "E:\file", last will not be "/ or "\"
	DstAbs  string
	SrcStat os.FileInfo
	DstStat os.FileInfo // check destination file or directory is exists
}

// ErrUserCanceled is an error about user cancel task.
var ErrUserCanceled = fmt.Errorf("user canceled")

// noticeSameFile is used to notice appear same name file.
func noticeSameFile(ps *noticePs, stats *SrcDstStat) (replace bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	switch code := ps.errCtrl(ps.ctx, ErrCtrlSameFile, nil, stats); code {
	case ErrCtrlOpReplace:
		replace = true
	case ErrCtrlOpSkip:
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown same file operation code: %d", code)
	}
	return
}

// noticeSameFileDir is used to notice appear same name about src file and dst dir.
func noticeSameFileDir(ps *noticePs, stats *SrcDstStat) (retry bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	switch code := ps.errCtrl(ps.ctx, ErrCtrlSameFileDir, nil, stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip:
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown same file dir operation code: %d", code)
	}
	return
}

// noticeSameDirFile is used to notice appear same name about src dir and dst file.
func noticeSameDirFile(ps *noticePs, stats *SrcDstStat) (retry bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	switch code := ps.errCtrl(ps.ctx, ErrCtrlSameDirFile, nil, stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip:
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown same dir file operation code: %d", code)
	}
	return
}

// noticeFailedToCollect is used to notice appear some error in collectDirInfo.
// stats to errCtrl can only get SrcAbs.
func noticeFailedToCollect(ps *noticePs, path string, extError error) (skip bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	stats := SrcDstStat{SrcAbs: path}
	switch code := ps.errCtrl(ps.ctx, ErrCtrlCollectFailed, extError, &stats); code {
	case ErrCtrlOpSkip:
		skip = true
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to collect operation code: %d", code)
	}
	return
}

// noticeFailedToCopy is used to notice appear some error about copy.
// stats.DstStat maybe nil.
func noticeFailedToCopy(ps *noticePs, stats *SrcDstStat, extError error) (retry bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	switch code := ps.errCtrl(ps.ctx, ErrCtrlCopyFailed, extError, stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip:
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to copy operation code: %d", code)
	}
	return
}

// noticeFailedToMove is used to notice appear some error about move.
// stats.DstStat maybe nil.
func noticeFailedToMove(ps *noticePs, stats *SrcDstStat, extError error) (retry bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	switch code := ps.errCtrl(ps.ctx, ErrCtrlMoveFailed, extError, stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip:
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to move operation code: %d", code)
	}
	return
}

// noticeFailedToDelete is used to notice appear some error about delete.
// stats to errCtrl can only get SrcAbs.
func noticeFailedToDelete(ps *noticePs, path string, extError error) (retry bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	stats := SrcDstStat{SrcAbs: path}
	switch code := ps.errCtrl(ps.ctx, ErrCtrlDeleteFailed, extError, &stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip:
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to delete operation code: %d", code)
	}
	return
}

// noticeFailedToZip is used to notice appear some error about zip.
// stats to errCtrl can only get SrcAbs.
func noticeFailedToZip(ps *noticePs, path string, extError error) (retry bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	stats := SrcDstStat{SrcAbs: path}
	switch code := ps.errCtrl(ps.ctx, ErrCtrlZipFailed, extError, &stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip:
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to zip operation code: %d", code)
	}
	return
}

// noticeFailedToUnZip is used to notice appear some error about unzip.
// stats to errCtrl can only get SrcAbs.
func noticeFailedToUnZip(ps *noticePs, path string, extError error) (retry bool, err error) {
	ps.task.Pause()
	defer ps.task.Continue()
	stats := SrcDstStat{SrcAbs: path}
	switch code := ps.errCtrl(ps.ctx, ErrCtrlUnZipFailed, extError, &stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip:
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to unzip operation code: %d", code)
	}
	return
}

// ioCopy is used to copy with task.Paused and add function is used to update task progress.
func ioCopy(task *task.Task, add func(int64), dst io.Writer, src io.Reader) (int64, error) {
	var (
		rn      int   // read
		re      error // read error
		wn      int   // write
		we      error // write error
		written int64
		err     error
	)
	buf := make([]byte, 32*1024)
	for {
		// check task is canceled
		if task.Canceled() {
			return written, context.Canceled
		}
		// copy
		rn, re = src.Read(buf)
		if rn > 0 {
			wn, we = dst.Write(buf[:rn])
			if wn > 0 {
				val := int64(wn)
				written += val
				add(val)
			}
			if we != nil {
				err = we
				break
			}
			if rn != wn {
				err = io.ErrShortWrite
				break
			}
		}
		if re != nil {
			if re != io.EOF {
				err = re
			}
			break
		}
	}
	return written, err
}

func startTask(ctx context.Context, task *task.Task, name string) error {
	if done := ctx.Done(); done != nil {
		// check ctx is canceled before start
		select {
		case <-done:
			return ctx.Err()
		default:
		}
		// start a goroutine to watch ctx
		finish := make(chan struct{})
		defer close(finish)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					xpanic.Log(r, name+"WithContext")
				}
			}()
			select {
			case <-done:
				task.Cancel()
			case <-finish:
			}
		}()
	}
	err := task.Start()
	if err != nil {
		return err
	}
	// check progress
	progress := task.Progress()
	if progress != "100%" {
		return errors.New("unexpected progress: " + progress)
	}
	return nil
}
