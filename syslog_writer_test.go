//go:build !windows && !plan9

package lork

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func newSyslogWriterTest(t *testing.T) (*syslogWriter, net.PacketConn) {
	t.Helper()
	server, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	writer := NewSyslogWriter(func(o *SyslogWriterOption) {
		o.Network, o.Address, o.Tag = "udp", server.LocalAddr().String(), "lork-review"
	}).(*eventWriter).ref.(*syslogWriter)
	t.Cleanup(func() {
		writer.Stop()
		_ = server.Close()
	})
	return writer, server
}

func TestSyslogWriterUDPLevelsAndRestart(t *testing.T) {
	writer, server := newSyslogWriterTest(t)
	writer.Stop()
	if err := writer.Write(nil); err != errSyslogNotStarted {
		t.Fatalf("Write before Start = %v", err)
	}
	for range 2 {
		writer.Start()
		if writer.sw == nil {
			t.Fatal("valid UDP syslog configuration did not start")
		}
		connection := writer.sw
		writer.Start()
		if writer.sw != connection {
			t.Fatal("repeated Start replaced the connection")
		}
		for _, tc := range []struct {
			level    Level
			priority int
		}{
			{TraceLevel, 7}, {DebugLevel, 7}, {InfoLevel, 6}, {WarnLevel, 4},
			{ErrorLevel, 3}, {FatalLevel, 2}, {PanicLevel, 0},
		} {
			event := NewLogEvent()
			event.appendLevel(tc.level)
			event.appendMessage("syslog payload")
			err := writer.Write(event)
			event.Recycle()
			if err != nil {
				t.Fatal(err)
			}
			if err := server.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
				t.Fatal(err)
			}
			var packet [1024]byte
			n, _, err := server.ReadFrom(packet[:])
			if err != nil {
				t.Fatalf("level %v was not delivered: %v", tc.level, err)
			}
			message := string(packet[:n])
			if !strings.HasPrefix(message, fmt.Sprintf("<%d>", tc.priority)) || !strings.Contains(message, "syslog payload") {
				t.Fatalf("unexpected syslog packet: %q", message)
			}
		}
		writer.Stop()
		writer.Stop()
		if err := writer.Write(nil); err != errSyslogNotStarted {
			t.Fatalf("Write after Stop = %v", err)
		}
	}
}

func TestSyslogWriterFailedStartCanRetry(t *testing.T) {
	writer, _ := newSyslogWriterTest(t)
	writer.opts.Network = "invalid-network"
	writer.Start()
	if writer.sw != nil || writer.startErr == nil {
		t.Fatal("failed startup did not leave a retryable error")
	}
	if err := writer.Write(nil); err != writer.startErr {
		t.Fatalf("Write after startup failure = %v, want %v", err, writer.startErr)
	}
	writer.opts.Network = "udp"
	writer.Start()
	if writer.sw == nil || writer.startErr != nil {
		t.Fatal("writer did not recover after valid configuration")
	}
}

func TestSyslogWriterConcurrentLifecycle(t *testing.T) {
	writer, _ := newSyslogWriterTest(t)
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for range 20 {
				writer.Start()
				event := NewLogEvent()
				event.appendLevel(DebugLevel)
				event.appendMessage("concurrent syslog message")
				_ = writer.Write(event)
				event.Recycle()
				writer.Stop()
			}
		}()
	}
	done := make(chan struct{})
	go func() { workers.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("concurrent syslog lifecycle did not complete")
	}
}
