package filemgr

import (
	"archive/zip"
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
	"project/internal/system"
	"project/internal/task"
	"project/internal/xpanic"
)

// zipTask implement task.Interface that is used to compress files into a zip file.
// It can pause in progress and get current progress and detail information.
type zipTask struct {
	errCtrl  ErrCtrl
	zipPath  string   // zip file absolute path that be created
	paths    []string // absolute path that will be compressed
	pathsLen int

	basePath  string      // for filepath.Rel() in Process
	files     []*fileStat // store all files stats that will be compressed
	skipDirs  []string    // store skipped directories
	zipWriter *zip.Writer

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

// NewZipTask is used to create a zip task that implement task.Interface.
// If files is nil, it will create a zip file with empty file.
// Files must in the same directory.
func NewZipTask(errCtrl ErrCtrl, callbacks fsm.Callbacks, zipPath string, paths ...string) *task.Task {
	zt := zipTask{
		errCtrl:    errCtrl,
		zipPath:    zipPath,
		paths:      paths,
		pathsLen:   len(paths),
		current:    big.NewFloat(0),
		total:      big.NewFloat(0),
		stopSignal: make(chan struct{}),
	}
	return task.New(TaskNameZip, &zt, callbacks)
}

// Prepare is used to check destination file path is not exist.
func (zt *zipTask) Prepare(context.Context) error {
	// check paths
	if zt.pathsLen == 0 {
		return errors.New("empty path")
	}
	// check destination path is not exist
	dstAbs, err := filepath.Abs(zt.zipPath)
	if err != nil {
		return errors.Wrap(err, "failed to get absolute file path")
	}
	dstStat, err := stat(dstAbs)
	if err != nil {
		return err
	}
	if dstStat != nil {
		return errors.Errorf("destination path %s is already exists", dstAbs)
	}
	// check paths is valid
	zt.basePath, err = validatePaths(zt.paths)
	if err != nil {
		return err
	}
	zt.zipPath = dstAbs
	zt.files = make([]*fileStat, 0, zt.pathsLen*4)
	go zt.watcher()
	return nil
}

func (zt *zipTask) Process(ctx context.Context, task *task.Task) error {
	// must collect files information because the zip file maybe in the same path
	for i := 0; i < zt.pathsLen; i++ {
		err := zt.collectPathInfo(ctx, task, zt.paths[i])
		if err != nil {
			return err
		}
	}
	// create zip file
	zipFile, err := system.OpenFile(zt.zipPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to create zip file")
	}
	defer func() { _ = zipFile.Close() }()
	// check files is empty
	l := len(zt.files)
	if l == 0 {
		// set fake progress for pass progress check
		zt.rwm.Lock()
		defer zt.rwm.Unlock()
		zt.current.SetUint64(1)
		zt.total.SetUint64(1)
		return nil
	}
	// compress files and add directories
	zt.zipWriter = zip.NewWriter(zipFile)
	for i := 0; i < l; i++ {
		err := zt.compress(ctx, task, zt.files[i])
		if err != nil {
			_ = zt.zipWriter.Close()
			return err
		}
	}
	// make final zip file
	err = zt.zipWriter.Close()
	if err != nil {
		return err
	}
	err = zipFile.Sync()
	if err != nil {
		return err
	}
	zt.updateDetail("finished")
	return nil
}

func (zt *zipTask) collectPathInfo(ctx context.Context, task *task.Task, srcPath string) error {
	walkFunc := func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			ps := noticePs{
				ctx:     ctx,
				task:    task,
				errCtrl: zt.errCtrl,
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
		zt.files = append(zt.files, &fileStat{
			path: path,
			stat: stat,
		})
		// update detail and total size
		if stat.IsDir() {
			// collect directory information
			// path: C:\testdata\test
			zt.updateDetail("collect directory information\npath: " + path)
			return nil
		}
		// collect file information
		// path: C:\testdata\test.dat
		zt.updateDetail("collect file information\npath: " + path)
		zt.addTotal(stat.Size())
		return nil
	}
	return filepath.Walk(srcPath, walkFunc)
}

func (zt *zipTask) compress(ctx context.Context, task *task.Task, file *fileStat) error {
	// skip file if it in skipped directories
	for i := 0; i < len(zt.skipDirs); i++ {
		if strings.HasPrefix(file.path, zt.skipDirs[i]) {
			zt.addCurrent(file.stat.Size())
			return nil
		}
	}
	// can't recover or file is root directory
	relPath, err := filepath.Rel(zt.basePath, file.path)
	if err != nil {
		return err
	}
	// is root directory
	if relPath == "." {
		return nil
	}
	relPath = strings.ReplaceAll(relPath, "\\", "/")
	if file.stat.IsDir() {
		return zt.mkdir(ctx, task, file.path, relPath)
	}
	return zt.writeFile(ctx, task, file, relPath)
}

func (zt *zipTask) mkdir(ctx context.Context, task *task.Task, dirPath, relPath string) error {
	// update current task detail, output:
	//   create directory, name: testdata
	//   src: C:\testdata
	//   dst: zip/testdata
	const format = "create directory, name: %s\nsrc: %s\ndst: zip/%s"
	dirName := filepath.Base(dirPath)
	zt.updateDetail(fmt.Sprintf(format, dirName, dirPath, relPath))
retry:
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	// update modification time
	dirStat, err := os.Stat(dirPath)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: zt.errCtrl,
		}
		retry, ne := noticeFailedToZip(&ps, dirPath, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		zt.skipDirs = append(zt.skipDirs, dirPath)
		return nil
	}
	// create a directory
	header := zip.FileHeader{
		Name:     relPath + "/",
		Method:   zip.Store,
		Modified: dirStat.ModTime().UTC(), // Convert to UTC for compatibility
	}
	header.SetMode(dirStat.Mode())
	_, err = zt.zipWriter.CreateHeader(&header)
	return err
}

