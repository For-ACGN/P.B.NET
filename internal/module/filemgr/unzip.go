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

// unZipTask implement task.Interface that is used to extract files from a zip file.
// It can pause in progress and get current progress and detail information.
type unZipTask struct {
	errCtrl  ErrCtrl
	src      string   // zip file absolute path
	dst      string   // destination path to store extract files
	paths    []string // files need extract
	pathsLen int

	dstStat  os.FileInfo     // for record destination folder is created
	zipFile  *zip.ReadCloser // zip reader and *os.File
	dirs     []*zip.File     // set modified time after extract files
	skipDirs []string        // store skipped directories

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

// NewUnZipTask is used to create a unzip task that implement task.Interface.
// If files is nil, extract all files from source zip file.
func NewUnZipTask(errCtrl ErrCtrl, callbacks fsm.Callbacks, src, dst string, path ...string) *task.Task {
	ut := unZipTask{
		errCtrl:    errCtrl,
		src:        src,
		dst:        dst,
		paths:      path,
		pathsLen:   len(path),
		current:    big.NewFloat(0),
		total:      big.NewFloat(0),
		stopSignal: make(chan struct{}),
	}
	return task.New(TaskNameUnZip, &ut, callbacks)
}

// Prepare is used to check destination is not exist or a file.
func (ut *unZipTask) Prepare(context.Context) error {
	// replace the relative path to the absolute path for
	// prevent program change current directory.
	srcAbs, err := filepath.Abs(ut.src)
	if err != nil {
		return err
	}
	dstAbs, err := filepath.Abs(ut.dst)
	if err != nil {
		return err
	}
	// check two path is valid
	srcStat, err := os.Stat(srcAbs)
	if err != nil {
		return err
	}
	dstStat, err := stat(dstAbs)
	if err != nil {
		return err
	}
	if srcStat.IsDir() {
		return errors.Errorf("source path \"%s\" is a directory", srcAbs)
	}
	if dstStat != nil && !dstStat.IsDir() {
		return errors.Errorf("destination path \"%s\" is a file", dstAbs)
	}
	ut.src = srcAbs
	ut.dst = dstAbs
	ut.dstStat = dstStat
	go ut.watcher()
	return nil
}

func (ut *unZipTask) Process(ctx context.Context, task *task.Task) error {
	// create destination directory if it not exists
	if ut.dstStat == nil {
		err := os.MkdirAll(ut.dst, 0750)
		if err != nil {
			return err
		}
	}
	// open and read zip file
	ut.updateDetail("read zip file")
	zipFile, err := zip.OpenReader(ut.src)
	if err != nil {
		return err
	}
	defer func() { _ = zipFile.Close() }()
	// extract files
	ut.zipFile = zipFile
	if ut.pathsLen == 0 {
		err = ut.extractAll(ctx, task)
	} else {
		err = ut.extractPart(ctx, task)
	}
	if err != nil {
		return err
	}
	// set extracted directory modification time again
next:
	for _, dir := range ut.dirs {
		// check task is canceled
		if task.Canceled() {
			return context.Canceled
		}
		dirPath := filepath.Clean(dir.Name)
		// skip file if it in skipped directories
		for i := 0; i < len(ut.skipDirs); i++ {
			path := strings.ReplaceAll(dirPath, "\\", "/")
			if strings.HasPrefix(path, ut.skipDirs[i]) {
				continue next
			}
		}
		err = os.Chtimes(filepath.Join(ut.dst, dirPath), time.Now(), dir.Modified)
		if err != nil {
			return errors.Wrap(err, "failed to change directory modification time")
		}
	}
	ut.updateDetail("finished")
	return nil
}

func (ut *unZipTask) extractAll(ctx context.Context, task *task.Task) error {
	sort.Sort(zipFiles(ut.zipFile.File)) // sort zip files
	return ut.extractZipFiles(ctx, task, ut.zipFile.File)
}

func (ut *unZipTask) extractPart(ctx context.Context, task *task.Task) error {
	// check file is in zip
	filesMap := make(map[string]*zip.File)
	for _, file := range ut.zipFile.File {
		filesMap[filepath.Clean(file.Name)] = file
	}
	// prevent add the same file or directory
	extFiles := make(map[string]struct{})
	// prevent add a directory or a file is sub file in directory
	dirs := make(map[string]struct{})
	// add files
	sort.Strings(ut.paths)
	files := make([]*zip.File, 0, ut.pathsLen)
next:
	for _, path := range ut.paths {
		cPath := filepath.Clean(path)
		if _, ok := extFiles[cPath]; ok {
			return errors.Errorf("appear the same path \"%s\"", cPath)
		}
		// check is added(already add dir that include this file)
		for dir := range dirs {
			if strings.HasPrefix(cPath, dir) {
				continue next
			}
		}
		if file, ok := filesMap[cPath]; ok {
			if file.FileInfo().IsDir() {
				// add all sub files in this directory
				for path, file := range filesMap {
					if strings.HasPrefix(path, cPath) {
						files = append(files, file)
					}
				}
				dirs[cPath] = struct{}{}
			}
			files = append(files, file)
			extFiles[cPath] = struct{}{}
		} else {
			return errors.Errorf("\"%s\" doesn't exist in zip file", path)
		}
	}
	sort.Sort(zipFiles(files)) // sort zip files
	return ut.extractZipFiles(ctx, task, files)
}

func (ut *unZipTask) extractZipFiles(ctx context.Context, task *task.Task, files []*zip.File) error {
	err := ut.collectFilesInfo(task, files)
	if err != nil {
		return err
	}
	for i := 0; i < len(files); i++ {
		err = ut.extractZipFile(ctx, task, files[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ut *unZipTask) collectFilesInfo(task *task.Task, files []*zip.File) error {
	for _, file := range files {
		// check task is canceled
		if task.Canceled() {
			return context.Canceled
		}
		path := strings.ReplaceAll(file.Name, "\\", "/")
		fi := file.FileInfo()
		if fi.IsDir() {
			ut.dirs = append(ut.dirs, file)
			// collect directory information
			// path: testdata/test
			ut.updateDetail("collect directory information\npath: " + path)
			continue
		}
		// collect file information
		// path: testdata/test
		ut.updateDetail("collect file information\npath: " + path)
		ut.addTotal(fi.Size())
	}
	return nil
}

func (ut *unZipTask) extractZipFile(ctx context.Context, task *task.Task, file *zip.File) error {
	path := strings.ReplaceAll(filepath.Clean(file.Name), "\\", "/")
	fi := file.FileInfo()
	// skip file if it in skipped directories
	for i := 0; i < len(ut.skipDirs); i++ {
		if strings.HasPrefix(path, ut.skipDirs[i]) {
			ut.updateCurrent(fi.Size(), true)
			return nil
		}
	}
	// destination
	stat := fileStat{
		path: filepath.Join(ut.dst, path),
		stat: fi,
	}
	if fi.IsDir() {
		return ut.mkdir(ctx, task, path, &stat)
	}
	return ut.extractFile(ctx, task, path, &stat, file)
}

// mkdir is used to create destination directory if it is not exists.
func (ut *unZipTask) mkdir(ctx context.Context, task *task.Task, src string, dir *fileStat) error {
	// update current task detail, output:
	//   create directory, name: testdata
	//   src: zip/testdata
	//   dst: C:\testdata
	const format = "create directory, name: %s\nsrc: zip/%s\ndst: %s"
	dirName := filepath.Base(src)
	ut.updateDetail(fmt.Sprintf(format, dirName, src, dir.path))
retry:
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	// check destination directory is exist
	dstStat, err := stat(dir.path)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ut.errCtrl,
		}
		retry, ne := noticeFailedToUnZip(&ps, src, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		ut.skipDirs = append(ut.skipDirs, src)
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
			errCtrl: ut.errCtrl,
		}
		stats := SrcDstStat{
			SrcAbs:  src,
			DstAbs:  dir.path,
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
		ut.skipDirs = append(ut.skipDirs, src)
		return nil
	}
	// create directory, must use os.MkdirAll, not os.Mkdir
	// because the target's parent directory maybe not exist.
	err = os.MkdirAll(dir.path, dir.stat.Mode().Perm())
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ut.errCtrl,
		}
		retry, ne := noticeFailedToUnZip(&ps, src, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		ut.skipDirs = append(ut.skipDirs, src)
	}
	return nil
}

func (ut *unZipTask) extractFile(
	ctx context.Context,
	task *task.Task,
	src string,
	file *fileStat,
	zipFile *zip.File,
) error {
	// update current task detail, output:
	//   extract file, name: test.dat, size: 1.127 MB
	//   src: zip/testdata/test.dat
	//   dst: C:\testdata\test.dat
	const format = "extract file, name: %s, size: %s\nsrc: zip/%s\ndst: %s"
	fileName := filepath.Base(src)
	fileSize := convert.FormatByte(uint64(file.stat.Size()))
	ut.updateDetail(fmt.Sprintf(format, fileName, fileSize, src, file.path))
	// check destination
	skipped, err := ut.checkDst(ctx, task, src, file)
	if err != nil {
		return err
	}
	if skipped {
		ut.updateCurrent(file.stat.Size(), true)
		return nil
	}
	// create destination file
retry:
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	perm := file.stat.Mode().Perm()
	dstFile, err := system.OpenFile(file.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ut.errCtrl,
		}
		retry, ne := noticeFailedToUnZip(&ps, src, err)
		if retry {
			goto retry
		}
		if ne != nil {
			return ne
		}
		ut.updateCurrent(file.stat.Size(), true)
		return nil
	}
	defer func() { _ = dstFile.Close() }()
	return ut.writeFile(ctx, task, dstFile, zipFile)
}

// checkDst is used to check destination file is already exists.
func (ut *unZipTask) checkDst(ctx context.Context, task *task.Task, src string, file *fileStat) (bool, error) {
retry:
	// check task is canceled
	if task.Canceled() {
		return false, context.Canceled
	}
	dstStat, err := stat(file.path)
	if err != nil {
		ps := noticePs{
			ctx:     ctx,
			task:    task,
			errCtrl: ut.errCtrl,
		}
		retry, ne := noticeFailedToUnZip(&ps, src, err)
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
	ps := noticePs{
		ctx:     ctx,
		task:    task,
		errCtrl: ut.errCtrl,
	}
	stats := SrcDstStat{
		SrcAbs:  src,
		DstAbs:  file.path,
		SrcStat: file.stat,
		DstStat: dstStat,
	}
	if dstStat.IsDir() {
		retry, ne := noticeSameFileDir(&ps, &stats)
		if retry {
			goto retry
		}
		if ne != nil {
			return false, ne
		}
		return true, nil
	}
	replace, ne := noticeSameFile(&ps, &stats)
	if !replace {
		return true, ne
	}
	return false, nil
}

func (ut *unZipTask) writeFile(ctx context.Context, task *task.Task, dst *os.File, src *zip.File) (err error) {
	dstPath := dst.Name()
	var copied int64
	defer func() {
		if err != nil && err != context.Canceled {
			ps := noticePs{
				ctx:     ctx,
				task:    task,
				errCtrl: ut.errCtrl,
			}
			var retry bool
			retry, err = noticeFailedToUnZip(&ps, src.Name, err)
			if retry {
				// reset current progress
				ut.updateCurrent(copied, false)
				err = ut.retry(ctx, task, dst, src)
				return
			}
			// if failed to extract, delete destination file
			_ = dst.Close()
			_ = os.Remove(dstPath)
			// user cancel
			if err != nil {
				return
			}
			// skipped
			ut.updateCurrent(src.FileInfo().Size()-copied, true)
		}
	}()
	// failed to open zip file can't recover
	rc, err := src.Open()
	if err != nil {
		return
	}
	defer func() { _ = rc.Close() }()
	copied, err = ioCopy(task, ut.addCurrent, dst, rc)
	if err != nil {
		return
	}
	err = dst.Sync()
	if err != nil {
		return
	}
	return os.Chtimes(dstPath, time.Now(), src.Modified)
}

func (ut *unZipTask) addCurrent(delta int64) {
	ut.updateCurrent(delta, true)
}

func (ut *unZipTask) retry(ctx context.Context, task *task.Task, dst *os.File, src *zip.File) error {
	// check task is canceled
	if task.Canceled() {
		return context.Canceled
	}
	// reset offset about opened destination file
	_, err := dst.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return ut.writeFile(ctx, task, dst, src)
}

// Progress is used to get progress about current unzip task.
//
// collect: "0%"
// unzip:   "15.22%|current/total|128 MB/s"
// finish:  "100%"
func (ut *unZipTask) Progress() string {
	ut.rwm.RLock()
	defer ut.rwm.RUnlock()
	// prevent / 0
	if ut.total.Cmp(zeroFloat) == 0 {
		return "0%"
	}
	switch ut.current.Cmp(ut.total) {
	case 0: // current == total
		return "100%"
	case 1: // current > total
		current := ut.current.Text('G', 64)
		total := ut.total.Text('G', 64)
		return fmt.Sprintf("error: current %s > total %s", current, total)
	}
	value := new(big.Float).Quo(ut.current, ut.total)
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
	current := ut.current.Text('G', 64)
	total := ut.total.Text('G', 64)
	speed := convert.FormatByte(ut.speed)
	return fmt.Sprintf("%s%%|%s/%s|%s/s", progress, current, total, speed)
}

func (ut *unZipTask) updateCurrent(delta int64, add bool) {
	d := new(big.Float).SetInt64(delta)
	ut.rwm.Lock()
	defer ut.rwm.Unlock()
	if add {
		ut.current.Add(ut.current, d)
	} else {
		ut.current.Sub(ut.current, d)
	}
}

func (ut *unZipTask) addTotal(delta int64) {
	d := new(big.Float).SetInt64(delta)
	ut.rwm.Lock()
	defer ut.rwm.Unlock()
	ut.total.Add(ut.total, d)
}

// Detail is used to get detail about unzip task.
// read zip file:
//   read zip file
// collect file info:
//   collect file information
//   path: testdata/test.dat
//
// extract file:
//   extract file, name: test.dat, size: 1.127 MB
//   src: testdata/test.dat
//   dst: C:\testdata\test.dat
func (ut *unZipTask) Detail() string {
	ut.rwm.RLock()
	defer ut.rwm.RUnlock()
	return ut.detail
}

func (ut *unZipTask) updateDetail(detail string) {
	ut.rwm.Lock()
	defer ut.rwm.Unlock()
	ut.detail = detail
}

// watcher is used to calculate current extract speed.
func (ut *unZipTask) watcher() {
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "unZipTask.watcher")
		}
	}()
	ticker := time.NewTicker(time.Second / time.Duration(len(ut.speeds)))
	defer ticker.Stop()
	current := new(big.Float)
	index := -1
	for {
		select {
		case <-ticker.C:
			index++
			if index >= len(ut.speeds) {
				index = 0
			}
			ut.watchSpeed(current, index)
		case <-ut.stopSignal:
			return
		}
	}
}

