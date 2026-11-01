package lork

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type callbackWriter struct {
	onStart func()
	onStop  func()
	onWrite func(*LogEvent) error
}

func (*callbackWriter) Name() string { return "callback" }

func (w *callbackWriter) Start() {
	if w.onStart != nil {
		w.onStart()
	}
}

func (w *callbackWriter) Stop() {
	if w.onStop != nil {
		w.onStop()
	}
}

func (w *callbackWriter) DoWrite(event *LogEvent) error {
	if w.onWrite != nil {
		return w.onWrite(event)
	}
	return nil
}

func waitWriterSignal(t *testing.T, signal <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", label)
	}
}

func TestMultiWriterResetWaitsForWritesAndUnlocksBeforeStop(t *testing.T) {
	for _, writeBytes := range []bool{false, true} {
		name := "WriteEvent"
		if writeBytes {
			name = "Write"
		}
		t.Run(name, func(t *testing.T) {
			multi := NewMultiWriter()
			entered, release := make(chan struct{}), make(chan struct{})
			stopping, releaseStop := make(chan struct{}), make(chan struct{})
			var releaseOnce, stopOnce sync.Once
			t.Cleanup(func() { releaseOnce.Do(func() { close(release) }); stopOnce.Do(func() { close(releaseStop) }) })
			sink := &callbackWriter{
				onWrite: func(*LogEvent) error { close(entered); <-release; return nil },
				onStop: func() {
					if multi.Size() != 0 || multi.GetWriter("callback") != nil {
						t.Error("Stop ran before the writer was detached")
					}
					close(stopping)
					<-releaseStop
				},
			}
			multi.AddWriter(sink)
			written := make(chan struct{})
			go func() {
				defer close(written)
				if writeBytes {
					data := []byte(`{"level":"INFO","message":"bytes"}`)
					if n, err := multi.Write(data); err != nil || n != len(data) {
						t.Errorf("Write = %d, %v", n, err)
					}
				} else if err := multi.WriteEvent(NewLogEvent()); err != nil {
					t.Error(err)
				}
			}()
			waitWriterSignal(t, entered, "active write")
			reset := make(chan struct{})
			go func() { multi.ResetWriter(); close(reset) }()
			deadline := time.Now().Add(2 * time.Second)
			for multi.locker.TryRLock() {
				multi.locker.RUnlock()
				if time.Now().After(deadline) {
					t.Fatal("ResetWriter did not wait for the active write")
				}
				runtime.Gosched()
			}
			select {
			case <-stopping:
				t.Fatal("downstream Stop ran during an active write")
			default:
			}
			releaseOnce.Do(func() { close(release) })
			waitWriterSignal(t, written, "completed write")
			waitWriterSignal(t, stopping, "downstream Stop")

			// Output and queries must remain available while downstream Stop blocks.
			queried := make(chan struct{})
			go func() {
				defer close(queried)
				if multi.Attached(sink) || multi.Size() != 0 {
					t.Error("detached writer remained visible")
				}
				if n, err := multi.Write([]byte("unused")); n != 0 || err != nil {
					t.Errorf("empty Write = %d, %v", n, err)
				}
				if err := multi.WriteEvent(NewLogEvent()); err != nil {
					t.Error(err)
				}
			}()
			waitWriterSignal(t, queried, "output while stopping")
			stopOnce.Do(func() { close(releaseStop) })
			waitWriterSignal(t, reset, "ResetWriter")
		})
	}
}

func TestMultiWriterConcurrentOutput(t *testing.T) {
	multi := NewMultiWriter()
	entered, release := make(chan struct{}, 2), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	multi.AddWriter(&callbackWriter{onWrite: func(*LogEvent) error {
		entered <- struct{}{}
		<-release
		return nil
	}})
	done := make(chan struct{}, 2)
	for range 2 {
		go func() { _ = multi.WriteEvent(NewLogEvent()); done <- struct{}{} }()
	}
	for range 2 {
		waitWriterSignal(t, entered, "concurrent output")
	}
	releaseOnce.Do(func() { close(release) })
	for range 2 {
		waitWriterSignal(t, done, "completed output")
	}
	multi.ResetWriter()
}

func TestMultiWriterResetDrainsAsyncWriter(t *testing.T) {
	multi := NewMultiWriter()
	writer := NewAsyncWriter(func(options *AsyncWriterOption) { options.QueueSize = 16 })
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	var writes, stops atomic.Int32
	writer.AddWriter(&callbackWriter{
		onWrite: func(*LogEvent) error {
			if writes.Add(1) == 1 {
				close(entered)
				<-release
			}
			return nil
		},
		onStop: func() {
			if writes.Load() != 4 {
				t.Error("downstream stopped before queued events were drained")
			}
			if multi.Size() != 0 {
				t.Error("async writer remained attached during Stop")
			}
			stops.Add(1)
		},
	})
	multi.AddWriter(writer)
	if err := multi.WriteEvent(NewLogEvent()); err != nil {
		t.Fatal(err)
	}
	waitWriterSignal(t, entered, "async output")
	for range 3 {
		if err := multi.WriteEvent(NewLogEvent()); err != nil {
			t.Fatal(err)
		}
	}
	reset := make(chan struct{})
	go func() { multi.ResetWriter(); close(reset) }()
	deadline := time.Now().Add(2 * time.Second)
	for multi.Size() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("async writer was not detached before draining")
		}
		runtime.Gosched()
	}
	select {
	case <-reset:
		t.Fatal("reset returned before active async output completed")
	default:
	}
	releaseOnce.Do(func() { close(release) })
	waitWriterSignal(t, reset, "reset draining queued events")
	if writes.Load() != 4 || stops.Load() != 1 {
		t.Fatalf("writes=%d stops=%d; want 4 writes and 1 stop", writes.Load(), stops.Load())
	}
}

