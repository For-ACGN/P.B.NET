package filemgr

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/looplab/fsm"
	"github.com/pkg/errors"

	"project/internal/convert"
	"project/internal/task"
	"project/internal/xpanic"
)

// copyTask implement task.Interface that is used to copy source to destination.
// It can pause in progress and get current progress and detail information.
type copyTask struct {
	errCtrl  ErrCtrl
	dst      string
	paths    []string // absolute path that will be copied
	pathsLen int

	dstStat  os.FileInfo // for record destination folder is created
	basePath string      // for filepath.Rel() in Process
	files    []*fileStat // store all files and directories that will be copied
	skipDirs []string    // store skipped directories

	// about progress, detail and speed
	current *big.Float
	total   *big.Float
	detail  string
	speed   uint64
	speeds  [10]uint64
	full    bool
	rwm     sync.RWMutex

	// control speed watcher
	stopSignal chan struct{}
}

// NewCopyTask is used to create a copy task that implement task.Interface.
// Files must in the same directory.
func NewCopyTask(errCtrl ErrCtrl, callbacks fsm.Callbacks, dst string, paths ...string) *task.Task {
	ct := copyTask{
		errCtrl:    errCtrl,
		dst:        dst,
		paths:      paths,
		pathsLen:   len(paths),
		current:    big.NewFloat(0),
		total:      big.NewFloat(0),
		stopSignal: make(chan struct{}),
	}
	return task.New(TaskNameCopy, &ct, callbacks)
}

// Prepare will check source and destination path.
func (ct *copyTask) Prepare(context.Context) error {
	// check paths
	if ct.pathsLen == 0 {
		return errors.New("empty path")
	}
	// check destination is not exist or a file.
	dstAbs, err := filepath.Abs(ct.dst)
	if err != nil {
		return errors.Wrap(err, "failed to get absolute file path")
	}
	dstStat, err := stat(dstAbs)
	if err != nil {
		return err
	}
	if dstStat != nil && !dstStat.IsDir() {
		return errors.Errorf("destination path \"%s\" is a file", dstAbs)
	}
	// check paths is valid
	basePath, err := validatePaths(ct.paths)
	if err != nil {
		return err
	}
	ct.dst = dstAbs
	ct.dstStat = dstStat
	ct.basePath = basePath
	ct.files = make([]*fileStat, 0, ct.pathsLen*4)
	go ct.watcher()
	return nil
}

func (ct *copyTask) Process(ctx context.Context, task *task.Task) error {
	// must collect files information because the destination maybe in the same path
	for i := 0; i < ct.pathsLen; i++ {
		err := ct.collectPathInfo(ctx, task, ct.paths[i])
		if err != nil {
			return err
		}
	}
	// create destination directory if it not exists
	if ct.dstStat == nil {
		err := os.MkdirAll(ct.dst, 0750)
		if err != nil {
			return err
		}
	}
	filesLen := len(ct.files)
	if filesLen == 0 { // no files
		ct.rwm.Lock()
		defer ct.rwm.Unlock()
		ct.current.Add(ct.current, oneFloat)
		ct.total.Add(ct.total, oneFloat)
		// already locked, can't use ct.updateDetail
		ct.detail = "finished"
		return nil
	}
	for i := 0; i < filesLen; i++ {
		err := ct.copy(ctx, task, ct.files[i])
		if err != nil {
			return err
		}
	}
	ct.updateDetail("finished")
	return nil
}

func (ct *copyTask) collectPathInfo(ctx context.Context, task *task.Task, srcPath string) error {
	walkFunc := func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			ps := noticePs{
				ctx:     ctx,
				task:    task,
				errCtrl: ct.errCtrl,
			}
			const format = "failed to walk \"%s\" in \"%s\": %s"
			err = fmt.Errorf(format, path, srcPath, err)
			skip, ne := noticeFailedToCollect(&ps, path, err)
			if skip {
				return filepath.SkipDir
			}
			return ne
		}
		// check task is canceled
		if task.Canceled() {
			return context.Canceled
		}
		ct.files = append(ct.files, &fileStat{
			path: path,
			stat: stat,
		})
		// update detail and total size
		if stat.IsDir() {
			// collect directory information
			// path: C:\testdata\test
			ct.updateDetail("collect directory information\npath: " + path)
			return nil
		}
		// collect file information
		// path: C:\testdata\test.dat
		ct.updateDetail("collect file information\npath: " + path)
		ct.addTotal(stat.Size())
		return nil
	}
	return filepath.Walk(srcPath, walkFunc)
}

