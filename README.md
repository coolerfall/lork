slago
=====
Simple Logging Abstraction for Go. Slago provides bridge and binder for logger which
can sent log from logger to another logger you preferred. Slago also provides unified writers, 
encoders and filters, it brings different logger with same apis and flexible configurations.

Install
=======
Add the following to your go.mod
```text
require (
	github.com/coolerfall/slago v0.5.0
)
```

Quick Start
==========
* Add logger you want to bind to:
```go
slago.Bind(salzero.NewZeroLogger())
```

* Install the bridges for other logger :
```go
slago.Install(bridge.NewLogBridge())
slago.Install(bridge.NewLogrusBridge())
slago.Install(bridge.NewZapBrige())
```

* Configure the output writer:
```go
cw := slago.NewConsoleWriter(func(o *slago.ConsoleWriterOption) {
		o.Encoder = slago.NewPatternEncoder(
			"#color(#date{2006-01-02T15:04:05.000Z07:00}){cyan} #color(#level) #message #fields")
	})
slago.Logger().AddWriter(cw)
fw := slago.NewFileWriter(func(o *slago.FileWriterOption) {
		o.Encoder = slago.NewJsonEncoder()
		o.Filter = slago.NewLevelFilter(slago.TraceLevel)
		o.Filename = "slago-test.log"
		o.RollingPolicy = slago.NewSizeAndTimeBasedRollingPolicy(
			func(o *slago.SizeAndTimeBasedRPOption) {
				o.FilenamePattern = "slago-archive.#date{2006-01-02}.#index.log"
				o.MaxFileSize = "10MB"
                o.MaxHistory = 10
			})
	})
slago.Logger().AddWriter(fw)
```

* Add logging:
```go
slago.Logger().Trace().Msg("slago")
slago.Logger().Info().Int("int", 88).Interface("slago", "val").Msg("")
```

* If you log with other logger, it will send to the bound logger:
```go
zap.L().With().Warn("this is zap")
log.Printf("this is builtin logger")
```
Note: only **global** logger will send log to bound logger if using logger like zap, zerolog, logrus or other loggers.  

Configuration
============
The following shows all the configurations of slago.

# Writer
Slago provides several writers for logging, and it supports to add multiple writers.

### Console Writer
This writer sends the logs to `Stdout` console. It supports the following options:
* `Encoder`, encoder of logs
* `Filter`, filter of logs

### File Writer
It supports the following options:
* `Encoder`, encoder of logs
* `Filter`, filter of logs
* `Filename`, the filename of the log file to write
* `RollingPolicy`, the policy to roll over log files. Slago provides`TimeBasedPpolicy` and `SizeAndTimeBasedPolicy`

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
Slago provides some builtin encoders which can be configured in wirters.

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
Filters can filter unused logs from origin logs. Slago provides some built in filters.

### Level Filter
This filter will filter all logs which is  lower than the level set.

### Keyword Filter
A simple keyword filter which matches the specified keyword.

Credits
======
[slf4j][1]: Simple Logging Facade for Java
[logback][2]: The reliable, generic, fast and flexible logging framework for Java.


License
=======

    Copyright (c) 2019-2020 Vincent Cheung (coolingfall@gmail.com).
    
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