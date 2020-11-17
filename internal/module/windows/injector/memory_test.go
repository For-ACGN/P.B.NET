package injector

import (
	"testing"
)

func TestSplitShellcode(t *testing.T) {
	shellcode := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	splitSize := len(shellcode) / 2

	t.Log(splitSize)

	// secondStage first copy for hide special header
	firstStage := shellcode[:splitSize]
	secondStage := shellcode[splitSize:]

	// first size must one byte for pass some AV
	nextSize := 1
	l := len(secondStage)
	for i := 0; i < l; {
		if i+nextSize > l {
			nextSize = l - i
		}

		t.Log("bytes:", secondStage[i:i+nextSize])
		t.Log("address:", splitSize+i)

		i += nextSize
		nextSize = 4 // set random
	}

	nextSize = 1
	l = len(firstStage)
	for i := 0; i < l; {
		if i+nextSize > l {
			nextSize = l - i
		}

		t.Log("bytes:", firstStage[i:i+nextSize])
		t.Log("address:", i)

		i += nextSize
		nextSize = 4 // random
	}

	// b [5]
	// addr 4
	// b [6 7 8 9]
	// addr 5
	// b [1]
	// addr 0
	// b [2 3 4]
	// addr 1
}
