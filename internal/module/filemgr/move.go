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

// moveTask implement task.Interface that is used to move source to destination.
// It can pause in progress and get current progress and detail information.
type moveTask struct {
	errCtrl  ErrCtrl
	dst      string
	paths    []string // absolute path that will be moved
	pathsLen int

	fastMode bool        // source and destination in the same volume
	dstStat  os.FileInfo // for record destination folder is created
	basePath string      // for filepath.Rel() in Process
	roots    []*file     // store all directories and files will move
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

// NewMoveTask is used to create a move task that implement task.Interface.
// Files must in the same directory.
func NewMoveTask(errCtrl ErrCtrl, callbacks fsm.Callbacks, dst string, paths ...string) *task.Task {
	mt := moveTask{
		errCtrl:    errCtrl,
		dst:        dst,
		paths:      paths,
		pathsLen:   len(paths),
		current:    big.NewFloat(0),
		total:      big.NewFloat(0),
		stopSignal: make(chan struct{}),
	}
	return task.New(TaskNameMove, &mt, callbacks)
}

// Prepare will check source and destination path.
func (mt *moveTask) Prepare(context.Context) error {
	// check paths
	if mt.pathsLen == 0 {
		return errors.New("empty path")
	}
	// check destination is not exist or a file.
	dstAbs, err := filepath.Abs(mt.dst)
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
	basePath, err := validatePaths(mt.paths)
	if err != nil {
		return err
	}
	// check can use fast mode
	vol := filepath.VolumeName(basePath)
	if vol != "" && vol == filepath.VolumeName(dstAbs) {
		mt.fastMode = true
	}
	mt.dst = dstAbs
	mt.dstStat = dstStat
	mt.basePath = basePath
	mt.roots = make([]*file, mt.pathsLen)
	go mt.watcher()
	return nil
}

func (mt *moveTask) Process(ctx context.Context, task *task.Task) error {
	// must collect files information because the destination maybe in the same path
	for i := 0; i < mt.pathsLen; i++ {
		err := mt.collectPathInfo(ctx, task, i)
		if err != nil {
			return err
		}
	}
	// create destination directory if it not exists
	if mt.dstStat == nil {
		err := os.MkdirAll(mt.dst, 0750)
		if err != nil {
			return err
		}
	}
	// move roots
	for i := 0; i < mt.pathsLen; i++ {
		err := mt.moveRoot(ctx, task, mt.roots[i])
		if err != nil {
			return err
		}
	}
	mt.updateDetail("finished")
	return nil
}

func (mt *moveTask) collectPathInfo(ctx context.Context, task *task.Task, i int) error {
	var (
		cDir  string // current directory
		cFile *file  // current file
	)
	srcPath := mt.paths[i]
	// for search dir faster, key is path
	dirs := make(map[string]*file, mt.pathsLen/4)
	walkFunc := func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			ps := noticePs{
				ctx:     ctx,
				task:    task,
				errCtrl: mt.errCtrl,
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
		f := &file{
			path: path,
			stat: stat,
		}
		isDir := stat.IsDir()
		// check is root path
		if mt.roots[i] == nil {
			mt.roots[i] = f
			// check root is file
			if !isDir {
				mt.addTotal(stat.Size())
				return nil
			}
			dirs[path] = f
			// set current data
			cDir = path
			cFile = f
			return nil
		}
		// update current directory and file
		dir := filepath.Dir(path)
		if dir != cDir {
			cDir = dir
			cFile = dirs[dir]
		}
		cFile.files = append(cFile.files, f)
		// update detail and total
		if isDir {
			cDir = path
			cFile = f
			dirs[path] = f
			// collect directory information
			// path: C:\testdata\test
			mt.updateDetail("collect directory information\npath: " + path)
			return nil
		}
		// collect file information
		// path: C:\testdata\test.dat
		mt.updateDetail("collect file information\npath: " + path)
		mt.addTotal(stat.Size())
		return nil
	}
	return filepath.Walk(srcPath, walkFunc)
}

func (mt *moveTask) moveRoot(ctx context.Context, task *task.Task, root *file) error {
	// if skip root directory, set fake progress for pass progress check
	if root == nil {
		mt.rwm.Lock()
		defer mt.rwm.Unlock()
		mt.current.Add(mt.current, oneFloat)
		mt.total.Add(mt.total, oneFloat)
		return nil
	}
	// check root is directory
	if !root.stat.IsDir() {
		_, err := mt.moveFile(ctx, task, root)
		if err != nil {
			return errors.WithMessage(err, "failed to move file")
		}
		return nil
	}
	// delete all directories and files in root directory
	_, err := mt.moveDir(ctx, task, root)
	if err != nil {
		return errors.WithMessage(err, "failed to move directory")
	}
	return nil
}

// returned bool is skipped this directory.
func (mt *moveTask) moveDir(ctx context.Context, task *task.Task, dir *file) (bool, error) {
	err := mt.mkdir(ctx, task, dir)
	if err != nil {
		return false, err
	}
	var skipped bool
	for _, file := range dir.files {
		// check task is canceled
		if task.Canceled() {
			return false, context.Canceled
		}
		var skip bool
		if file.stat.IsDir() {
			skip, err = mt.moveDir(ctx, task, file)
		} else {
			skip, err = mt.moveFile(ctx, task, file)
		}
		if err != nil {
			return false, err
		}
		if skip && !skipped {
			skipped = true
		}
	}
	// if move all files successfully, remove the directory
	if !skipped && !isRoot(dir.path) {
		return false, os.Remove(dir.path)
	}
	return true, nil
}

// mkdir is used to create destination directory if it is not exists.
func (mt *moveTask) mkdir(ctx context.Context, task *task.Task, dir *file) error {
	skip, relPath, err := mt.checkDstDir(dir)
	if skip || err != nil {
		return err
	}
	dstAbs := filepath.Join(mt.dst, relPath)
	// update current task detail, output:
	//   create directory, name: testdata
	//   src: C:\testdata
	//   dst: D:\testdata
	const format = "create directory, name: %s\nsrc: %s\ndst: %s"
	dirName := filepath.Base(dstAbs)
	mt.updateDetail(fmt.Sprintf(format, dirName, dir.path, dstAbs))
retry:
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	// check destination directory is exist
	dstStat, err := stat(dstAbs)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: mt.errCtrl,
		}
		stats := SrcDstStat{
			SrcAbs:  dir.path,
			DstAbs:  dstAbs,
			SrcStat: dir.stat,
			DstStat: nil,
		}
		retry, ne := noticeFailedToMove(&ps, &stats, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		mt.skipDirs = append(mt.skipDirs, dir.path)
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
			errCtrl: mt.errCtrl,
		}
		stats := SrcDstStat{
			SrcAbs:  dir.path,
			DstAbs:  dstAbs,
			SrcStat: dir.stat,
			DstStat: dstStat,
		}
		retry, ne := noticeSameDirFile(&ps, &stats)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		mt.skipDirs = append(mt.skipDirs, dir.path)
		return nil
	}
	// create directory
	err = os.Mkdir(dstAbs, dir.stat.Mode().Perm())
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: mt.errCtrl,
		}
		stats := SrcDstStat{
			SrcAbs:  dir.path,
			DstAbs:  dstAbs,
			SrcStat: dir.stat,
			DstStat: nil,
		}
		retry, ne := noticeFailedToMove(&ps, &stats, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		mt.skipDirs = append(mt.skipDirs, dir.path)
	}
	return nil
}