// copy is used to copy file or create directory.
func (ct *copyTask) copy(ctx context.Context, task *task.Task, file *fileStat) error {
	// skip if it in skipped directories
	for i := 0; i < len(ct.skipDirs); i++ {
		if strings.HasPrefix(file.path, ct.skipDirs[i]) {
			ct.updateCurrent(file.stat.Size(), true)
			return nil
		}
	}
	// can't recover
	relPath, err := filepath.Rel(ct.basePath, file.path)
	if err != nil {
		return err
	}
	// is root directory
	if relPath == "." {
		return nil
	}
	stats := &SrcDstStat{
		SrcAbs:  file.path,
		DstAbs:  filepath.Join(ct.dst, relPath),
		SrcStat: file.stat,
		DstStat: nil,
	}
	if file.stat.IsDir() {
		return ct.mkdir(ctx, task, stats)
	}
	return ct.copyFile(ctx, task, stats)
}

func (ct *copyTask) mkdir(ctx context.Context, task *task.Task, stats *SrcDstStat) error {
	// update current task detail, output:
	//   create directory, name: testdata
	//   src: C:\testdata
	//   dst: D:\testdata
	const format = "create directory, name: %s\nsrc: %s\ndst: %s"
	dirName := filepath.Base(stats.SrcAbs)
	ct.updateDetail(fmt.Sprintf(format, dirName, stats.SrcAbs, stats.DstAbs))
retry:
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	// check destination directory is exist
	dstStat, err := stat(stats.DstAbs)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ct.errCtrl,
		}
		retry, ne := noticeFailedToCopy(&ps, stats, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		ct.skipDirs = append(ct.skipDirs, stats.SrcAbs)
		return nil
	}
	// destination is already exists
	if dstStat != nil {
		if dstStat.IsDir() {
			return nil
		}
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ct.errCtrl,
		}
		stats.DstStat = dstStat
		retry, ne := noticeSameDirFile(&ps, stats)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		ct.skipDirs = append(ct.skipDirs, stats.SrcAbs)
		return nil
	}
	// create directory
	err = os.Mkdir(stats.DstAbs, stats.SrcStat.Mode().Perm())
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ct.errCtrl,
		}
		retry, ne := noticeFailedToCopy(&ps, stats, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		ct.skipDirs = append(ct.skipDirs, stats.SrcAbs)
	}
	return nil
}

func (ct *copyTask) copyFile(ctx context.Context, task *task.Task, stats *SrcDstStat) error {
	// update current task detail, output:
	//   copy file, name: test.dat, size: 1.127 MB
	//   src: C:\testdata\test.dat
	//   dst: D:\testdata\test.dat
	const format = "copy file, name: %s, size: %s\nsrc: %s\ndst: %s"
	fileName := filepath.Base(stats.SrcAbs)
	fileSize := convert.FormatByte(uint64(stats.SrcStat.Size()))
	ct.updateDetail(fmt.Sprintf(format, fileName, fileSize, stats.SrcAbs, stats.DstAbs))
	// check destination file
	skipped, err := ct.checkDstFile(ctx, task, stats)
	if err != nil {
		return err
	}
	if skipped {
		ct.updateCurrent(stats.SrcStat.Size(), true)
		return nil
	}
	// create destination file
retry:
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	perm := stats.SrcStat.Mode().Perm()
	dstFile, err := os.OpenFile(stats.DstAbs, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm) // #nosec
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ct.errCtrl,
		}
		retry, ne := noticeFailedToCopy(&ps, stats, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		ct.updateCurrent(stats.SrcStat.Size(), true)
		return nil
	}
	defer func() { _ = dstFile.Close() }()
	return ct.copySrcFile(ctx, task, stats, dstFile)
}

