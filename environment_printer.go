package pp_ioc

import (
    log "github.com/sirupsen/logrus"
    logCtx "github.com/wlad031/pp-logging"
)

type EnvironmentPrinter interface {
    PostProcessor
}

func NewEnvironmentPrinter() EnvironmentPrinter {
    return &environmentPrinterImpl{logger: logCtx.Get("EnvironmentPrinter")}
}

type environmentPrinterImpl struct {
    logger logCtx.NamedLogger
}

func (ep *environmentPrinterImpl) PostProcess(ctx Context) error {
    properties := ctx.GetEnvironment().GetAllProperties()
    for key, value := range properties {
        ep.logger.WithFields(log.Fields{
            "key": key,
            "value": value,
        }).Info("Found property")
    }
    return nil
}
