module gitlab.com/anbillon/slago/example

go 1.12

require (
	github.com/json-iterator/go v1.1.7
	github.com/rs/zerolog v1.17.2
	github.com/sirupsen/logrus v1.4.2
	gitlab.com/anbillon/slago/log-to-slago v0.0.0-00010101000000-000000000000
	gitlab.com/anbillon/slago/logrus-to-slago v0.0.0-00010101000000-000000000000
	gitlab.com/anbillon/slago/slago-api v0.1.0
	gitlab.com/anbillon/slago/slago-zerolog v0.0.0-00010101000000-000000000000
	gitlab.com/anbillon/slago/zap-to-slago v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.13.0
)

replace (
	gitlab.com/anbillon/slago/log-to-slago => ../log-to-slago
	gitlab.com/anbillon/slago/logrus-to-slago => ../logrus-to-slago
	gitlab.com/anbillon/slago/slago-api => ../slago-api
	gitlab.com/anbillon/slago/slago-logrus => ../slago-logrus
	gitlab.com/anbillon/slago/slago-zap => ../slago-zap
	gitlab.com/anbillon/slago/slago-zerolog => ../slago-zerolog
	gitlab.com/anbillon/slago/zap-to-slago => ../zap-to-slago
	gitlab.com/anbillon/slago/zerolog-to-slago => ../zerolog-to-slago
)
