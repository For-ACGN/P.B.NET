package filemgr

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/system"
	"project/internal/testsuite"
)

func testCreateFile(t *testing.T, name string) {
	data := testsuite.Bytes()
	err := system.WriteFile(name, data)
	require.NoError(t, err)
}

func testCreateFile2(t *testing.T, name string) {
	data := append(testsuite.Bytes(), testsuite.Bytes()...)
	err := system.WriteFile(name, data)
	require.NoError(t, err)
}

func testCompareFile(t *testing.T, a, b string) {
	aFile, err := os.Open(a)
	require.NoError(t, err)
	defer func() { _ = aFile.Close() }()
	bFile, err := os.Open(b)
	require.NoError(t, err)
	defer func() { _ = bFile.Close() }()

	// compare stat
	aStat, err := aFile.Stat()
	require.NoError(t, err)
	bStat, err := bFile.Stat()
	require.NoError(t, err)

	require.Equal(t, aStat.Size(), bStat.Size())
	require.Equal(t, aStat.Mode(), bStat.Mode())
	require.Equal(t, aStat.IsDir(), bStat.IsDir())

	if !aStat.IsDir() {
		// compare data
		aFileData, err := ioutil.ReadAll(aFile)
		require.NoError(t, err)
		bFileData, err := ioutil.ReadAll(bFile)
		require.NoError(t, err)
		require.Equal(t, aFileData, bFileData)

		// mod time is not equal about wall
		// directory stat may be changed
		const format = "2006-01-02 15:04:05"
		am := aStat.ModTime().Format(format)
		bm := bStat.ModTime().Format(format)
		require.Equal(t, am, bm)
	}
}

func testCompareDirectory(t *testing.T, a, b string) {
	aFiles := make([]string, 0, 4)
	bFiles := make([]string, 0, 4)
	err := filepath.Walk(a, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if path != a {
			aFiles = append(aFiles, path)
		}
		return nil
	})
	require.NoError(t, err)
	err = filepath.Walk(b, func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if path != b {
			bFiles = append(bFiles, path)
		}
		return nil
	})
	require.NoError(t, err)

	// compare file numbers
	aFilesLen := len(aFiles)
	bFilesLen := len(bFiles)
	require.Equal(t, aFilesLen, bFilesLen)

	// compare each file
	for i := 0; i < aFilesLen; i++ {
		testCompareFile(t, aFiles[i], bFiles[i])
	}
}

