package filemgr

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

const (
	testZipDir     = "testdata/zip/"             // zip test root path
	testZipDst     = testZipDir + "zip_test.zip" // destination zip file path
	testZipSrcFile = testZipDir + "file1.dat"    // path is a file
	testZipSrcDir  = testZipDir + "dir"          // path is a directory

	// files in the test directory
	testZipSrcFile1 = testZipSrcDir + "/afile1.dat"  // testdata/zip/dir/afile1.dat
	testZipSrcDir1  = testZipSrcDir + "/dir1"        // testdata/zip/dir/dir1
	testZipSrcFile2 = testZipSrcDir1 + "/afile2.dat" // testdata/zip/dir/dir1/afile2.dat
	testZipSrcDir2  = testZipSrcDir1 + "/dir2"       // testdata/zip/dir/dir1/dir2
	testZipSrcDir3  = testZipSrcDir + "/dir3"        // testdata/zip/dir/dir3
	testZipSrcDir4  = testZipSrcDir3 + "/dir4"       // testdata/zip/dir/dir3/dir4
	testZipSrcFile3 = testZipSrcDir4 + "/file3.dat"  // testdata/zip/dir/dir3/dir4/file3.dat
	testZipSrcFile4 = testZipSrcDir3 + "/file4.dat"  // testdata/zip/dir/dir3/file4.dat
	testZipSrcFile5 = testZipSrcDir + "/file5.dat"   // testdata/zip/dir/file5.dat
)

func testCreateZipSrcFile(t *testing.T) {
	testCreateFile(t, testZipSrcFile)
}

func testCreateZipSrcDir(t *testing.T) {
	err := os.MkdirAll(testZipSrcDir, 0750)
	require.NoError(t, err)

	testCreateFile(t, testZipSrcFile1)
	err = os.Mkdir(testZipSrcDir1, 0750)
	require.NoError(t, err)
	testCreateFile2(t, testZipSrcFile2)
	err = os.Mkdir(testZipSrcDir2, 0750)
	require.NoError(t, err)
	err = os.Mkdir(testZipSrcDir3, 0750)
	require.NoError(t, err)
	err = os.Mkdir(testZipSrcDir4, 0750)
	require.NoError(t, err)
	testCreateFile(t, testZipSrcFile3)
	testCreateFile2(t, testZipSrcFile4)
	testCreateFile2(t, testZipSrcFile5)
}

func testRemoveZipDir(t *testing.T) {
	err := os.RemoveAll(testZipDir)
	require.NoError(t, err)
}

func testCheckZipWithFile(t *testing.T) {
	zipFile, err := zip.OpenReader(testZipDst)
	require.NoError(t, err)
	defer func() { _ = zipFile.Close() }()

	require.Len(t, zipFile.File, 1)
	rc, err := zipFile.File[0].Open()
	require.NoError(t, err)

	data, err := ioutil.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, testsuite.Bytes(), data)

	err = rc.Close()
	require.NoError(t, err)
}

func testCheckZipWithDir(t *testing.T) {
	zipFile, err := zip.OpenReader(testZipDst)
	require.NoError(t, err)
	defer func() { _ = zipFile.Close() }()

	require.Len(t, zipFile.File, 10)
	for i, item := range [...]*struct {
		name  string
		data  []byte
		isDir bool
	}{
		{testZipSrcDir, nil, true},
		{testZipSrcFile1, testsuite.Bytes(), false},
		{testZipSrcDir1, nil, true},
		{testZipSrcFile2, bytes.Repeat(testsuite.Bytes(), 2), false},
		{testZipSrcDir2, nil, true},
		{testZipSrcDir3, nil, true},
		{testZipSrcDir4, nil, true},
		{testZipSrcFile3, testsuite.Bytes(), false},
		{testZipSrcFile4, bytes.Repeat(testsuite.Bytes(), 2), false},
		{testZipSrcFile5, bytes.Repeat(testsuite.Bytes(), 2), false},
	} {
		file := zipFile.File[i]
		// check is dir
		require.Equal(t, item.isDir, file.FileInfo().IsDir())
		// check name
		expectName := strings.ReplaceAll(item.name, testZipDir, "")
		expectName = strings.ReplaceAll(expectName, "/", "\\")
		if item.isDir {
			expectName += "/"
		}
		require.Equal(t, expectName, file.Name)
		// check file data
		if item.isDir {
			require.Equal(t, file.FileInfo().Size(), int64(0))
			continue
		}
		rc, err := file.Open()
		require.NoError(t, err)

		data, err := ioutil.ReadAll(rc)
		require.NoError(t, err)
		require.Equal(t, item.data, data)

		err = rc.Close()
		require.NoError(t, err)
	}
}

