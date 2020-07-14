package filemgr

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"project/internal/module/task"
)

// name about task
const (
	TaskNameCopy       = "copy"
	TaskNameMove       = "move"
	TaskNameDelete     = "delete"
	TaskNameCompress   = "compress"
	TaskNameDecompress = "decompress"
)

// ErrCtrl is used to tell Move or Copy function how to control the same file,
// directory, or copy, move error. src and dst is the absolute file path.
// err and fileStat in stats maybe nil.
type ErrCtrl func(ctx context.Context, typ uint8, err error, stats *SrcDstStat) uint8

var (
	// ReplaceAll is used to replace all src file to dst file.
	ReplaceAll = func(context.Context, uint8, error, *SrcDstStat) uint8 { return ErrCtrlOpReplace }

	// SkipAll is used to skip all existed file or other error.
	SkipAll = func(context.Context, uint8, error, *SrcDstStat) uint8 { return ErrCtrlOpSkip }
)

// errors about ErrCtrl
const (
	_                    uint8 = iota
	ErrCtrlSameFile            // two same name file
	ErrCtrlSameFileDir         // same src file name with dst directory
	ErrCtrlSameDirFile         // same src directory name with dst file name
	ErrCtrlCollectFailed       // appear error in collectDirInfo()
	ErrCtrlCopyDirFailed       // appear error in copyDirFile()
	ErrCtrlCopyFailed          // appear error in copyFile()
)

// operation code about ErrCtrl
const (
	ErrCtrlOpInvalid uint8 = iota
	ErrCtrlOpReplace       // replace same name file
	ErrCtrlOpRetry         // try to copy or move again
	ErrCtrlOpSkip          // skip same name file, directory or copy
	ErrCtrlOpCancel        // cancel whole copy or move operation
)

var zeroFloat = big.NewFloat(0)

// stat is used to get file stat, if err is NotExist, it will return nil error and os.FileInfo.
func stat(name string) (os.FileInfo, error) {
	stat, err := os.Stat(name)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return stat, nil
}

// SrcDstStat contains absolute path and file stat about src and dst.
type SrcDstStat struct {
	SrcAbs  string // "E:\file.dat" "E:\file", last will not be "/ or "\"
	DstAbs  string
	SrcStat os.FileInfo
	DstStat os.FileInfo // check destination file or directory is exists

	SrcIsFile bool
}

// src path [file], dst path [file] --valid
// src path [file], dst path [dir]  --valid
// src path [dir],  dst path [dir]  --valid
// src path [dir],  dst path [file] --invalid
func checkSrcDstPath(src, dst string) (*SrcDstStat, error) {
	if src == "" {
		return nil, errors.New("empty src path")
	}
	if dst == "" {
		return nil, errors.New("empty dst path")
	}
	// replace the relative path to the absolute path for
	// prevent change current directory.
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return nil, err
	}
	dstAbs, err := filepath.Abs(dst)
	if err != nil {
		return nil, err
	}
	if srcAbs == dstAbs {
		return nil, errors.New("src path as same as the dst path")
	}
	// check two path is valid
	srcStat, err := os.Stat(srcAbs)
	if err != nil {
		return nil, err
	}
	dstStat, err := stat(dstAbs)
	if err != nil {
		return nil, err
	}
	srcIsDir := srcStat.IsDir()
	if srcIsDir && dstStat != nil && !dstStat.IsDir() {
		const format = "\"%s\" is a directory but \"%s\" is a file"
		return nil, fmt.Errorf(format, srcAbs, dstAbs)
	}
	return &SrcDstStat{
		SrcAbs:    srcAbs,
		DstAbs:    dstAbs,
		SrcStat:   srcStat,
		DstStat:   dstStat,
		SrcIsFile: !srcIsDir,
	}, nil
}

type fileStat struct {
	path string // abs
	stat os.FileInfo
}

// ErrUserCanceled is an error about user cancel copy or move.
var ErrUserCanceled = fmt.Errorf("user canceled")

// noticeSameFile is used to notice appear same name file.
func noticeSameFile(
	ctx context.Context,
	task *task.Task,
	errCtrl ErrCtrl,
	stats *SrcDstStat,
) (replace bool, err error) {
	task.Pause()
	defer task.Continue()
	switch code := errCtrl(ctx, ErrCtrlSameFile, nil, stats); code {
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
func noticeSameFileDir(
	ctx context.Context,
	task *task.Task,
	errCtrl ErrCtrl,
	stats *SrcDstStat,
) (retry bool, err error) {
	task.Pause()
	defer task.Continue()
	switch code := errCtrl(ctx, ErrCtrlSameFileDir, nil, stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip, ErrCtrlOpReplace: // for ReplaceAll
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown same file dir operation code: %d", code)
	}
	return
}

// noticeSameDirFile is used to notice appear same name about src dir and dst file.
func noticeSameDirFile(
	ctx context.Context,
	task *task.Task,
	errCtrl ErrCtrl,
	stats *SrcDstStat,
) (retry bool, err error) {
	task.Pause()
	defer task.Continue()
	switch code := errCtrl(ctx, ErrCtrlSameDirFile, nil, stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip, ErrCtrlOpReplace: // for ReplaceAll
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown same dir file operation code: %d", code)
	}
	return
}

// noticeFailedToCollect is used to notice appear some error in collectDirInfo.
func noticeFailedToCollect(
	ctx context.Context,
	task *task.Task,
	errCtrl ErrCtrl,
	path string,
	extError error,
) (retry bool, err error) {
	task.Pause()
	defer task.Continue()
	stats := SrcDstStat{SrcAbs: path}
	switch code := errCtrl(ctx, ErrCtrlCollectFailed, extError, &stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip, ErrCtrlOpReplace: // for ReplaceAll
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to collect operation code: %d", code)
	}
	return
}

// noticeFailedToCopy is used to notice appear some error about copy or move.
func noticeFailedToCopy(
	ctx context.Context,
	task *task.Task,
	errCtrl ErrCtrl,
	stats *SrcDstStat,
	extError error,
) (retry bool, err error) {
	task.Pause()
	defer task.Continue()
	switch code := errCtrl(ctx, ErrCtrlCopyFailed, extError, stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip, ErrCtrlOpReplace: // for ReplaceAll
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to copy operation code: %d", code)
	}
	return
}

// noticeFailedToCopyDir is used to notice appear some error about copyDirFile.
func noticeFailedToCopyDir(
	ctx context.Context,
	task *task.Task,
	errCtrl ErrCtrl,
	path string,
	extError error,
) (retry bool, err error) {
	task.Pause()
	defer task.Continue()
	stats := SrcDstStat{SrcAbs: path}
	switch code := errCtrl(ctx, ErrCtrlCopyDirFailed, extError, &stats); code {
	case ErrCtrlOpRetry:
		retry = true
	case ErrCtrlOpSkip, ErrCtrlOpReplace: // for ReplaceAll
	case ErrCtrlOpCancel:
		err = ErrUserCanceled
	default:
		err = errors.Errorf("unknown failed to copy operation code: %d", code)
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