func (ut *unZipTask) watchSpeed(current *big.Float, index int) {
	ut.rwm.Lock()
	defer ut.rwm.Unlock()
	delta := new(big.Float).Sub(ut.current, current)
	current.Add(current, delta)
	// update speed
	ut.speeds[index], _ = delta.Uint64()
	if ut.full {
		ut.speed = 0
		for i := 0; i < len(ut.speeds); i++ {
			ut.speed += ut.speeds[i]
		}
		return
	}
	if index == len(ut.speeds)-1 {
		ut.full = true
	}
	// calculate average speed
	var speed float64 // current speed
	for i := 0; i < index+1; i++ {
		speed += float64(ut.speeds[i])
	}
	ut.speed = uint64(speed / float64(index+1) * float64(len(ut.speeds)))
}

// Clean is used to send stop signal to watcher.
func (ut *unZipTask) Clean() {
	close(ut.stopSignal)
}

// UnZip is used to create a unzip task to extract files from zip file.
func UnZip(errCtrl ErrCtrl, src, dst string, paths ...string) error {
	return UnZipWithContext(context.Background(), errCtrl, src, dst, paths...)
}

// UnZipWithContext is used to create a unzip task with context to extract files from zip file.
func UnZipWithContext(ctx context.Context, errCtrl ErrCtrl, src, dst string, paths ...string) error {
	ut := NewUnZipTask(errCtrl, nil, src, dst, paths...)
	return startTask(ctx, ut, "UnZip")
}
