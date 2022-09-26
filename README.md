lork
=====
The flexible, extensible and structured logging for Go. Lork provides bridge and binder for logger 
which can send log from logger to another logger you preferred. Lork also provides unified writers, 
encoders and filters, it brings different logger with same apis and flexible configurations.

Install
=======
Add the following to your go.mod
```text
require (
	github.com/coolerfall/lork v0.5.4
)
```

Quick Start
==========
* Configure the output writer:
```go
lork.Logger().AddWriter(lork.NewConsoleWriter(func(o *lork.ConsoleWriterOption) {
    o.Encoder = lork.NewPatternEncoder(func(opt *lork.PatternEncoderOption) {
        opt.Pattern = "#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(" +
"#level) #color([#logger{32}]){magenta} : #message #fields"
    })
}))
fw := lork.NewFileWriter(func(o *lork.FileWriterOption) {
    o.Encoder = lork.NewJsonEncoder()
    o.Filter = lork.NewThresholdFilter(lork.DebugLevel)
    o.Filename = "example/lork-test.log"
    o.RollingPolicy = lork.NewSizeAndTimeBasedRollingPolicy(
        func(o *lork.SizeAndTimeBasedRPOption) {
            o.FilenamePattern = "example/lork-archive.#date{2006-01-02}.#index.log"
            o.MaxFileSize = "10MB"
            o.MaxHistory = 10
    })
})
aw := lork.NewAsyncWriter(func(o *lork.AsyncWriterOption) {
    o.Ref = fw
})
lork.Logger().AddWriter(aw)

```

* Add logging:
```go
lork.Logger().Trace().Msg("lork")
lork.Logger().Info().Int("int", 88).Any("lork", "val").Msge()
```

* If you log with other logger, it will send to the bound logger:
```go
lork.Install(lork.NewLogBridge())
lork.Install(bridge.NewZapBrige())

zap.L().With().Warn("this is zap")
log.Printf("this is builtin logger")
```
Note: only **global** logger will send log to bound logger if using logger like zap, zerolog, logrus or other loggers.  

Configuration
============
The following shows all the configurations of lork.

# Writer
lork provides several writers for logging, and it supports to add multiple writers.

### Console Writer
This writer sends the logs to `Stdout` console. It supports the following options:
* `Encoder`, encoder of logs
* `Filter`, filter of logs

### File Writer
It supports the following options:
* `Encoder`, encoder of logs
* `Filter`, filter of logs
* `Filename`, the filename of the log file to write
* `RollingPolicy`, the policy to roll over log files. lork provides`TimeBasedPpolicy` and `SizeAndTimeBasedPolicy`

### Asynchronous Writer
This writer wraps `Console Writer` or `File Writer` to write log in background. It supports the following options: 
* `Ref`, the referenced writer.
* `QueueSize`, the size of the blocking queue.

### Socket Writer
This writer sends logs to remote server via socket. It supports the following options:
* `RemoteUrl`, url of remote server
* `QueueSize`, the size of queue
* `ReconnectionDelay`, delay milliseconds when reconnecting server
* `Filter`, filters of logs

The server should start `Socket Reader`to receive logs, and it supports the following options:
* `Path`, the path of the url
* `Port`, the port of this server will listen

## Encoder
lork provides some builtin encoders which can be configured in wirters.

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
Normal colors supported: `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`
Bright colors supported: `blackbr`, `redbr`, `greenbr`, `yellowbr`, `bluebr`, `magentabr`, `cyanbr`, `whitebr`

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
#### fileds
This pattern adds key-value fields in logs.
```text
#fields
```

### Json Encoder
Encode logs with json format.

## Filter
Filters can filter unused logs from origin logs. lork provides some built in filters.

### Threshold Filter
This filter will deny all logs which is lower than the level set.

### Keyword Filter
A simple keyword filter which matches the specified keyword.

Benchmarks
==========
Benchmarks with complex log field, diferent encoder and writer.
```text
cpu: Intel(R) Core(TM) i7-8550U CPU @ 1.80GHz
BenchmarkJsonFileWirter-8      	  250128        5290 ns/op      0 B/op      0 allocs/op
BenchmarkPatternFileWirter-8   	  313402        3777 ns/op      0 B/op      0 allocs/op
BenchmarkAsyncFileWirter-8     	 1107603        1060 ns/op      0 B/op      0 allocs/op
BenchmarkNoWirter-8            	 1441761        843.5 ns/op     0 B/op      0 allocs/op
```

Credits
======
* [slf4j][1]: Simple Logging Facade for Java
* [logback][2]: The reliable, generic, fast and flexible logging framework for Java.

Supports
=======
If you enjoy this project and want to support it, you can buy a coffee.

<a href="https://www.buymeacoffee.com/coolerfall" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/v2/default-yellow.png" alt="Buy Me A Coffee" style="height: 60px !important;width: 217px !important;" ></a>


License
=======

    Copyright (c) 2019-2022 Vincent Cheung (coolingfall@gmail.com).
    
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
    
         http://www.apache.org/licenses/LICENSE-2.0
    
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.


[1]: https://github.com/qos-ch/slf4j
[2]: https://github.com/qos-ch/logback