func (ct *copyTask) checkDstFile(ctx context.Context, task *task.Task, stats *SrcDstStat) (bool, error) {
retry:
	// check task is canceled
	if task.Canceled() {
		return false, context.Canceled
	}
	dstStat, err := stat(stats.DstAbs)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ct.errCtrl,
		}
		retry, ne := noticeFailedToCopy(&ps, stats, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return false, ne
		}
		return true, nil
	}
	// destination is not exist
	if dstStat == nil {
		return false, nil
	}
	stats.DstStat = dstStat
	ps := noticePs{
		ctx:     ctx,
		task:    task,
		errCtrl: ct.errCtrl,
	}
	if dstStat.IsDir() {
		retry, ne := noticeSameFileDir(&ps, stats)
		if retry {
			goto retry
		}
		if ne != nil {
			return false, ne
		}
		return true, nil
	}
	replace, ne := noticeSameFile(&ps, stats)
	if !replace {
		return true, ne
	}
	return false, nil
}

func (ct *copyTask) copySrcFile(
	ctx context.Context,
	task *task.Task,
	stats *SrcDstStat,
	dstFile *os.File,
) (err error) {
	var copied int64
	defer func() {
		if err != nil && err != context.Canceled {
			ps := noticePs{
				ctx:     ctx,
				task:    task,
				errCtrl: ct.errCtrl,
			}
			var retry bool
			retry, err = noticeFailedToCopy(&ps, stats, err)
			if retry {
				// reset current progress
				ct.updateCurrent(copied, false)
				err = ct.retry(ctx, task, stats, dstFile)
				return
			}
			// if failed to copy, delete destination file
			_ = dstFile.Close()
			_ = os.Remove(stats.DstAbs)
			// user cancel
			if err != nil {
				return
			}
			// skipped
			ct.updateCurrent(stats.SrcStat.Size()-copied, true)
		}
	}()
	srcFile, err := os.Open(stats.SrcAbs)
	if err != nil {
		return
	}
	defer func() { _ = srcFile.Close() }()
	err = ct.updateSrcFileStat(srcFile, stats)
	if err != nil {
		return
	}
	// prevent file become big
	srcSize := stats.SrcStat.Size()
	lr := io.LimitReader(srcFile, srcSize)
	copied, err = ioCopy(task, ct.addCurrent, dstFile, lr)
	if err != nil {
		return
	}
	// prevent file become small
	copied += srcSize - copied
	// prevent data lost
	err = dstFile.Sync()
	if err != nil {
		return
	}
	// set the modification time about the destination file
	return os.Chtimes(stats.DstAbs, time.Now(), stats.SrcStat.ModTime())
}

func (ct *copyTask) updateSrcFileStat(srcFile *os.File, stats *SrcDstStat) error {
	srcStat, err := srcFile.Stat()
	if err != nil {
		return err
	}
	// update total(must update in one operation)
	// total - old size + new size = total + (new size - old size)
	newSize := srcStat.Size()
	oldSize := stats.SrcStat.Size()
	if newSize != oldSize {
		delta := newSize - oldSize
		ct.addTotal(delta)
	}
	stats.SrcStat = srcStat
	return nil
}

func (ct *copyTask) addCurrent(delta int64) {
	ct.updateCurrent(delta, true)
}

func (ct *copyTask) retry(ctx context.Context, task *task.Task, stats *SrcDstStat, dstFile *os.File) error {
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	// reset offset about opened destination file
	_, err := dstFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return ct.copySrcFile(ctx, task, stats, dstFile)
}