// checkDstDir is used to check destination directory is need to create.
func (mt *moveTask) checkDstDir(dir *file) (bool, string, error) {
	// skip dir if it in skipped directories
	for i := 0; i < len(mt.skipDirs); i++ {
		if strings.HasPrefix(dir.path, mt.skipDirs[i]) {
			return true, "", nil
		}
	}
	// can't recover
	relPath, err := filepath.Rel(mt.basePath, dir.path)
	if err != nil {
		return false, "", err
	}
	// is root directory
	if relPath == "." {
		return true, "", nil
	}
	return false, relPath, nil
}

// returned bool is skipped this file.
func (mt *moveTask) moveFile(ctx context.Context, task *task.Task, file *file) (bool, error) {
	// skip file if it in skipped directories
	for i := 0; i < len(mt.skipDirs); i++ {
		if strings.HasPrefix(file.path, mt.skipDirs[i]) {
			mt.updateCurrent(file.stat.Size(), true)
			return true, nil
		}
	}
	// can't recover
	relPath, err := filepath.Rel(mt.basePath, file.path)
	if err != nil {
		return false, err
	}
	dstAbs := filepath.Join(mt.dst, relPath)
	// update current task detail, output:
	//   move file, name: test.dat, size: 1.127 MB
	//   src: C:\testdata\test.dat
	//   dst: D:\testdata\test.dat
	const format = "move file, name: %s, size: %s\nsrc: %s\ndst: %s"
	fileName := filepath.Base(dstAbs)
	fileSize := convert.FormatByte(uint64(file.stat.Size()))
	mt.updateDetail(fmt.Sprintf(format, fileName, fileSize, file.path, dstAbs))
	// check destination file
	stats := &SrcDstStat{
		SrcAbs:  file.path,
		DstAbs:  dstAbs,
		SrcStat: file.stat,
		DstStat: nil,
	}
	skipped, err := mt.checkDstFile(ctx, task, stats)
	if err != nil {
		return false, err
	}
	if skipped {
		mt.updateCurrent(file.stat.Size(), true)
		return true, nil
	}
	// try to use fast mode first
	if mt.fastMode {
		return mt.moveFileFast(ctx, task, stats)
	}
	// create destination file
retry:
	// check task is canceled
	if task.Canceled() {
		return false, context.Canceled
	}
	perm := file.stat.Mode().Perm()
	dstFile, err := os.OpenFile(dstAbs, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm) // #nosec
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: mt.errCtrl,
		}
		retry, ne := noticeFailedToMove(&ps, stats, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return false, ne
		}
		mt.updateCurrent(file.stat.Size(), true)
		return true, nil
	}
	defer func() { _ = dstFile.Close() }()
	return mt.moveFileCommon(ctx, task, stats, dstFile)
}