func testCheckZipWithMulti(t *testing.T) {
	zipFile, err := zip.OpenReader(testZipDst)
	require.NoError(t, err)
	defer func() { _ = zipFile.Close() }()

	require.Len(t, zipFile.File, 10+1)
	for i, item := range [...]*struct {
		name  string
		data  []byte
		isDir bool
	}{
		{testZipSrcDir, nil, true},
		{testZipSrcFile1, testsuite.Bytes(), false},
		{testZipSrcDir1, nil, true},
		{testZipSrcFile2, bytes.Repeat(testsuite.Bytes(), 2), false},
		{testZipSrcDir2, nil, true},
		{testZipSrcDir3, nil, true},
		{testZipSrcDir4, nil, true},
		{testZipSrcFile3, testsuite.Bytes(), false},
		{testZipSrcFile4, bytes.Repeat(testsuite.Bytes(), 2), false},
		{testZipSrcFile5, bytes.Repeat(testsuite.Bytes(), 2), false},
		{testZipSrcFile, testsuite.Bytes(), false},
	} {
		file := zipFile.File[i]
		// check name
		expectName := strings.ReplaceAll(item.name, testZipDir, "")
		expectName = strings.ReplaceAll(expectName, "/", "\\")
		if item.isDir {
			expectName += "/"
		}
		require.Equal(t, expectName, file.Name)
		// check is dir
		require.Equal(t, item.isDir, file.FileInfo().IsDir())
		// check file data
		if item.isDir {
			require.Equal(t, file.FileInfo().Size(), int64(0))
			continue
		}
		rc, err := file.Open()
		require.NoError(t, err)

		data, err := ioutil.ReadAll(rc)
		require.NoError(t, err)
		require.Equal(t, item.data, data)

		err = rc.Close()
		require.NoError(t, err)
	}
}

func TestZip(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("file", func(t *testing.T) {
		testCreateZipSrcFile(t)
		defer testRemoveZipDir(t)

		err := Zip(SkipAll, testZipDst, testZipSrcFile)
		require.NoError(t, err)

		testCheckZipWithFile(t)
	})

	t.Run("directory", func(t *testing.T) {
		testCreateZipSrcDir(t)
		defer testRemoveZipDir(t)

		err := Zip(SkipAll, testZipDst, testZipSrcDir)
		require.NoError(t, err)

		testCheckZipWithDir(t)
	})

	t.Run("multi", func(t *testing.T) {
		t.Run("file first", func(t *testing.T) {
			testCreateZipSrcFile(t)
			testCreateZipSrcDir(t)
			defer testRemoveZipDir(t)

			err := Zip(SkipAll, testZipDst, testZipSrcFile, testZipSrcDir)
			require.NoError(t, err)

			testCheckZipWithMulti(t)
		})

		t.Run("directory first", func(t *testing.T) {
			testCreateZipSrcDir(t)
			testCreateZipSrcFile(t)
			defer testRemoveZipDir(t)

			err := Zip(SkipAll, testZipDst, testZipSrcDir, testZipSrcFile)
			require.NoError(t, err)

			testCheckZipWithMulti(t)
		})
	})

	t.Run("empty path", func(t *testing.T) {
		err := Zip(SkipAll, testZipDst)
		require.Error(t, err)
	})

	t.Run("path doesn't exist", func(t *testing.T) {
		const path = "not exist"

		t.Run("cancel", func(t *testing.T) {
			count := 0
			ec := func(_ context.Context, typ uint8, err error, _ *SrcDstStat) uint8 {
				require.Equal(t, ErrCtrlCollectFailed, typ)
				require.Error(t, err)
				count++
				return ErrCtrlOpCancel
			}
			err := Zip(ec, testZipDst, path)
			require.Equal(t, ErrUserCanceled, err)

			testIsNotExist(t, testZipDst)

			require.Equal(t, 1, count)
		})

		t.Run("skip", func(t *testing.T) {
			count := 0
			ec := func(_ context.Context, typ uint8, err error, _ *SrcDstStat) uint8 {
				require.Equal(t, ErrCtrlCollectFailed, typ)
				require.Error(t, err)
				count++
				return ErrCtrlOpSkip
			}
			err := Zip(ec, testZipDst, path)
			require.NoError(t, err)

			// it will a create a empty zip file
			err = os.Remove(testZipDst)
			require.NoError(t, err)

			require.Equal(t, 1, count)
		})
	})
}

func TestZipTask_Progress(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	pg := testPatchTaskCanceled()
	defer pg.Unpatch()

	t.Run("common", func(t *testing.T) {
		testCreateZipSrcDir(t)
		defer testRemoveZipDir(t)

		zt := NewZipTask(SkipAll, nil, testZipDst, testZipSrcDir)

		done := make(chan struct{})
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-done:
					return
				default:
				}
				fmt.Println("progress:", zt.Progress())
				fmt.Println("detail:", zt.Detail())
				fmt.Println()
				time.Sleep(250 * time.Millisecond)
			}
		}()

		err := zt.Start()
		require.NoError(t, err)

		close(done)
		wg.Wait()

		fmt.Println("progress:", zt.Progress())
		fmt.Println("detail:", zt.Detail())

		rzt := zt.Task().(*zipTask)
		testsuite.IsDestroyed(t, zt)
		testsuite.IsDestroyed(t, rzt)

		testCheckZipWithDir(t)
	})
}
