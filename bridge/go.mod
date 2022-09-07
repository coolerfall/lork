module github.com/coolerfall/slago/bridge

go 1.16

replace github.com/coolerfall/slago => ../

require (
	github.com/coolerfall/slago v0.5.5
	github.com/rs/zerolog v1.26.1
	github.com/sirupsen/logrus v1.8.1
	go.uber.org/zap v1.21.0
)