// checkDstFile is used to check destination file is already exists.
func (mt *moveTask) checkDstFile(ctx context.Context, task *task.Task, stats *SrcDstStat) (bool, error) {
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
			errCtrl: mt.errCtrl,
		}
		retry, ne := noticeFailedToMove(&ps, stats, err)
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
		errCtrl: mt.errCtrl,
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

// moveFileFast will use os.Rename() to move file, it can move faster if two file in
// the same volume. Windows is finished, other platform need thinking.
func (mt *moveTask) moveFileFast(ctx context.Context, task *task.Task, stats *SrcDstStat) (bool, error) {
retry:
	// check task is canceled
	if task.Canceled() {
		return false, context.Canceled
	}
	err := os.Rename(stats.SrcAbs, stats.DstAbs)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: mt.errCtrl,
		}
		retry, ne := noticeFailedToMove(&ps, stats, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return false, ne
		}
		mt.updateCurrent(stats.SrcStat.Size(), true)
		return true, nil
	}
	mt.updateCurrent(stats.SrcStat.Size(), true)
	return false, nil
}

// moveFileCommon is used to use common mode to move file: first copy source file to
// destination, then delete source file.
func (mt *moveTask) moveFileCommon(
	ctx context.Context,
	task *task.Task,
	stats *SrcDstStat,
	dstFile *os.File,
) (skip bool, err error) {
	var copied int64
	defer func() {
		if err != nil && err != context.Canceled {
			ps := noticePs{
				ctx:     ctx,
				task:    task,
				errCtrl: mt.errCtrl,
			}
			var retry bool
			retry, err = noticeFailedToMove(&ps, stats, err)
			if retry {
				// reset current progress
				mt.updateCurrent(copied, false)
				skip, err = mt.retry(ctx, task, stats, dstFile)
				return
			}
			// if failed to move, delete destination file
			_ = dstFile.Close()
			_ = os.Remove(stats.DstAbs)
			// user cancel
			if err != nil {
				return
			}
			// skipped
			mt.updateCurrent(stats.SrcStat.Size()-copied, true)
		}
	}()
	srcFile, err := os.Open(stats.SrcAbs)
	if err != nil {
		return
	}
	var ok bool
	defer func() {
		_ = srcFile.Close()
		if ok {
			err = os.Remove(stats.SrcAbs)
		}
	}()
	err = mt.updateSrcFileStat(srcFile, stats)
	if err != nil {
		return
	}
	// prevent file become big
	srcSize := stats.SrcStat.Size()
	lr := io.LimitReader(srcFile, srcSize)
	copied, err = ioCopy(task, mt.addCurrent, dstFile, lr)
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
	err = os.Chtimes(stats.DstAbs, time.Now(), stats.SrcStat.ModTime())
	if err != nil {
		return
	}
	ok = true
	return
}

func (mt *moveTask) updateSrcFileStat(srcFile *os.File, stats *SrcDstStat) error {
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
		mt.addTotal(delta)
	}
	stats.SrcStat = srcStat
	return nil
}

func (mt *moveTask) addCurrent(delta int64) {
	mt.updateCurrent(delta, true)
}

