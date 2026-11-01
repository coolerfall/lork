package lork

import (
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingTestProvider struct {
	Provider
	prepares  atomic.Int32
	preparing func()
}

func (p *countingTestProvider) Prepare() {
	p.prepares.Add(1)
	if p.preparing != nil {
		p.preparing()
	}
	p.Provider.Prepare()
}

func TestFactoryInitializesOnce(t *testing.T) {
	_ = LoggerC()
	provider := &countingTestProvider{Provider: NewClassicProvider()}
	f := &loggerFactory{substProvider: newSubstituteProvider(), providers: []Provider{provider}}
	provider.preparing = func() {
		if f.provider() != f.substProvider {
			t.Error("reentrant initialization did not use the substitute provider")
		}
	}
	f.lock.Lock()
	var workers sync.WaitGroup
	const count = 16
	for i := 0; i < count; i++ {
		workers.Add(1)
		go func() { defer workers.Done(); f.provider() }()
	}
	// Hold initialization until every caller has entered the provider method.
	deadline := time.Now().Add(2 * time.Second)
	for {
		stack := make([]byte, 256*1024)
		n := runtime.Stack(stack, true)
		if strings.Count(string(stack[:n]), "(*loggerFactory).provider(") >= count {
			break
		}
		if time.Now().After(deadline) {
			f.lock.Unlock()
			workers.Wait()
			t.Fatal("callers did not reach the initialization lock")
		}
		runtime.Gosched()
	}
	f.lock.Unlock()
	workers.Wait()
	if got := provider.prepares.Load(); got != 1 {
		t.Fatalf("Prepare called %d times, want 1", got)
	}
	ctx := provider.Provider.(*classicProvider).context
	defer ctx.rootLogger.ResetWriter()
	if got := ctx.rootLogger.multiWriter.Size(); got != 1 {
		t.Fatalf("installed %d writers, want 1", got)
	}
	if f.provider() != provider {
		t.Fatal("initialized provider was not published")
	}
}

type handoffTestFactory struct {
	lock     sync.Mutex
	messages []string
}

func (f *handoffTestFactory) Logger(name string) ILogger {
	return newEventCacheLogger(name, f)
}

func (f *handoffTestFactory) WriteEvent(event *LogEvent) error {
	f.lock.Lock()
	f.messages = append(f.messages, string(event.Message()))
	f.lock.Unlock()
	event.Recycle()
	return nil
}

type handoffTestProvider struct {
	factory   *handoffTestFactory
	preparing func()
}

func (p *handoffTestProvider) Name() string { return "handoff-test" }
func (p *handoffTestProvider) Prepare() {
	if p.preparing != nil {
		p.preparing()
	}
}
func (p *handoffTestProvider) LoggerFactory() ILoggerFactory { return p.factory }

func TestFactoryLateInitializationRecord(t *testing.T) {
	_ = LoggerC()
	output := &handoffTestFactory{}
	provider := &handoffTestProvider{factory: output}
	f := &loggerFactory{substProvider: newSubstituteProvider(), providers: []Provider{provider}}
	var pending Record
	var startupFactory ILoggerFactory
	provider.preparing = func() {
		startupFactory = f.ILoggerFactory()
		logger := startupFactory.Logger("startup")
		logger.Info().Msg("queued")
		pending = logger.Info()
	}
	f.provider()
	pending.Msg("completed after initialization")
	startupFactory.Logger("late logger").Info().Msg("created after initialization")
	if got := len(output.messages); got != 3 {
		t.Fatalf("delivered %d events, want 3", got)
	}
	if got := f.substProvider.factory.eventQueue.Len(); got != 0 {
		t.Fatalf("%d events remained in the startup queue", got)
	}
}

func TestFactoryConcurrentInitializationHandoff(t *testing.T) {
	_ = LoggerC()
	output := &handoffTestFactory{}
	preparing, release := make(chan struct{}), make(chan struct{})
	provider := &handoffTestProvider{factory: output, preparing: func() {
		close(preparing)
		<-release
	}}
	f := &loggerFactory{substProvider: newSubstituteProvider(), providers: []Provider{provider}}
	initialized := make(chan struct{})
	go func() { f.provider(); close(initialized) }()
	<-preparing
	startupFactory := f.ILoggerFactory()
	var ready, workers sync.WaitGroup
	const count, records = 8, 16
	ready.Add(count)
	workers.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer workers.Done()
			logger := startupFactory.Logger("shared")
			pending := logger.Info()
			ready.Done()
			<-release
			pending.Msg("pending")
			for j := 0; j < records; j++ {
				startupFactory.Logger("shared").Info().Msg("concurrent")
			}
		}()
	}
	ready.Wait()
	close(release)
	workers.Wait()
	<-initialized
	if got := len(output.messages); got != count*(records+1) {
		t.Fatalf("delivered %d events, want %d", got, count*(records+1))
	}
	if got := f.substProvider.factory.eventQueue.Len(); got != 0 {
		t.Fatalf("%d events remained in the startup queue", got)
	}
}

func TestFactoryFullInitializationQueue(t *testing.T) {
	_ = LoggerC()
	output := &handoffTestFactory{}
	provider := &handoffTestProvider{factory: output}
	f := &loggerFactory{substProvider: newSubstituteProvider(), providers: []Provider{provider}}
	provider.preparing = func() {
		logger := f.Logger("startup")
		for i := 0; i < DefaultQueueSize+8; i++ {
			logger.Info().Msg("startup")
		}
	}
	done := make(chan struct{})
	go func() { f.provider(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("full startup queue blocked initialization")
	}
	if got := len(output.messages); got != DefaultQueueSize {
		t.Fatalf("delivered %d events, want %d", got, DefaultQueueSize)
	}
}

func TestFactoryInitializationAfterReset(t *testing.T) {
	_ = LoggerC()
	output := &handoffTestFactory{}
	provider := &handoffTestProvider{factory: output}
	f := &loggerFactory{substProvider: newSubstituteProvider(), providers: []Provider{provider}}
	provider.preparing = func() {
		before := len(output.messages)
		f.Logger("startup").Info().Msg("startup")
		if len(output.messages) != before {
			t.Error("event reached the provider before Prepare completed")
		}
	}
	f.provider()
	f.Reset()
	f.provider()
	if got := len(output.messages); got != 2 {
		t.Fatalf("delivered %d events across reset, want 2", got)
	}
}
