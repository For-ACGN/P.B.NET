package nmap

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"

	"project/internal/logger"
	"project/internal/testsuite"
)

func testGenerateScanner() (*Scanner, chan<- *Job) {
	opts := Options{
		NoPing:     true,
		Service:    true,
		OS:         true,
		OutputPath: "testdata/output",
	}
	jobCh := make(chan *Job, 1024)
	scanner := New(jobCh, 4, logger.Test, &opts)
	return scanner, jobCh
}

func TestScanner(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	defer func() {
		err := os.RemoveAll("testdata/output")
		require.NoError(t, err)
	}()

	scanner, jobCh := testGenerateScanner()

	// send job
	jobCh <- &Job{
		Protocol: "tcp",
		Target:   "127.0.0.1",
		Port:     "1-1024",
		Extra:    "test",
	}
	err := scanner.Start()
	require.NoError(t, err)

	select {
	case result := <-scanner.Result:
		require.NoError(t, result.Error)
		require.Equal(t, "test", result.Job.Extra)
		spew.Dump(result.Output)
		fmt.Println("worker id:", result.WorkerID)
		fmt.Println("elapsed time:", result.ElapsedTime)
	case <-time.After(3 * time.Minute):
		t.Fatal("receive scan result timeout")
	}

	scanner.Stop()

	testsuite.IsDestroyed(t, scanner)
}
