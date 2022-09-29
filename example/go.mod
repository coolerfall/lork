module github.com/coolerfall/lork/example

go 1.16

replace github.com/coolerfall/lork => ../

replace github.com/coolerfall/lork/bind/logrus => ../bind/logrus

replace github.com/coolerfall/lork/bind/zap => ../bind/zap

replace github.com/coolerfall/lork/bind/zero => ./../bind/zero

replace github.com/coolerfall/lork/bridge => ../bridge

require (
	github.com/coolerfall/lork v0.6.0
	github.com/coolerfall/lork/bind/logrus v0.6.0
	github.com/coolerfall/lork/bind/zap v0.6.0
	github.com/coolerfall/lork/bind/zero v0.6.0
	github.com/coolerfall/lork/bridge v0.6.0
	github.com/rs/zerolog v1.26.1
	github.com/sirupsen/logrus v1.8.1
	go.uber.org/zap v1.21.0
)
