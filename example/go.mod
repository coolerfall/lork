module github.com/coolerfall/slago/example

go 1.16

replace github.com/coolerfall/slago => ../

replace github.com/coolerfall/slago/binder/logrus => ../binder/logrus

replace github.com/coolerfall/slago/binder/zap => ../binder/zap

replace github.com/coolerfall/slago/binder/zero => ../binder/zero

replace github.com/coolerfall/slago/bridge => ../bridge

require (
	github.com/coolerfall/slago v0.5.5
	github.com/coolerfall/slago/binder/logrus v0.5.5
	github.com/coolerfall/slago/binder/zap v0.5.5
	github.com/coolerfall/slago/binder/zero v0.5.5
	github.com/coolerfall/slago/bridge v0.5.5
	github.com/sirupsen/logrus v1.8.1
	go.uber.org/zap v1.21.0
)
