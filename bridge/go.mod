module github.com/coolerfall/lork/bridge

go 1.16

replace github.com/coolerfall/lork => ../

require (
	github.com/coolerfall/lork v0.6.0
	github.com/rs/zerolog v1.26.1
	github.com/sirupsen/logrus v1.8.1
	go.uber.org/zap v1.21.0
)