func TestCopy(t *testing.T) {
	t.Run("src is file", func(t *testing.T) {
		const (
			src     = "testdata/file.dat"
			dstFile = "testdata/file/file.dat"
			dstDir  = "testdata/file/"
		)

		// create test file
		testCreateFile(t, src)
		defer func() {
			err := os.Remove(src)
			require.NoError(t, err)
		}()

		t.Run("to file path", func(t *testing.T) {
			t.Run("destination doesn't exist", func(t *testing.T) {
				defer func() {
					err := os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				err := Copy(ReplaceAll, src, dstFile)
				require.NoError(t, err)

				testCompareFile(t, src, dstFile)
			})

			t.Run("destination exists", func(t *testing.T) {
				defer func() {
					err := os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				t.Run("file", func(t *testing.T) {
					// create test file (exists)
					testCreateFile(t, dstFile)
					defer func() {
						err := os.Remove(dstFile)
						require.NoError(t, err)
					}()

					count := 0
					err := Copy(func(typ uint8, err error, src, dst string) uint8 {
						require.Equal(t, ErrCtrlSameFile, typ)
						count++
						return ErrCtrlOpReplace
					}, src, dstFile)
					require.NoError(t, err)

					testCompareFile(t, src, dstFile)
					require.Equal(t, 1, count)
				})

				t.Run("directory", func(t *testing.T) {
					// create test directory (exists)
					err := os.MkdirAll(dstFile, 0750)
					require.NoError(t, err)
					defer func() {
						err = os.RemoveAll(dstFile)
						require.NoError(t, err)
					}()

					err = Copy(ReplaceAll, src, dstFile)
					require.NoError(t, err)
				})
			})
		})

		t.Run("to directory path", func(t *testing.T) {
			t.Run("destination doesn't exist", func(t *testing.T) {
				defer func() {
					err := os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				err := Copy(ReplaceAll, src, dstDir)
				require.NoError(t, err)

				testCompareFile(t, src, dstFile)
			})

			t.Run("destination exists", func(t *testing.T) {
				defer func() {
					err := os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				t.Run("file", func(t *testing.T) {
					// create test file(exists)
					testCreateFile(t, dstFile)
					defer func() {
						err := os.Remove(dstFile)
						require.NoError(t, err)
					}()

					count := 0
					err := Copy(func(typ uint8, err error, src, dst string) uint8 {
						require.Equal(t, ErrCtrlSameFile, typ)
						count++
						return ErrCtrlOpReplace
					}, src, dstDir)
					require.NoError(t, err)

					testCompareFile(t, src, dstFile)
					require.Equal(t, 1, count)
				})

				t.Run("directory", func(t *testing.T) {
					// create test directory (exists)
					err := os.MkdirAll(dstFile, 0750)
					require.NoError(t, err)
					defer func() {
						err = os.RemoveAll(dstFile)
						require.NoError(t, err)
					}()

					count := 0
					err = Copy(func(typ uint8, err error, src, dst string) uint8 {
						require.Equal(t, ErrCtrlSameFileDir, typ)
						count++
						return ErrCtrlOpSkip
					}, src, dstDir)
					require.NoError(t, err)

					require.Equal(t, 1, count)
				})
			})
		})
	})

	t.Run("src is directory", func(t *testing.T) {
		const (
			srcDir   = "testdata/dir"
			srcFile1 = "file1.dat"
			srcDir2  = "dir2"
			srcFile2 = "dir2/file2.dat"
		)

		// create test directory
		err := os.MkdirAll(srcDir, 0750)
		require.NoError(t, err)
		defer func() {
			err = os.RemoveAll(srcDir)
			require.NoError(t, err)
		}()
		// create test file
		testCreateFile(t, filepath.Join(srcDir, srcFile1))
		// create dir2
		err = os.MkdirAll(filepath.Join(srcDir, srcDir2), 0750)
		require.NoError(t, err)
		// create test file 2
		testCreateFile2(t, filepath.Join(srcDir, srcFile2))

		t.Run("to directory path", func(t *testing.T) {
			const dstDir = "testdata/dir-dir/"

			t.Run("destination doesn't exist", func(t *testing.T) {
				defer func() {
					err = os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				err = Copy(ReplaceAll, srcDir, dstDir)
				require.NoError(t, err)

				testCompareDirectory(t, srcDir, dstDir)
			})

			t.Run("destination exists", func(t *testing.T) {
				err := os.MkdirAll(dstDir, 0750)
				require.NoError(t, err)
				defer func() {
					err := os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				err = Copy(ReplaceAll, srcDir, dstDir)
				require.NoError(t, err)

				testCompareDirectory(t, srcDir, dstDir)
			})
		})

		t.Run("to file path", func(t *testing.T) {
			const dstDir = "testdata/dir-dir"

			t.Run("destination doesn't exist", func(t *testing.T) {
				defer func() {
					err := os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				err := Copy(ReplaceAll, srcDir, dstDir)
				require.NoError(t, err)

				testCompareDirectory(t, srcDir, dstDir)
			})

			t.Run("destination exists", func(t *testing.T) {
				t.Run("file", func(t *testing.T) {
					defer func() {
						err := os.RemoveAll(dstDir)
						require.NoError(t, err)
					}()

					// create test file(exists)
					testCreateFile(t, dstDir)
					defer func() {
						err := os.Remove(dstDir)
						require.NoError(t, err)
					}()

					err := Copy(ReplaceAll, srcDir, dstDir)
					require.Error(t, err)
				})

				t.Run("directory", func(t *testing.T) {
					err := os.MkdirAll(dstDir, 0750)
					require.NoError(t, err)
					defer func() {
						err := os.RemoveAll(dstDir)
						require.NoError(t, err)
					}()

					err = Copy(ReplaceAll, srcDir, dstDir)
					require.NoError(t, err)

					testCompareDirectory(t, srcDir, dstDir)
				})
			})
		})

		t.Run("sub file exist in dst directory", func(t *testing.T) {
			const dstDir = "testdata/dir-dir/"

			t.Run("file to directory", func(t *testing.T) {
				defer func() {
					err := os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				// create exists same name directory
				path := filepath.Join(dstDir, srcFile1)
				err := os.MkdirAll(path, 0750)
				require.NoError(t, err)
				defer func() {
					err := os.Remove(path)
					require.NoError(t, err)
				}()

				count := 0
				err = Copy(func(typ uint8, err error, src string, dst string) uint8 {
					require.Equal(t, ErrCtrlSameFileDir, typ)
					count++
					return ErrCtrlOpSkip
				}, srcDir, dstDir)
				require.NoError(t, err)

				require.Equal(t, 1, count)
			})

			// C:\test\dir2[dir] exists in D:\test\dir2[file]
			// need skip all files in C:\test\dir2
			// like skip C:\test\dir2\file2.dat
			t.Run("directory to file", func(t *testing.T) {
				defer func() {
					err := os.RemoveAll(dstDir)
					require.NoError(t, err)
				}()

				// create exists same name file
				path := filepath.Join(dstDir, srcDir2)
				testCreateFile(t, path)
				defer func() {
					err := os.Remove(path)
					require.NoError(t, err)
				}()

				count := 0
				err := Copy(func(typ uint8, err error, src string, dst string) uint8 {
					require.Equal(t, ErrCtrlSameDirFile, typ)
					count++
					return ErrCtrlOpSkip
				}, srcDir, dstDir)
				require.NoError(t, err)

				require.Equal(t, 1, count)
			})
		})
	})

	t.Run("with context", func(t *testing.T) {
		t.Run("copy file", func(t *testing.T) {
			const (
				src = "testdata/file.dat"
				dst = "testdata/file/file.dat"
				dir = "testdata/file"
			)

			// create test file
			testCreateFile(t, src)
			defer func() {
				err := os.Remove(src)
				require.NoError(t, err)
			}()

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			err := CopyWithContext(ctx, SkipAll, src, dst)
			require.Equal(t, context.Canceled, err)

			exist, err := system.IsExist(dir)
			require.NoError(t, err)
			require.True(t, exist)

			err = os.RemoveAll(dir)
			require.NoError(t, err)
		})

		t.Run("copy directory", func(t *testing.T) {
			const (
				srcDir   = "testdata/dir"
				srcFile1 = "file1.dat"
				srcDir2  = "dir2"
				srcFile2 = "dir2/file2.dat"
				dstDir   = "testdata/dir-dir/"
			)

			// create test directory
			err := os.MkdirAll(srcDir, 0750)
			require.NoError(t, err)
			defer func() {
				err = os.RemoveAll(srcDir)
				require.NoError(t, err)
			}()
			// create test file
			testCreateFile(t, filepath.Join(srcDir, srcFile1))
			// create dir2
			err = os.MkdirAll(filepath.Join(srcDir, srcDir2), 0750)
			require.NoError(t, err)
			// create test file 2
			testCreateFile2(t, filepath.Join(srcDir, srcFile2))

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			err = CopyWithContext(ctx, SkipAll, srcDir, dstDir)
			require.Equal(t, context.Canceled, err)

			exist, err := system.IsNotExist(dstDir)
			require.NoError(t, err)
			require.True(t, exist)
		})
	})

	t.Run("retry", func(t *testing.T) {

	})
}
