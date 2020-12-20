package unsafe

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

type testStruct struct {
	a int32
	b int64
}

func TestSizeof(t *testing.T) {
	require.Equal(t, unsafe.Sizeof(int64(1)), sizeof(int64(1)))
	require.Equal(t, unsafe.Sizeof(int32(1)), sizeof(int32(1)))
	require.Equal(t, unsafe.Sizeof(testStruct{}), sizeof(testStruct{}))
}

func TestAlignof(t *testing.T) {
	require.Equal(t, unsafe.Alignof(testStruct{}), alignof(testStruct{}))
	require.Equal(t, unsafe.Alignof(testStruct{}.a), alignof(testStruct{}.a))
	require.Equal(t, unsafe.Alignof(testStruct{}.b), alignof(testStruct{}.b))
}

func TestOffsetof(t *testing.T) {
	require.Equal(t, unsafe.Offsetof(testStruct{}.a), offsetof(testStruct{}, "a"))
	require.Equal(t, unsafe.Offsetof(testStruct{}.b), offsetof(testStruct{}, "b"))
}
func TestOffsetof_NotExist(t *testing.T) {
	defer testsuite.DeferForPanic(t)
	offsetof(testStruct{}, "c")
}
