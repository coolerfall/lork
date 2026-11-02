package lork

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

type asyncLifecycleSink struct {
	mu       sync.Mutex
	running  bool
	starts   int
	stops    int
	invalid  bool
	messages []string
}

func (s *asyncLifecycleSink) Name() string { return "sink" }

func (s *asyncLifecycleSink) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invalid = s.invalid || s.running
	s.running = true
	s.starts++
}

func (s *asyncLifecycleSink) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invalid = s.invalid || !s.running
	s.running = false
	s.stops++
}

func (s *asyncLifecycleSink) DoWrite(event *LogEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		s.invalid = true
		return errors.New("write to stopped sink")
	}
	s.messages = append(s.messages, string(event.Message()))
	return nil
}

func (s *asyncLifecycleSink) check(t *testing.T, starts, stops int, messages ...string) {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.invalid || s.starts != starts || s.stops != stops || !slices.Equal(s.messages, messages) {
		t.Fatalf("sink: starts=%d stops=%d invalid=%v messages=%q; want starts=%d stops=%d messages=%q",
			s.starts, s.stops, s.invalid, s.messages, starts, stops, messages)
	}
}

func writeAsyncMessage(t *testing.T, writer *AsyncWriter, message string) {
	t.Helper()
	event := NewLogEvent()
	defer event.Recycle()
	event.appendLevel(InfoLevel)
	event.appendMessage(message)
	if err := writer.DoWrite(event); err != nil {
		t.Fatal(err)
	}
}

func TestAsyncWriterRestartRetainsBindings(t *testing.T) {
	writer := NewAsyncWriter()
	t.Cleanup(writer.Stop)
	sink := &asyncLifecycleSink{}
	writer.AddWriter(sink)
	sink.check(t, 1, 0)
	var messages []string
	for cycle := range 3 {
		writer.Start()
		writer.Start()
		message := fmt.Sprintf("cycle %d", cycle)
		writeAsyncMessage(t, writer, message)
		messages = append(messages, message)
		writer.Stop()
		writer.Stop()
		if !writer.Attached(sink) || writer.GetWriter(sink.Name()) != sink {
			t.Fatal("Stop removed the downstream writer")
		}
		sink.check(t, cycle+1, cycle+1, messages...)
		writeAsyncMessage(t, writer, "discarded while stopped")
	}
}

func TestAsyncWriterStopBeforeStartDrainsAndRestarts(t *testing.T) {
	writer := NewAsyncWriter()
	t.Cleanup(writer.Stop)
	sink := &asyncLifecycleSink{}
	writer.AddWriter(sink)
	writeAsyncMessage(t, writer, "queued before Start")
	writer.Stop()
	sink.check(t, 1, 1, "queued before Start")
	writer.Start()
	writeAsyncMessage(t, writer, "after restart")
	writer.Stop()
	sink.check(t, 2, 2, "queued before Start", "after restart")
}

func TestAsyncWriterAddWhileStoppedDefersStart(t *testing.T) {
	writer := NewAsyncWriter()
	t.Cleanup(writer.Stop)
	first, second := &asyncLifecycleSink{}, &asyncLifecycleSink{}
	writer.AddWriter(first)
	writer.Start()
	writer.Stop()
	writer.AddWriter(second)
	second.check(t, 0, 0)
	writer.Start()
	writer.Start()
	writeAsyncMessage(t, writer, "both writers")
	writer.Stop()
	first.check(t, 2, 2, "both writers")
	second.check(t, 1, 1, "both writers")
}

func TestAsyncWriterResetWhileStoppedRemovesBindings(t *testing.T) {
	writer := NewAsyncWriter()
	t.Cleanup(writer.Stop)
	first, second := &asyncLifecycleSink{}, &asyncLifecycleSink{}
	writer.AddWriter(first)
	writer.Stop()
	writer.AddWriter(second)
	writer.ResetWriter()
	writer.ResetWriter()
	if writer.Attached(first) || writer.Attached(second) || writer.GetWriter(first.Name()) != nil {
		t.Fatal("ResetWriter retained a downstream writer")
	}
	writer.Start()
	writeAsyncMessage(t, writer, "no destination")
	writer.Stop()
	first.check(t, 1, 1)
	second.check(t, 0, 0)
}

func TestAsyncWriterResetBeforeStartStopsBindings(t *testing.T) {
	writer := NewAsyncWriter()
	t.Cleanup(writer.Stop)
	removed, replacement := &asyncLifecycleSink{}, &asyncLifecycleSink{}
	writer.AddWriter(removed)
	writer.ResetWriter()
	removed.check(t, 1, 1)
	writer.AddWriter(replacement)
	writer.Start()
	writeAsyncMessage(t, writer, "replacement")
	writer.Stop()
	removed.check(t, 1, 1)
	replacement.check(t, 1, 1, "replacement")
}

func TestAsyncWriterNestedRestart(t *testing.T) {
	outer, inner := NewAsyncWriter(), NewAsyncWriter()
	t.Cleanup(outer.Stop)
	sink := &asyncLifecycleSink{}
	inner.AddWriter(sink)
	outer.AddWriter(inner)
	for cycle := range 2 {
		outer.Start()
		writeAsyncMessage(t, outer, fmt.Sprintf("nested %d", cycle))
		outer.Stop()
	}
	sink.check(t, 2, 2, "nested 0", "nested 1")
}

func TestAsyncWriterFileRestartClosesAndReopens(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "async.log")
	file := NewFileWriter(func(o *FileWriterOption) { o.Filename = filename })
	raw := file.(*syncWriter).ref.(*bytesWriter).ref.(*fileWriter)
	writer := NewAsyncWriter()
	t.Cleanup(writer.Stop)
	writer.AddWriter(file)
	firstFile := raw.file
	for cycle := range 3 {
		writer.Start()
		if cycle == 0 && raw.file != firstFile {
			t.Fatal("first Start reopened an already started file writer")
		}
		if raw.file == nil {
			t.Fatal("Start did not open the downstream file")
		}
		opened := raw.file
		writer.Start()
		if raw.file != opened {
			t.Fatal("repeated Start replaced the downstream file")
		}
		writeAsyncMessage(t, writer, fmt.Sprintf("file cycle %d", cycle))
		writer.Stop()
		if raw.file != nil {
			t.Fatal("Stop did not close the downstream file")
		}
		if _, err := opened.Stat(); !errors.Is(err, os.ErrClosed) {
			t.Fatalf("previous file remains open: %v", err)
		}
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	for cycle := range 3 {
		message := fmt.Sprintf("file cycle %d", cycle)
		if strings.Count(string(data), message) != 1 {
			t.Fatalf("message %q missing or duplicated in %q", message, data)
		}
	}
}

func TestAsyncWriterConcurrentRestart(t *testing.T) {
	writer := NewAsyncWriter()
	sink := &asyncLifecycleSink{}
	writer.AddWriter(sink)
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 20 {
				writer.Start()
				event := NewLogEvent()
				event.appendMessage("concurrent lifecycle")
				_ = writer.DoWrite(event)
				event.Recycle()
				writer.Stop()
			}
		}()
	}
	done := make(chan struct{})
	go func() { workers.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("concurrent Start/Stop did not complete")
	}
	writer.Stop()
	sink.mu.Lock()
	defer sink.mu.Unlock()
	if sink.invalid || sink.running || sink.starts != sink.stops {
		t.Fatalf("invalid downstream lifecycle: starts=%d stops=%d running=%v invalid=%v",
			sink.starts, sink.stops, sink.running, sink.invalid)
	}
}
