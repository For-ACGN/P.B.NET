package filemgr

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

const (
	testDeleteDir = "testdata/delete/"

	// src is file
	testDeleteSrcFile = testDeleteDir + "file1.dat"

	// src is directory
	testDeleteSrcDir = testDeleteDir + "dir"

	// src files in directory
	testDeleteSrcFile1 = testDeleteSrcDir + "/afile1.dat"  // testdata/delete/dir/afile1.dat
	testDeleteSrcDir1  = testDeleteSrcDir + "/dir1"        // testdata/delete/dir/dir1
	testDeleteSrcFile2 = testDeleteSrcDir1 + "/afile2.dat" // testdata/delete/dir/dir1/afile2.dat
	testDeleteSrcDir2  = testDeleteSrcDir1 + "/dir2"       // testdata/delete/dir/dir1/dir2
	testDeleteSrcDir3  = testDeleteSrcDir + "/dir3"        // testdata/delete/dir/dir3
	testDeleteSrcDir4  = testDeleteSrcDir3 + "/dir4"       // testdata/delete/dir/dir3/dir4
	testDeleteSrcFile3 = testDeleteSrcDir4 + "/file3.dat"  // testdata/delete/dir/dir3/dir4/file3.dat
	testDeleteSrcFile4 = testDeleteSrcDir3 + "/file4.dat"  // testdata/delete/dir/dir3/file4.dat
	testDeleteSrcFile5 = testDeleteSrcDir + "/file5.dat"   // testdata/delete/dir/file5.dat
)

func testCreateDeleteSrcFile(t *testing.T) {
	testCreateFile(t, testDeleteSrcFile)
}

func testCreateDeleteSrcDir(t *testing.T) {
	err := os.MkdirAll(testDeleteSrcDir, 0750)
	require.NoError(t, err)

	testCreateFile(t, testDeleteSrcFile1)
	err = os.Mkdir(testDeleteSrcDir1, 0750)
	require.NoError(t, err)
	testCreateFile2(t, testDeleteSrcFile2)
	err = os.Mkdir(testDeleteSrcDir2, 0750)
	require.NoError(t, err)
	err = os.Mkdir(testDeleteSrcDir3, 0750)
	require.NoError(t, err)
	err = os.Mkdir(testDeleteSrcDir4, 0750)
	require.NoError(t, err)
	testCreateFile(t, testDeleteSrcFile3)
	testCreateFile2(t, testDeleteSrcFile4)
	testCreateFile2(t, testDeleteSrcFile5)
}

func testRemoveDeleteDir(t *testing.T) {
	err := os.RemoveAll(testDeleteDir)
	require.NoError(t, err)
}

func TestDelete(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("file", func(t *testing.T) {
		testCreateDeleteSrcFile(t)
		defer testRemoveDeleteDir(t)

		err := Delete(SkipAll, testDeleteSrcFile)
		require.NoError(t, err)

		testIsNotExist(t, testDeleteSrcFile)
	})

	t.Run("directory", func(t *testing.T) {
		testCreateDeleteSrcDir(t)
		defer testRemoveDeleteDir(t)

		err := Delete(SkipAll, testDeleteSrcDir)
		require.NoError(t, err)

		testIsNotExist(t, testDeleteSrcDir)
	})

	t.Run("multi", func(t *testing.T) {
		t.Run("file first", func(t *testing.T) {
			testCreateDeleteSrcFile(t)
			testCreateDeleteSrcDir(t)
			defer testRemoveDeleteDir(t)

			err := Delete(SkipAll, testDeleteSrcFile, testDeleteSrcDir)
			require.NoError(t, err)

			testIsNotExist(t, testDeleteSrcFile)
			testIsNotExist(t, testDeleteSrcDir)
		})

		t.Run("directory first", func(t *testing.T) {
			testCreateDeleteSrcDir(t)
			testCreateDeleteSrcFile(t)
			defer testRemoveDeleteDir(t)

			err := Delete(SkipAll, testDeleteSrcDir, testDeleteSrcFile)
			require.NoError(t, err)

			testIsNotExist(t, testDeleteSrcDir)
			testIsNotExist(t, testDeleteSrcFile)
		})
	})

	t.Run("path doesn't exist", func(t *testing.T) {
		const path = "not exist"

		count := 0
		ec := func(_ context.Context, typ uint8, err error, _ *SrcDstStat) uint8 {
			require.Equal(t, ErrCtrlCollectFailed, typ)
			require.Error(t, err)
			count++
			return ErrCtrlOpSkip
		}
		err := Delete(ec, path)
		require.NoError(t, err)

		testIsNotExist(t, path)
		require.Equal(t, 1, count)
	})

	t.Run("failed to remove file", func(t *testing.T) {
		testCreateDeleteSrcFile(t)
		defer testRemoveDeleteDir(t)

		patch := func(string) error {
			return monkey.Error
		}
		pg := monkey.Patch(os.Remove, patch)
		defer pg.Unpatch()

		count := 0
		ec := func(_ context.Context, typ uint8, err error, _ *SrcDstStat) uint8 {
			require.Equal(t, ErrCtrlDeleteFailed, typ)
			require.Error(t, err)
			count++
			return ErrCtrlOpSkip
		}
		err := Delete(ec, testDeleteSrcFile)
		require.NoError(t, err)

		testIsExist(t, testDeleteSrcFile)
	})
}

func TestDelete_File(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		testCreateDeleteSrcFile(t)
		defer testRemoveDeleteDir(t)

		err := Delete(SkipAll, testDeleteSrcFile)
		require.NoError(t, err)

		testIsNotExist(t, testDeleteSrcFile)
	})

}

func TestDelete_Directory(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {
		testCreateDeleteSrcDir(t)
		defer testRemoveDeleteDir(t)

		err := Delete(SkipAll, testDeleteSrcDir)
		require.NoError(t, err)

		testIsNotExist(t, testDeleteSrcDir)
	})
}

func TestDelete_Multi(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("ok", func(t *testing.T) {

	})
}