func (mt *moveTask) retry(ctx context.Context, task *task.Task, stats *SrcDstStat, dst *os.File) (bool, error) {
	// check task is canceled
	if task.Canceled() {
		return false, context.Canceled
	}
	// reset offset about opened destination file
	_, err := dst.Seek(0, io.SeekStart)
	if err != nil {
		return false, err
	}
	return mt.moveFileCommon(ctx, task, stats, dst)
}

// Progress is used to get progress about current move task.
//
// collect: "0%"
// move:    "15.22%|current/total|128 MB/s"
// finish:  "100%"
func (mt *moveTask) Progress() string {
	mt.rwm.RLock()
	defer mt.rwm.RUnlock()
	// prevent / 0
	if mt.total.Cmp(zeroFloat) == 0 {
		return "0%"
	}
	switch mt.current.Cmp(mt.total) {
	case 0: // current == total
		return "100%"
	case 1: // current > total
		current := mt.current.Text('G', 64)
		total := mt.total.Text('G', 64)
		return fmt.Sprintf("error: current %s > total %s", current, total)
	}
	value := new(big.Float).Quo(mt.current, mt.total)
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
	current := mt.current.Text('G', 64)
	total := mt.total.Text('G', 64)
	speed := convert.FormatByte(mt.speed)
	return fmt.Sprintf("%s%%|%s/%s|%s/s", progress, current, total, speed)
}

func (mt *moveTask) updateCurrent(delta int64, add bool) {
	d := new(big.Float).SetInt64(delta)
	mt.rwm.Lock()
	defer mt.rwm.Unlock()
	if add {
		mt.current.Add(mt.current, d)
	} else {
		mt.current.Sub(mt.current, d)
	}
}

func (mt *moveTask) addTotal(delta int64) {
	d := new(big.Float).SetInt64(delta)
	mt.rwm.Lock()
	defer mt.rwm.Unlock()
	mt.total.Add(mt.total, d)
}

// Detail is used to get detail about move task.
//
// collect dir info:
//   collect directory information
//   path: C:\testdata\test
//
// move file:
//   move file, name: test.dat, size: 1.127 MB
//   src: C:\testdata\test.dat
//   dst: D:\test\test.dat
func (mt *moveTask) Detail() string {
	mt.rwm.RLock()
	defer mt.rwm.RUnlock()
	return mt.detail
}

func (mt *moveTask) updateDetail(detail string) {
	mt.rwm.Lock()
	defer mt.rwm.Unlock()
	mt.detail = detail
}

// watcher is used to calculate current move speed.
func (mt *moveTask) watcher() {
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "moveTask.watcher")
		}
	}()
	ticker := time.NewTicker(time.Second / time.Duration(len(mt.speeds)))
	defer ticker.Stop()
	current := new(big.Float)
	index := -1
	for {
		select {
		case <-ticker.C:
			index++
			if index >= len(mt.speeds) {
				index = 0
			}
			mt.watchSpeed(current, index)
		case <-mt.stopSignal:
			return
		}
	}
}

func (mt *moveTask) watchSpeed(current *big.Float, index int) {
	mt.rwm.Lock()
	defer mt.rwm.Unlock()
	delta := new(big.Float).Sub(mt.current, current)
	current.Add(current, delta)
	// update speed
	mt.speeds[index], _ = delta.Uint64()
	if mt.full {
		mt.speed = 0
		for i := 0; i < len(mt.speeds); i++ {
			mt.speed += mt.speeds[i]
		}
		return
	}
	if index == len(mt.speeds)-1 {
		mt.full = true
	}
	// calculate average speed
	var speed float64 // current speed
	for i := 0; i < index+1; i++ {
		speed += float64(mt.speeds[i])
	}
	mt.speed = uint64(speed / float64(index+1) * float64(len(mt.speeds)))
}

// Clean is used to send stop signal to watcher.
func (mt *moveTask) Clean() {
	close(mt.stopSignal)
}

// Move is used to create a move task to move paths to destination.
func Move(errCtrl ErrCtrl, dst string, paths ...string) error {
	return MoveWithContext(context.Background(), errCtrl, dst, paths...)
}

// MoveWithContext is used to create a move task with context to move paths to destination.
func MoveWithContext(ctx context.Context, errCtrl ErrCtrl, dst string, paths ...string) error {
	mt := NewMoveTask(errCtrl, nil, dst, paths...)
	return startTask(ctx, mt, "Move")
}
