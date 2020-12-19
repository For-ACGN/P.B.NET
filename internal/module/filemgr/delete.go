package filemgr

import (
	"context"
	"fmt"
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
	"project/internal/logger"
	"project/internal/task"
	"project/internal/xpanic"
)

// deleteTask is implement task.Interface that is used to delete file or files in one
// directory. It can pause in progress and get current progress and detail information.
type deleteTask struct {
	errCtrl  ErrCtrl
	paths    []string // absolute path that will be deleted
	pathsLen int

	roots []*file // store all directories and files will delete

	// about progress, speed and detail
	current *big.Float
	total   *big.Float
	speed   uint64
	speeds  [10]uint64
	full    bool
	detail  string
	rwm     sync.RWMutex

	// control speed watcher
	stopSignal chan struct{}
}

// NewDeleteTask is used to create a delete task that implement task.Interface.
func NewDeleteTask(errCtrl ErrCtrl, callbacks fsm.Callbacks, paths ...string) *task.Task {
	dt := deleteTask{
		errCtrl:    errCtrl,
		paths:      paths,
		pathsLen:   len(paths),
		current:    big.NewFloat(0),
		total:      big.NewFloat(0),
		stopSignal: make(chan struct{}),
	}
	return task.New(TaskNameDelete, &dt, callbacks)
}

// Prepare is used to check directory about source paths is same.
func (dt *deleteTask) Prepare(context.Context) error {
	// check paths
	if dt.pathsLen == 0 {
		return errors.New("empty path")
	}
	// check paths is valid
	_, err := validatePaths(dt.paths)
	if err != nil {
		return err
	}
	dt.roots = make([]*file, dt.pathsLen)
	go dt.watcher()
	return nil
}

func (dt *deleteTask) Process(ctx context.Context, task *task.Task) error {
	for i := 0; i < dt.pathsLen; i++ {
		err := dt.collectPathInfo(ctx, task, i)
		if err != nil {
			return err
		}
	}
	for i := 0; i < dt.pathsLen; i++ {
		err := dt.deleteRoot(ctx, task, dt.roots[i])
		if err != nil {
			return err
		}
	}
	dt.updateDetail("finished")
	return nil
}

func (dt *deleteTask) collectPathInfo(ctx context.Context, task *task.Task, i int) error {
	var (
		cDir  string // current directory
		cFile *file  // current file
	)
	srcPath := dt.paths[i]
	// for search dir faster, key is path
	dirs := make(map[string]*file, dt.pathsLen/4)
	walkFunc := func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			ps := noticePs{
				ctx:     ctx,
				task:    task,
				errCtrl: dt.errCtrl,
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
		if dt.roots[i] == nil {
			dt.roots[i] = f
			// check root is file
			if !isDir {
				dt.addTotal()
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
			dt.updateDetail("collect directory information\npath: " + path)
			return nil
		}
		// collect file information
		// path: C:\testdata\test.dat
		dt.updateDetail("collect file information\npath: " + path)
		dt.addTotal()
		return nil
	}
	return filepath.Walk(srcPath, walkFunc)
}

func (dt *deleteTask) deleteRoot(ctx context.Context, task *task.Task, root *file) error {
	// if skip root directory, set fake progress for pass progress check
	if root == nil {
		dt.rwm.Lock()
		defer dt.rwm.Unlock()
		dt.current.Add(dt.current, oneFloat)
		dt.total.Add(dt.total, oneFloat)
		return nil
	}
	// check root is directory
	if !root.stat.IsDir() {
		_, err := dt.deleteFile(ctx, task, root)
		if err != nil {
			return errors.WithMessage(err, "failed to delete file")
		}
		return nil
	}
	// delete all directories and files in root directory
	_, err := dt.deleteDir(ctx, task, root)
	if err != nil {
		return errors.WithMessage(err, "failed to delete directory")
	}
	return nil
}

// returned bool is skipped this directory.
func (dt *deleteTask) deleteDir(ctx context.Context, task *task.Task, dir *file) (bool, error) {
	var (
		skipped bool
		err     error
	)
	for _, file := range dir.files {
		// check task is canceled
		if task.Canceled() {
			return false, context.Canceled
		}
		var skip bool
		if file.stat.IsDir() {
			skip, err = dt.deleteDir(ctx, task, file)
		} else {
			skip, err = dt.deleteFile(ctx, task, file)
		}
		if err != nil {
			return false, err
		}
		if skip && !skipped {
			skipped = true
		}
	}
	// if delete all files successfully, delete the directory
	if !skipped && !isRoot(dir.path) {
		return false, os.Remove(dir.path)
	}
	return true, nil
}

// returned bool is skipped this file.
func (dt *deleteTask) deleteFile(ctx context.Context, task *task.Task, file *file) (bool, error) {
	// update current task detail, output:
	//   delete file, name: test.dat
	//   path: C:\testdata\test.dat
	//   modify time: 2020-07-27 19:12:17
	const format = "delete file, name: %s\npath: %s\nmodify time: %s"
	fileName := filepath.Base(file.path)
	modTime := file.stat.ModTime().Format(logger.TimeLayout)
	dt.updateDetail(fmt.Sprintf(format, fileName, file.path, modTime))
retry:
	// check task is canceled
	if task.Canceled() {
		return false, context.Canceled
	}
	// delete file
	err := os.Remove(file.path)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: dt.errCtrl,
		}
		retry, ne := noticeFailedToDelete(&ps, file.path, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return false, ne
		}
		dt.addCurrent()
		return true, nil
	}
	dt.addCurrent()
	return false, nil
}

