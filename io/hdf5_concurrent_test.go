package io

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/flowmatters/openwater-core/data"

	"github.com/stretchr/testify/assert"
	"gonum.org/v1/hdf5"
)

// These tests exercise the locking in hdf5_util.go and are the main reason to
// run the suite against both libhdf5 configurations:
//
//	source ./build/hdf5/env.sh          # thread-safe libhdf5 (recommended)
//	OW_TEST_PATH=$PWD/test/files go test ./io/
//	source ./build/hdf5/env.sh system   # whatever libhdf5 the system provides
//	OW_TEST_PATH=$PWD/test/files go test ./io/
//
// The code takes materially different paths in each case: a thread-safe
// libhdf5 serializes internally per C call and the Go-side mutex is skipped,
// otherwise every HDF5 call runs under a single process-wide Go mutex.

// CI sets OW_EXPECT_HDF5_THREADSAFE so that a run meant to cover one locking
// path fails loudly if it ends up linked against the other kind of libhdf5.
func TestHDF5ThreadSafetyReported(t *testing.T) {
	t.Logf("libhdf5 thread-safe: %v", IsHDF5ThreadSafe())

	expected := os.Getenv("OW_EXPECT_HDF5_THREADSAFE")
	if expected == "" {
		return
	}
	want, err := strconv.ParseBool(expected)
	if err != nil {
		t.Fatalf("OW_EXPECT_HDF5_THREADSAFE=%q is not a boolean", expected)
	}
	assert.Equal(t, want, IsHDF5ThreadSafe(),
		"linked against the wrong libhdf5 for this run — check CGO_CFLAGS/CGO_LDFLAGS")
}

// TestWriteTwoFilesUnderOneLock covers ow-sim writing outputs and final states
// to separate files. Nesting WithWriteFile for that deadlocks on the Go-side
// mutex; WithWriteFiles takes it once.
func TestWriteTwoFilesUnderOneLock(t *testing.T) {
	assert := assert.New(t)
	dir := t.TempDir()

	the_data, err := data.ARange[float64](1000).Reshape([]int{10, 20, 5})
	assert.Nil(err)

	outputs := filepath.Join(dir, "_outputs.h5")
	states := filepath.Join(dir, "_states.h5")
	for _, fn := range []string{outputs, states} {
		assert.Nil(H5Ref[float64]{Filename: fn, Dataset: "data"}.Write(the_data))
	}

	done := make(chan error, 1)
	go func() {
		done <- WithWriteFiles(outputs, states, func(f1, f2 *hdf5.File) error {
			if f1 == f2 {
				return fmt.Errorf("separate files share a handle")
			}
			for _, f := range []*hdf5.File{f1, f2} {
				if _, err := (H5Ref[float64]{Dataset: "data"}).loadFromOpenFile(f); err != nil {
					return err
				}
			}
			return nil
		})
	}()
	select {
	case err := <-done:
		assert.Nil(err)
	case <-time.After(30 * time.Second):
		t.Fatal("WithWriteFiles did not return — deadlocked on the HDF5 mutex?")
	}

	assert.Nil(WithWriteFiles(outputs, outputs, func(f1, f2 *hdf5.File) error {
		assert.Same(f1, f2)
		return nil
	}))
}

