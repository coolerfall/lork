package lork

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type eventWriterTestSink struct {
	synchronized bool
	filter       Filter
	onWrite      func(*LogEvent) error
	starts       int
	stops        int
}

func (*eventWriterTestSink) Name() string         { return "event-writer-test" }
func (w *eventWriterTestSink) Filter() Filter     { return w.filter }
func (w *eventWriterTestSink) Synchronized() bool { return w.synchronized }
func (w *eventWriterTestSink) Start()             { w.starts++ }
func (w *eventWriterTestSink) Stop()              { w.stops++ }
func (w *eventWriterTestSink) Write(e *LogEvent) error {
	if w.onWrite != nil {
		return w.onWrite(e)
	}
	return nil
}

type eventWriterTestFilter func(*LogEvent) FilterReply

func (f eventWriterTestFilter) Do(e *LogEvent) FilterReply { return f(e) }

func TestEventWriterSynchronization(t *testing.T) {
	for _, synchronized := range []bool{true, false} {
		name := "concurrent"
		if synchronized {
			name = "synchronized"
		}
		for _, stage := range []string{"filter", "write"} {
			t.Run(name+"/"+stage, func(t *testing.T) {
				entered, release := make(chan struct{}, 2), make(chan struct{})
				var releaseOnce sync.Once
				var workers sync.WaitGroup
				t.Cleanup(func() {
					releaseOnce.Do(func() { close(release) })
					workers.Wait()
				})
				block := func() { entered <- struct{}{}; <-release }
				sink := &eventWriterTestSink{synchronized: synchronized}
				if stage == "filter" {
					sink.filter = eventWriterTestFilter(func(*LogEvent) FilterReply { block(); return Accept })
				} else {
					sink.onWrite = func(*LogEvent) error { block(); return nil }
				}
				writer := NewEventWriter(sink)
				done := make(chan struct{}, 2)
				write := func() {
					defer workers.Done()
					event := NewLogEvent()
					defer event.Recycle()
					if err := writer.DoWrite(event); err != nil {
						t.Error(err)
					}
					done <- struct{}{}
				}
				workers.Add(1)
				go write()
				waitWriterSignal(t, entered, "first call")
				workers.Add(1)
				go write()
				if synchronized {
					select {
					case <-entered:
						t.Error("Synchronized=true allowed concurrent calls")
					case <-time.After(100 * time.Millisecond):
					}
				} else {
					waitWriterSignal(t, entered, "concurrent call")
				}
				releaseOnce.Do(func() { close(release) })
				waitWriterSignal(t, done, "first completion")
				waitWriterSignal(t, done, "second completion")
			})
		}
	}
}

func TestEventWriterPreservesBehavior(t *testing.T) {
	for _, synchronized := range []bool{true, false} {
		sink := &eventWriterTestSink{synchronized: synchronized}
		writer := NewEventWriter(sink)
		if writer.Name() != sink.Name() {
			t.Fatal("writer name changed")
		}
		lifecycle := writer.(Lifecycle)
		lifecycle.Start()
		event := NewLogEvent()
		event.unixNano = 123456789
		wantErr := errors.New("write failed")
		calls := 0
		sink.onWrite = func(got *LogEvent) error {
			calls++
			if got != event || got.Timestamp() != 123456789 {
				t.Error("event or timestamp changed")
			}
			return wantErr
		}
		if err := writer.DoWrite(event); err != wantErr {
			t.Errorf("DoWrite error = %v, want %v", err, wantErr)
		}
		sink.filter = eventWriterTestFilter(func(*LogEvent) FilterReply { return Deny })
		if err := writer.DoWrite(event); err != nil {
			t.Errorf("filtered DoWrite error = %v", err)
		}
		lifecycle.Stop()
		event.Recycle()
		if calls != 1 || sink.starts != 1 || sink.stops != 1 {
			t.Errorf("write/start/stop calls = %d/%d/%d, want 1/1/1", calls, sink.starts, sink.stops)
		}
	}
}

func TestEventWriterZeroAllocations(t *testing.T) {
	for _, synchronized := range []bool{true, false} {
		writer := NewEventWriter(&eventWriterTestSink{synchronized: synchronized})
		event := NewLogEvent()
		allocs := testing.AllocsPerRun(1000, func() {
			if err := writer.DoWrite(event); err != nil {
				t.Error(err)
			}
		})
		event.Recycle()
		if allocs != 0 {
			t.Errorf("Synchronized=%v: %v allocations per write, want 0", synchronized, allocs)
		}
	}
}