// Progress is used to get progress about current delete task.
//
// collect: "0%"
// delete:  "15.22%|current/total|4,523 file/s"
// finish:  "100%"
func (dt *deleteTask) Progress() string {
	dt.rwm.RLock()
	defer dt.rwm.RUnlock()
	// prevent / 0
	if dt.total.Cmp(zeroFloat) == 0 {
		return "0%"
	}
	switch dt.current.Cmp(dt.total) {
	case 0: // current == total
		return "100%"
	case 1: // current > total
		current := dt.current.Text('G', 64)
		total := dt.total.Text('G', 64)
		return fmt.Sprintf("error: current %s > total %s", current, total)
	}
	value := new(big.Float).Quo(dt.current, dt.total)
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
	current := dt.current.Text('G', 64)
	total := dt.total.Text('G', 64)
	speed := convert.FormatNumber(strconv.FormatUint(dt.speed, 10))
	return fmt.Sprintf("%s%%|%s/%s|%s file/s", progress, current, total, speed)
}

func (dt *deleteTask) addCurrent() {
	dt.rwm.Lock()
	defer dt.rwm.Unlock()
	dt.current.Add(dt.current, oneFloat)
}

func (dt *deleteTask) addTotal() {
	dt.rwm.Lock()
	defer dt.rwm.Unlock()
	dt.total.Add(dt.total, oneFloat)
}

// Detail is used to get detail about delete task.
//
// collect dir info:
//   collect directory information
//   path: C:\testdata\test
//
// delete file:
//   delete file, name: test.dat
//   path: C:\testdata\test.dat
//   modify time: 2020-07-27 19:12:17
func (dt *deleteTask) Detail() string {
	dt.rwm.RLock()
	defer dt.rwm.RUnlock()
	return dt.detail
}

func (dt *deleteTask) updateDetail(detail string) {
	dt.rwm.Lock()
	defer dt.rwm.Unlock()
	dt.detail = detail
}

// watcher is used to calculate current delete file speed.
func (dt *deleteTask) watcher() {
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "deleteTask.watcher")
		}
	}()
	ticker := time.NewTicker(time.Second / time.Duration(len(dt.speeds)))
	defer ticker.Stop()
	current := new(big.Float)
	index := -1
	for {
		select {
		case <-ticker.C:
			index++
			if index >= len(dt.speeds) {
				index = 0
			}
			dt.watchSpeed(current, index)
		case <-dt.stopSignal:
			return
		}
	}
}

func (dt *deleteTask) watchSpeed(current *big.Float, index int) {
	dt.rwm.Lock()
	defer dt.rwm.Unlock()
	delta := new(big.Float).Sub(dt.current, current)
	current.Add(current, delta)
	// update speed
	dt.speeds[index], _ = delta.Uint64()
	if dt.full {
		dt.speed = 0
		for i := 0; i < len(dt.speeds); i++ {
			dt.speed += dt.speeds[i]
		}
		return
	}
	if index == len(dt.speeds)-1 {
		dt.full = true
	}
	// calculate average speed
	var speed uint64 // current speed
	for i := 0; i < index+1; i++ {
		speed += dt.speeds[i]
	}
	dt.speed = speed / uint64(index+1) * uint64(len(dt.speeds))
}

// Clean is used to send stop signal to watcher.
func (dt *deleteTask) Clean() {
	close(dt.stopSignal)
}

// Delete is used to create a delete task to delete paths.
func Delete(errCtrl ErrCtrl, paths ...string) error {
	return DeleteWithContext(context.Background(), errCtrl, paths...)
}

// DeleteWithContext is used to create a delete task with context to delete paths.
func DeleteWithContext(ctx context.Context, errCtrl ErrCtrl, paths ...string) error {
	dt := NewDeleteTask(errCtrl, nil, paths...)
	return startTask(ctx, dt, "Delete")
}