// TestConcurrentReadWrite mimics ow-sim's access pattern: several reader
// goroutines pulling datasets while another goroutine writes, across a handful
// of files.
func TestConcurrentReadWrite(t *testing.T) {
	assert := assert.New(t)
	dir := t.TempDir()

	const nFiles = 4
	const nReaders = 16
	const nRounds = 8

	shape := []int{10, 20, 5}
	expected, err := data.ARange[float64](1000).Reshape(shape)
	assert.Nil(err)

	filename := func(i int) string {
		return filepath.Join(dir, fmt.Sprintf("_concurrent_%d.h5", i))
	}

	for i := 0; i < nFiles; i++ {
		ref := H5Ref[float64]{Filename: filename(i), Dataset: "data"}
		assert.Nil(ref.Write(expected))
	}

	writeTarget := filepath.Join(dir, "_concurrent_writes.h5")

	errs := make(chan error, (nReaders+1)*nRounds)
	var wg sync.WaitGroup

	for r := 0; r < nReaders; r++ {
		wg.Add(1)
		go func(reader int) {
			defer wg.Done()
			for round := 0; round < nRounds; round++ {
				ref := H5Ref[float64]{
					Filename: filename((reader + round) % nFiles),
					Dataset:  "data",
				}
				actual, err := ref.Load()
				if err != nil {
					errs <- fmt.Errorf("reader %d round %d: %w", reader, round, err)
					return
				}
				// Spot check rather than compare all 1000 values, so a torn
				// read shows up without dominating the run time.
				for _, idx := range [][]int{{0, 0, 0}, {5, 14, 2}, {9, 19, 4}} {
					if actual.Get(idx) != expected.Get(idx) {
						errs <- fmt.Errorf("reader %d round %d: %v = %v, want %v",
							reader, round, idx, actual.Get(idx), expected.Get(idx))
						return
					}
				}
			}
		}(r)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for round := 0; round < nRounds; round++ {
			ref := H5Ref[float64]{
				Filename: writeTarget,
				Dataset:  fmt.Sprintf("round_%d", round),
			}
			if err := ref.Write(expected); err != nil {
				errs <- fmt.Errorf("writer round %d: %w", round, err)
				return
			}
		}
	}()

	wg.Wait()
	close(errs)
	for err := range errs {
		assert.Nil(err)
	}

	for round := 0; round < nRounds; round++ {
		ref := H5Ref[float64]{Filename: writeTarget, Dataset: fmt.Sprintf("round_%d", round)}
		actual, err := ref.Load()
		assert.Nil(err)
		if err == nil {
			assert.Equal(expected.Get([]int{5, 14, 2}), actual.Get([]int{5, 14, 2}))
		}
	}
}

// TestConcurrentAccessDoesNotLeakThreads guards the lock ordering in
// hdf5_util.go. lockHDF5 pins the goroutine to its OS thread (so HDF5's
// per-thread error-print setting applies to the thread that runs the
// subsequent cgo calls). If that pin were taken *before* the Go mutex,
// every goroutine waiting on the mutex would park while holding an OS
// thread, and the runtime would spin up one replacement thread per waiter.
func TestConcurrentAccessDoesNotLeakThreads(t *testing.T) {
	if IsHDF5ThreadSafe() {
		t.Skip("no Go-side mutex to contend on when libhdf5 is thread-safe")
	}

	assert := assert.New(t)
	dir := t.TempDir()
	fn := filepath.Join(dir, "_thread_growth.h5")

	the_data, err := data.ARange[float64](1000).Reshape([]int{10, 20, 5})
	assert.Nil(err)
	assert.Nil(H5Ref[float64]{Filename: fn, Dataset: "data"}.Write(the_data))

	const nReaders = 64
	before := pprof.Lookup("threadcreate").Count()

	var wg sync.WaitGroup
	for r := 0; r < nReaders; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 8; i++ {
				ref := H5Ref[float64]{Filename: fn, Dataset: "data"}
				if _, err := ref.Load(); err != nil {
					return
				}
			}
		}()
	}
	wg.Wait()

	growth := pprof.Lookup("threadcreate").Count() - before
	// What we are ruling out is growth proportional to nReaders. Measured on
	// a 32-core box: ~3 threads with the current ordering, ~66 with the pin
	// taken before the mutex. Half of nReaders sits well clear of both.
	limit := nReaders / 2
	t.Logf("threads created by %d concurrent readers: %d (limit %d, GOMAXPROCS %d)",
		nReaders, growth, limit, runtime.GOMAXPROCS(0))
	assert.Less(growth, limit,
		"OS thread count grew with the number of waiters — check the lock/pin order in lockHDF5")
}

// silenceChildEnv switches the test binary into the child half of
// TestConcurrentLookupsPrintNoHDF5Diagnostics.
const silenceChildEnv = "OW_HDF5_SILENCE_CHILD"

