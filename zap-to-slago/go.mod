module gitlab.com/anbillon/slago/zap-to-slago

go 1.12

require (
	gitlab.com/anbillon/slago/slago-api v0.0.0-00010101000000-000000000000
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
)

replace gitlab.com/anbillon/slago/slago-api => ../slago-api