// Progress is used to get progress about current copy task.
//
// collect: "0%"
// copy:    "15.22%|current/total|128 MB/s"
// finish:  "100%"
func (ct *copyTask) Progress() string {
	ct.rwm.RLock()
	defer ct.rwm.RUnlock()
	// prevent / 0
	if ct.total.Cmp(zeroFloat) == 0 {
		return "0%"
	}
	switch ct.current.Cmp(ct.total) {
	case 0: // current == total
		return "100%"
	case 1: // current > total
		current := ct.current.Text('G', 64)
		total := ct.total.Text('G', 64)
		return fmt.Sprintf("error: current %s > total %s", current, total)
	}
	value := new(big.Float).Quo(ct.current, ct.total)
	// split result
	text := value.Text('G', 64)
	// 0.999999999...999 -> 0.9999
	if len(text) > 6 {
		text = text[:6]
	}
	// format result
	result, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return "error: " + err.Error()
	}
	// 0.9999 -> 99.99%
	progress := strconv.FormatFloat(result*100, 'f', -1, 64)
	offset := strings.Index(progress, ".")
	if offset != -1 {
		if len(progress[offset+1:]) > 2 {
			progress = progress[:offset+3]
		}
	}
	// progress|current/total|speed
	current := ct.current.Text('G', 64)
	total := ct.total.Text('G', 64)
	speed := convert.FormatByte(ct.speed)
	return fmt.Sprintf("%s%%|%s/%s|%s/s", progress, current, total, speed)
}

func (ct *copyTask) updateCurrent(delta int64, add bool) {
	d := new(big.Float).SetInt64(delta)
	ct.rwm.Lock()
	defer ct.rwm.Unlock()
	if add {
		ct.current.Add(ct.current, d)
	} else {
		ct.current.Sub(ct.current, d)
	}
}

func (ct *copyTask) addTotal(delta int64) {
	d := new(big.Float).SetInt64(delta)
	ct.rwm.Lock()
	defer ct.rwm.Unlock()
	ct.total.Add(ct.total, d)
}

// Detail is used to get detail about copy task.
//
// collect dir info:
//   collect directory information
//   path: C:\testdata\test
//
// copy file:
//   copy file, name: test.dat, size: 1.127 MB
//   src: C:\testdata\test.dat
//   dst: D:\test\test.dat
func (ct *copyTask) Detail() string {
	ct.rwm.RLock()
	defer ct.rwm.RUnlock()
	return ct.detail
}

func (ct *copyTask) updateDetail(detail string) {
	ct.rwm.Lock()
	defer ct.rwm.Unlock()
	ct.detail = detail
}

// watcher is used to calculate current copy speed.
func (ct *copyTask) watcher() {
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "copyTask.watcher")
		}
	}()
	ticker := time.NewTicker(time.Second / time.Duration(len(ct.speeds)))
	defer ticker.Stop()
	current := new(big.Float)
	index := -1
	for {
		select {
		case <-ticker.C:
			index++
			if index >= len(ct.speeds) {
				index = 0
			}
			ct.watchSpeed(current, index)
		case <-ct.stopSignal:
			return
		}
	}
}

func (ct *copyTask) watchSpeed(current *big.Float, index int) {
	ct.rwm.Lock()
	defer ct.rwm.Unlock()
	delta := new(big.Float).Sub(ct.current, current)
	current.Add(current, delta)
	// update speed
	ct.speeds[index], _ = delta.Uint64()
	if ct.full {
		ct.speed = 0
		for i := 0; i < len(ct.speeds); i++ {
			ct.speed += ct.speeds[i]
		}
		return
	}
	if index == len(ct.speeds)-1 {
		ct.full = true
	}
	// calculate average speed
	var speed float64 // current speed
	for i := 0; i < index+1; i++ {
		speed += float64(ct.speeds[i])
	}
	ct.speed = uint64(speed / float64(index+1) * float64(len(ct.speeds)))
}

// Clean is used to send stop signal to watcher.
func (ct *copyTask) Clean() {
	close(ct.stopSignal)
}

// Copy is used to create a copy task to copy paths to destination.
func Copy(errCtrl ErrCtrl, dst string, paths ...string) error {
	return CopyWithContext(context.Background(), errCtrl, dst, paths...)
}

// CopyWithContext is used to create a copy task with context to copy paths to destination.
func CopyWithContext(ctx context.Context, errCtrl ErrCtrl, dst string, paths ...string) error {
	ct := NewCopyTask(errCtrl, nil, dst, paths...)
	return startTask(ctx, ct, "Copy")
}
