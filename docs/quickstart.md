# Quick Start

* Add the following to your `go.mod`

```go.mod
require (
	github.com/coolerfall/lork latest
)
```

* Now logging in your project:

```go
lork.Logger().Info().Msg("Hello lork")
lork.LoggerC().Info().Msg("This message will show caller")

logger := lork.Logger("github.com/lork")
logger.Debug().Msg("lork sub logger")
logger.SetLevel(lork.InfoLevel)
logger.Trace().Msg("this will not print")
logger.Info().Any("any", map[string]interface{}{
    "name": "dog",
    "age":  2,
}).Msg("this is interface")
```

This will log with default console writer with pattern format.

* Add writter with rolling policy to output your logs:

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

For more details, see [configuration](/configuration)