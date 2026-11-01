package lork

import (
	"fmt"
	"slices"
	"sync"
	"testing"
)

type namedLevelSink struct {
	mu       sync.Mutex
	messages []string
}

func (*namedLevelSink) Name() string { return "level-sink" }

func (s *namedLevelSink) DoWrite(event *LogEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, string(event.Message()))
	return nil
}

func checkNamedLoggerThreshold(t *testing.T, logger *namedLogger, sink *namedLevelSink, threshold Level) {
	t.Helper()
	sink.mu.Lock()
	sink.messages = nil
	sink.mu.Unlock()
	methods := []func() Record{
		logger.Trace, logger.Debug, logger.Info, logger.Warn,
		logger.Error, logger.Fatal, logger.Panic,
	}
	var expected []string
	for level := TraceLevel; level <= PanicLevel; level++ {
		methodMessage, genericMessage := "method:"+level.String(), "generic:"+level.String()
		methods[level]().Msg(methodMessage)
		logger.Level(level).Msg(genericMessage)
		if level >= threshold {
			expected = append(expected, methodMessage, genericMessage)
		}
	}
	sink.mu.Lock()
	defer sink.mu.Unlock()
	if !slices.Equal(sink.messages, expected) {
		t.Fatalf("logger %q with threshold %v: got %q, want %q", logger.Name(), threshold, sink.messages, expected)
	}
}

func TestNamedLoggerLevelFiltering(t *testing.T) {
	ctx := NewLoggerContext(NewClassicLogger)
	sink := &namedLevelSink{}
	ctx.rootLogger.AddWriter(sink)
	logger := ctx.RealLogger("app")
	checkNamedLoggerThreshold(t, logger, sink, TraceLevel)
	for level := PanicLevel; level >= TraceLevel; level-- {
		logger.SetLevel(level)
		checkNamedLoggerThreshold(t, logger, sink, level)
	}
}

func TestNamedLoggerLevelInheritance(t *testing.T) {
	ctx := NewLoggerContext(NewClassicLogger)
	sink := &namedLevelSink{}
	ctx.rootLogger.AddWriter(sink)
	early := ctx.RealLogger("app/early")
	ctx.rootLogger.SetLevel(WarnLevel)
	late := ctx.RealLogger("app/late")
	checkNamedLoggerThreshold(t, early, sink, WarnLevel)
	checkNamedLoggerThreshold(t, late, sink, WarnLevel)

	early.SetLevel(ErrorLevel)
	// Setting the same parent level preserves an existing child override.
	ctx.rootLogger.SetLevel(WarnLevel)
	nested := ctx.RealLogger("app/early/nested")
	checkNamedLoggerThreshold(t, early, sink, ErrorLevel)
	checkNamedLoggerThreshold(t, nested, sink, ErrorLevel)
	checkNamedLoggerThreshold(t, late, sink, WarnLevel)

	ctx.rootLogger.SetLevel(InfoLevel)
	for _, logger := range []*namedLogger{ctx.rootLogger, early, late, nested} {
		checkNamedLoggerThreshold(t, logger, sink, InfoLevel)
	}
	if early.FindChild(nested.Name()) != nested || early.FindChild("missing") != nil {
		t.Fatal("child lookup changed after level propagation")
	}
}

func TestNamedLoggerConcurrentSetLevel(t *testing.T) {
	ctx := NewLoggerContext(NewClassicLogger)
	logger := ctx.RealLogger("app/worker")
	start := make(chan struct{})
	var workers sync.WaitGroup
	for worker := range 6 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for range 1000 {
				switch worker {
				case 0:
					ctx.rootLogger.SetLevel(InfoLevel)
					ctx.rootLogger.SetLevel(WarnLevel)
				case 1:
					logger.SetLevel(DebugLevel)
					logger.SetLevel(ErrorLevel)
				default:
					logger.Debug().Msg("debug")
					logger.Info().Msg("info")
					logger.Error().Msg("error")
				}
			}
		}()
	}
	close(start)
	workers.Wait()
	logger.SetLevel(WarnLevel)
	sink := &namedLevelSink{}
	logger.AddWriter(sink)
	checkNamedLoggerThreshold(t, logger, sink, WarnLevel)
}

func TestNamedLoggerConcurrentLevelInheritance(t *testing.T) {
	ctx := NewLoggerContext(NewClassicLogger)
	parent := ctx.RealLogger("app")
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		<-start
		for range 1000 {
			ctx.rootLogger.SetLevel(InfoLevel)
			ctx.rootLogger.SetLevel(WarnLevel)
		}
	}()
	for worker := range 4 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			branch := fmt.Sprintf("app/worker%d", worker)
			for index := range 50 {
				logger := ctx.RealLogger(fmt.Sprintf("%s/child%d", branch, index))
				logger.Info().Msg("new child")
				if parent.FindChild(branch) == nil {
					t.Error("new child branch was not visible")
				}
			}
		}()
	}
	close(start)
	workers.Wait()

	// A completed update must include all children published during creation.
	ctx.rootLogger.SetLevel(ErrorLevel)
	sink := &namedLevelSink{}
	ctx.rootLogger.AddWriter(sink)
	for worker := range 4 {
		for index := range 50 {
			logger := ctx.RealLogger(fmt.Sprintf("app/worker%d/child%d", worker, index))
			checkNamedLoggerThreshold(t, logger, sink, ErrorLevel)
		}
	}
	late := ctx.RealLogger("app/created-after-update")
	checkNamedLoggerThreshold(t, late, sink, ErrorLevel)
}