func TestMultiWriterLifecycleCallbacksCanQuery(t *testing.T) {
	multi := NewMultiWriter()
	sink := &callbackWriter{onStart: func() {
		if multi.Size() != 0 || multi.GetWriter("callback") != nil {
			t.Error("writer published before Start completed")
		}
	}}
	sink.onStop = func() {
		if multi.Size() != 0 || multi.GetWriter("callback") != nil || multi.Attached(sink) {
			t.Error("writer remained attached during Stop")
		}
	}
	done := make(chan struct{})
	go func() { multi.AddWriter(sink); multi.ResetWriter(); close(done) }()
	waitWriterSignal(t, done, "lifecycle callbacks")
}

func TestMultiWriterSerializesResetAndAdd(t *testing.T) {
	multi := NewMultiWriter()
	stopping, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })
	var running atomic.Bool
	var starts, stops atomic.Int32
	sink := &callbackWriter{
		onStart: func() {
			if running.Swap(true) {
				t.Error("Start overlapped the previous lifecycle")
			}
			starts.Add(1)
		},
		onStop: func() {
			if stops.Add(1) == 1 {
				close(stopping)
				<-release
			}
			running.Store(false)
		},
	}
	multi.AddWriter(sink)
	reset := make(chan struct{})
	go func() { multi.ResetWriter(); close(reset) }()
	waitWriterSignal(t, stopping, "first Stop")
	added := make(chan struct{})
	go func() { multi.AddWriter(sink); close(added) }()
	releaseOnce.Do(func() { close(release) })
	waitWriterSignal(t, reset, "first reset")
	waitWriterSignal(t, added, "re-attachment")
	if starts.Load() != 2 || stops.Load() != 1 || !multi.Attached(sink) {
		t.Fatal("writer was not restarted after Stop completed")
	}
	multi.ResetWriter()
	if running.Load() || stops.Load() != 2 {
		t.Fatal("restarted writer was not stopped")
	}
}

func TestAsyncWriterLifecycleCallbacksCanQuery(t *testing.T) {
	writer := NewAsyncWriter()
	sink := &callbackWriter{onStart: func() { _ = writer.GetWriter("callback") }}
	sink.onStop = func() {
		if writer.GetWriter("callback") != sink || !writer.Attached(sink) {
			t.Error("Stop removed the retained binding")
		}
	}
	done := make(chan struct{})
	go func() {
		writer.AddWriter(sink)
		for range 2 {
			writer.Start()
			writer.Stop()
		}
		writer.ResetWriter()
		close(done)
	}()
	waitWriterSignal(t, done, "async lifecycle callbacks")
}

func TestMultiWriterConcurrentConfiguration(t *testing.T) {
	multi := NewMultiWriter()
	sink := &asyncLifecycleSink{}
	multi.AddWriter(sink)
	var workers sync.WaitGroup
	for worker := range 6 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 500 {
				switch worker {
				case 0:
					multi.ResetWriter()
					multi.AddWriter(sink)
				case 1:
					_, _ = multi.Write([]byte(`{"level":"INFO","message":"bytes"}`))
				case 2:
					_ = multi.GetWriter("sink")
					_ = multi.Attached(sink)
					_ = multi.Size()
				default:
					_ = multi.WriteEvent(NewLogEvent())
				}
			}
		}()
	}
	workers.Wait()
	multi.ResetWriter()
	sink.mu.Lock()
	defer sink.mu.Unlock()
	if sink.invalid || sink.running || sink.starts != sink.stops {
		t.Fatalf("invalid lifecycle: starts=%d stops=%d running=%v invalid=%v",
			sink.starts, sink.stops, sink.running, sink.invalid)
	}
}

func TestAsyncWriterConcurrentConfiguration(t *testing.T) {
	writer := NewAsyncWriter()
	sink := &asyncLifecycleSink{}
	writer.AddWriter(sink)
	writer.Start()
	var workers sync.WaitGroup
	for worker := range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 500 {
				switch worker {
				case 0:
					writer.ResetWriter()
					writer.AddWriter(sink)
				case 1:
					_ = writer.GetWriter("sink")
					_ = writer.Attached(sink)
				default:
					event := NewLogEvent()
					_ = writer.DoWrite(event)
					event.Recycle()
				}
			}
		}()
	}
	workers.Wait()
	writer.Stop()
	writer.ResetWriter()
	sink.mu.Lock()
	defer sink.mu.Unlock()
	if sink.invalid || sink.running || sink.starts != sink.stops {
		t.Fatalf("invalid async lifecycle: starts=%d stops=%d running=%v invalid=%v",
			sink.starts, sink.stops, sink.running, sink.invalid)
	}
}