func (zt *zipTask) writeFile(ctx context.Context, task *task.Task, file *fileStat, relPath string) error {
	// update current task detail, output:
	//   compress file, name: test.dat, size: 1.127 MB
	//   src: C:\testdata\test.dat
	//   dst: zip/testdata/test.dat
	const format = "compress file, name: %s, size: %s\nsrc: %s\ndst: zip/%s"
	fileName := filepath.Base(file.path)
	fileSize := convert.FormatByte(uint64(file.stat.Size()))
	zt.updateDetail(fmt.Sprintf(format, fileName, fileSize, file.path, relPath))
retry:
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	// open source file
	srcFile, err := os.Open(file.path)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: zt.errCtrl,
		}
		retry, ne := noticeFailedToZip(&ps, file.path, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		zt.addCurrent(file.stat.Size())
		return nil
	}
	defer func() { _ = srcFile.Close() }()
	srcStat, err := srcFile.Stat()
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: zt.errCtrl,
		}
		retry, ne := noticeFailedToZip(&ps, file.path, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		zt.addCurrent(file.stat.Size())
		return nil
	}
	// create a file
	header := zip.FileHeader{
		Name:               relPath,
		Method:             zip.Deflate,
		Modified:           srcStat.ModTime().UTC(), // Convert to UTC for compatibility
		UncompressedSize64: uint64(srcStat.Size()),
	}
	header.SetMode(srcStat.Mode())
	zipFile, err := zt.zipWriter.CreateHeader(&header)
	if err != nil {
		return err
	}
	// compress file
	var copied int64
	size := srcStat.Size()
	defer func() {
		// check size after copy
		if size == copied {
			return
		}
		zt.addCurrent(size - copied)
	}()
	lr := io.LimitReader(srcFile, size)
	copied, err = ioCopy(task, zt.addCurrent, zipFile, lr)
	return err
}

