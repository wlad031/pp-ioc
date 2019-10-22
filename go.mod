module github.com/wlad031/pp-ioc

go 1.13

require (
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/wlad031/pp-algo v0.0.1
	github.com/wlad031/pp-logging v0.0.4
	github.com/wlad031/pp-properties v0.0.1
)

replace github.com/wlad031/pp-algo => ../pp-algo

replace github.com/wlad031/pp-logging => ../pp-logging

replace github.com/wlad031/pp-properties => ../pp-properties
