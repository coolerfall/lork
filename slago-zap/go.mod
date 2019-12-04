module gitlab.com/anbillon/slago/slago-zap

go 1.12

require (
	gitlab.com/anbillon/slago/slago-api v0.1.0
	go.uber.org/zap v1.13.0
)

replace gitlab.com/anbillon/slago/slago-api => ../slago-api