// Progress is used to get progress about current zip task.
//
// collect: "0%"
// zip:     "15.22%|current/total|128 MB/s"
// finish:  "100%"
func (zt *zipTask) Progress() string {
	zt.rwm.RLock()
	defer zt.rwm.RUnlock()
	// prevent / 0
	if zt.total.Cmp(zeroFloat) == 0 {
		return "0%"
	}
	switch zt.current.Cmp(zt.total) {
	case 0: // current == total
		return "100%"
	case 1: // current > total
		current := zt.current.Text('G', 64)
		total := zt.total.Text('G', 64)
		return fmt.Sprintf("error: current %s > total %s", current, total)
	}
	value := new(big.Float).Quo(zt.current, zt.total)
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
	current := zt.current.Text('G', 64)
	total := zt.total.Text('G', 64)
	speed := convert.FormatByte(zt.speed)
	return fmt.Sprintf("%s%%|%s/%s|%s/s", progress, current, total, speed)
}

func (zt *zipTask) addCurrent(delta int64) {
	d := new(big.Float).SetInt64(delta)
	zt.rwm.Lock()
	defer zt.rwm.Unlock()
	zt.current.Add(zt.current, d)
}

func (zt *zipTask) addTotal(delta int64) {
	d := new(big.Float).SetInt64(delta)
	zt.rwm.Lock()
	defer zt.rwm.Unlock()
	zt.total.Add(zt.total, d)
}

// Detail is used to get detail about zip task.
// collect file or directory info:
//   collect file information
//   path: C:\testdata\test.dat
//
// compress file:
//   compress file, name: test.dat, size: 1.127 MB
//   src: C:\testdata\test.dat
//   dst: zip/testdata/test.dat
func (zt *zipTask) Detail() string {
	zt.rwm.RLock()
	defer zt.rwm.RUnlock()
	return zt.detail
}

func (zt *zipTask) updateDetail(detail string) {
	zt.rwm.Lock()
	defer zt.rwm.Unlock()
	zt.detail = detail
}

// watcher is used to calculate current compress speed.
func (zt *zipTask) watcher() {
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "zipTask.watcher")
		}
	}()
	ticker := time.NewTicker(time.Second / time.Duration(len(zt.speeds)))
	defer ticker.Stop()
	current := new(big.Float)
	index := -1
	for {
		select {
		case <-ticker.C:
			index++
			if index >= len(zt.speeds) {
				index = 0
			}
			zt.watchSpeed(current, index)
		case <-zt.stopSignal:
			return
		}
	}
}

func (zt *zipTask) watchSpeed(current *big.Float, index int) {
	zt.rwm.Lock()
	defer zt.rwm.Unlock()
	delta := new(big.Float).Sub(zt.current, current)
	current.Add(current, delta)
	// update speed
	zt.speeds[index], _ = delta.Uint64()
	if zt.full {
		zt.speed = 0
		for i := 0; i < len(zt.speeds); i++ {
			zt.speed += zt.speeds[i]
		}
		return
	}
	if index == len(zt.speeds)-1 {
		zt.full = true
	}
	// calculate average speed
	var speed float64 // current speed
	for i := 0; i < index+1; i++ {
		speed += float64(zt.speeds[i])
	}
	zt.speed = uint64(speed / float64(index+1) * float64(len(zt.speeds)))
}

// Clean is used to send stop signal to watcher.
func (zt *zipTask) Clean() {
	close(zt.stopSignal)
}

// Zip is used to create a zip task to compress files into a zip file.
func Zip(errCtrl ErrCtrl, zipPath string, paths ...string) error {
	return ZipWithContext(context.Background(), errCtrl, zipPath, paths...)
}

// ZipWithContext is used to create a zip task with context to compress files into a zip file.
func ZipWithContext(ctx context.Context, errCtrl ErrCtrl, zipPath string, paths ...string) error {
	zt := NewZipTask(errCtrl, nil, zipPath, paths...)
	return startTask(ctx, zt, "Zip")
}