// TestConcurrentLookupsPrintNoHDF5Diagnostics guards the per-thread error
// silencing in hdf5_util.go. In a thread-safe libhdf5 the auto-error-print
// setting lives on each thread's own error stack, so switching it off once on
// the main thread (as ow-sim does at startup) leaves every other thread
// printing an HDF5-DIAG trace each time a lookup misses. ow-sim misses on
// purpose all the time — it probes for datasets and groups before creating
// them — and a 1.0.2 run against thread-safe libhdf5 produced a log made up
// almost entirely of these traces.
//
// libhdf5 prints from C straight to the process's stderr, which Go can't
// intercept in-process on every platform, so the test re-runs itself as a
// child process and inspects the child's stderr.
func TestConcurrentLookupsPrintNoHDF5Diagnostics(t *testing.T) {
	if os.Getenv(silenceChildEnv) != "" {
		missedConcurrentLookups(t)
		return
	}

	cmd := exec.Command(os.Args[0],
		"-test.run=^TestConcurrentLookupsPrintNoHDF5Diagnostics$", "-test.v")
	cmd.Env = append(os.Environ(), silenceChildEnv+"=1")
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("child run failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "--- PASS: TestConcurrentLookupsPrintNoHDF5Diagnostics") {
		t.Fatalf("child did not run the lookups:\n%s", stdout.String())
	}

	combined := stdout.String() + stderr.String()
	if n := strings.Count(combined, "HDF5-DIAG"); n > 0 {
		first := combined[strings.Index(combined, "HDF5-DIAG"):]
		if len(first) > 600 {
			first = first[:600]
		}
		t.Errorf("libhdf5 printed %d error traces during concurrent lookups "+
			"(thread-safe libhdf5: %v) — check h5_ensure_silenced in hdf5_util.go. First:\n%s",
			n, IsHDF5ThreadSafe(), first)
	}
}

// missedConcurrentLookups is the child half: it silences HDF5 on the main
// thread the way ow-sim does, then looks up datasets that don't exist from
// many goroutines, each on an OS thread of its own.
func missedConcurrentLookups(t *testing.T) {
	hdf5.DisplayErrors(false)

	dir := t.TempDir()
	fn := filepath.Join(dir, "_silence.h5")
	the_data, err := data.ARange[float64](1000).Reshape([]int{10, 20, 5})
	if err != nil {
		t.Fatal(err)
	}
	if err := (H5Ref[float64]{Filename: fn, Dataset: "data"}).Write(the_data); err != nil {
		t.Fatal(err)
	}

	const nWorkers = 32
	var wg sync.WaitGroup
	var hits sync.Map
	for w := 0; w < nWorkers; w++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			// Lock this goroutine to its thread and never unlock: the runtime
			// then retires the thread when the goroutine exits, so each worker
			// runs on a thread no earlier worker has touched (and silenced).
			runtime.LockOSThread()

			missing := []string{
				"no_such_dataset",                          // H5Dopen2 miss
				fmt.Sprintf("no/such/group_%d/ds", worker), // group traversal miss
			}
			for _, ds := range missing {
				if _, err := (H5Ref[float64]{Filename: fn, Dataset: ds}).Load(); err == nil {
					hits.Store(ds, true)
				}
			}
			// A write to a new nested dataset takes ow-sim's
			// probe-then-create path through openOrCreateDataset.
			// Each worker writes its own file: concurrent read-write and
			// read-only opens of one file are refused by HDF5 file locking.
			out := filepath.Join(dir, fmt.Sprintf("_silence_out_%d.h5", worker))
			ds := "created/nested/data"
			if err := (H5Ref[float64]{Filename: out, Dataset: ds}).Write(the_data); err != nil {
				t.Errorf("worker %d: writing %s: %v", worker, ds, err)
			}
		}(w)
	}
	wg.Wait()

	hits.Range(func(k, _ any) bool {
		t.Errorf("lookup of missing dataset %v succeeded — the test is not exercising a miss", k)
		return true
	})
}
