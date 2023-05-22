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

* If you log with other logger, it will send to the bound logger:

```text
require (
	github.com/coolerfall/lork/bind/zap latest
)
```

```go
lork.Install(lork.NewLogBridge())
lork.Load(zap.NewZapProvider())

zap.L().With().Warn("this is zap")
log.Printf("this is builtin logger")
```

Note: only **global** logger will send log to bound logger if using logger like zap, zerolog, logrus or other loggers.

For more details, see [configuration](/configuration)