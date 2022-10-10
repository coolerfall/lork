module github.com/coolerfall/lork/example

go 1.16

replace github.com/coolerfall/lork => ../

replace github.com/coolerfall/lork/binder/logrus => ../binder/logrus

replace github.com/coolerfall/lork/binder/zap => ../binder/zap

replace github.com/coolerfall/lork/binder/zero => ../binder/zero

replace github.com/coolerfall/lork/bridge => ../bridge

require (
	github.com/coolerfall/lork v0.6.0
	github.com/coolerfall/lork/binder/logrus v0.6.0
	github.com/coolerfall/lork/binder/zap v0.6.0
	github.com/coolerfall/lork/binder/zero v0.6.0
	github.com/coolerfall/lork/bridge v0.6.0
	github.com/rs/zerolog v1.26.1
	github.com/sirupsen/logrus v1.8.1
	go.uber.org/zap v1.21.0
)
