module gitlab.com/anbillon/slago/slago-zap

go 1.12

require (
	github.com/pkg/errors v0.8.1 // indirect
	gitlab.com/anbillon/slago/slago-api v0.1.0
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
)

replace gitlab.com/anbillon/slago/slago-api => ../slago-api
