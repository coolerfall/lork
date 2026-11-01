## Writer

Lork provides several writers for logging, and it supports to add multiple writers.

### Console Writer

This writer sends the logs to `Stdout` console. It supports the following options:

* `Encoder`, encoder of logs
* `Filter`, filter of logs

```go
cw := lork.NewConsoleWriter(func(o *lork.ConsoleWriterOption) {
		o.Name = "CONSOLE"
		o.Encoder = lork.NewPatternEncoder(func(opt *lork.PatternEncoderOption) {
			opt.Pattern = "#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
				"#level) #color([#logger{36}]){magenta} : #message #fields"
		})
	})
lork.Manual().AddWriter(cw)
```

### File Writer

It supports the following options:

* `Encoder`, encoder of logs
* `Filter`, filter of logs
* `Filename`, the filename of the log file to write
* `RollingPolicy`, the policy to roll over log files. Lork provides`TimeBasedPpolicy` and
  `SizeAndTimeBasedPolicy`

```go
fw := lork.NewFileWriter(func(o *lork.FileWriterOption) {
    o.Name = "FILE"
    o.Encoder = lork.NewJsonEncoder()
    o.Filter = lork.NewThresholdFilter(lork.InfoLevel)
    o.Filename = "/tmp/lork/lork-test.log"
    o.RollingPolicy = lork.NewSizeAndTimeBasedRollingPolicy(
        func(o *lork.SizeAndTimeBasedRPOption) {
            o.FilenamePattern = "/tmp/lork/lork-archive.#date{2006-01-02}.#index.log"
            o.MaxFileSize = "10MB"
            o.MaxHistory = 10
        })
})
lork.Manual().AddWriter(fw)
```

### Asynchronous Writer

This writer wraps `Console Writer` or `File Writer` to write log in background. It
supports the following options:

* `QueueSize`, the size of the blocking queue.

```go
aw := lork.NewAsyncWriter(func(o *lork.AsyncWriterOption) {
    o.Name = "ASYNC-FILE"
})
aw.AddWriter(fw)

lork.Manual().AddWriter(aw)
```

`Stop()` drains pending events and stops downstream writers without removing their
bindings. `Start()` restarts those writers before resuming background output.
Writers added after `Stop()` start on the next `Start()`. Call `ResetWriter()` to
remove the downstream writers explicitly. Configure bindings before the first
`Start()` or while the asynchronous writer is stopped.

### Socket Writer

This writer sends logs to remote server via socket. It supports the following options:

* `RemoteUrl`, url of remote server
* `QueueSize`, the size of queue
* `ReconnectionDelay`, delay milliseconds when reconnecting server
* `Filter`, filters of logs

The server should start `Socket Reader`to receive logs, and it supports the following
options:

* `Path`, the path of the url
* `Port`, the port of this server will listen

### Syslog Writer

This writer is an implementation for syslog. It supports the following options:

* `Tag`, the tag of syslog.
* `Address`, address of syslog server
* `Network`, network of syslog server, see `net.Dial`
* `Filter`, filters of logs

## Encoder

Lork provides some builtin encoders which can be configured in writers.

Each JSON or Pattern encoder instance must belong to a single writer. These encoders
reuse their output buffers; sharing an instance across writers is unsupported and
can overwrite log contents during concurrent writes. The default console, file and
socket writers each create an independent encoder.

When configuring multiple writers explicitly, create a new encoder for each one,
even when they use the same format:

```go
first := lork.NewFileWriter(func(o *lork.FileWriterOption) {
    o.Filename = "first.log"
    o.Encoder = lork.NewJsonEncoder()
})
second := lork.NewFileWriter(func(o *lork.FileWriterOption) {
    o.Filename = "second.log"
    o.Encoder = lork.NewJsonEncoder()
})
lork.Manual().AddWriter(first, second)
```

### Pattern Encoder

Encode logs with custom pattern format layout, for example:

```text
#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(#level) #color([#logger{16}]){magenta} : #message #fields
```

#### color

This pattern adds specified color the content. This only works in console writer.

```text
#color(theContent){colorValue}
```

* Normal colors : `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`
* Bright colors : `blackbr`, `redbr`, `greenbr`, `yellowbr`, `bluebr`,
  `magentabr`, `cyanbr`, `whitebr`

#### level

This pattern adds level information in logs.

```text
#level
```

#### logger

This pattern adds logger name in logs.

```text
#logger{length}
```

#### message

This pattern adds message in logs.

```text
#message
```

#### fields

This pattern adds key-value fields in logs.

```text
#fields
```

#### custom

You can add your own pattern keyword and add convert options in `PatternEncoder`. 

### Json Encoder

Encode logs with json format.

## Filter

Filters can filter unused logs from origin logs. Lork provides some built in filters.

### Threshold Filter

This filter will deny all logs which is lower than the level set.

### Keyword Filter

A simple keyword filter which matches the specified keyword.

## Provider

Lork provides providers which will output log finally. A default provider is
provided if no provider set, and it's more efficient to use the default one. An
alternative provider can be set like `zerolog` provider:

```go
lork.Load(zero.NewZeroProvider())
```

## Bridge

Lork can take over logs from other logger like `zerolog`, and send logs to the provider
to output:

```go
lork.Install(lork.NewLogBridge())
lork.Load(zap.NewZapProvider())

zap.L().With().Warn("this is zap")
log.Printf("this is builtin logger")

```

!> Note: only **global** logger will send log to bound logger if using logger like zap,
zerolog, logrus or other loggers.  

