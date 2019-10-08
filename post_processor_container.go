package pp_ioc

import (
    log "github.com/sirupsen/logrus"
    logCtx "github.com/wlad031/pp-logging"
)

type postProcessorContainer struct {
    logger *logCtx.NamedLogger
    ls     []PostProcessor
}

func newPostProcessorContainer() *postProcessorContainer {
    return &postProcessorContainer{
        logger: logCtx.Get("IOC.PostProcessorContainer"),
        ls:     []PostProcessor{},
    }
}

func (container *postProcessorContainer) iterate() <-chan PostProcessor {
    c := make(chan PostProcessor)
    go func() {
        for _, v := range container.ls {
            c <- v
        }
        close(c)
    }()
    return c
}

func (container *postProcessorContainer) add(b *bean) error {
    postProcessor := b.instance.(PostProcessor)
    container.ls = append(container.ls, postProcessor)
    container.logger.WithFields(log.Fields{
        "beanDef": b.definition.shortString(),
    }).Info("Post processor added")
    return nil
}
