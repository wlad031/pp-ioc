package pp_ioc

import (
    log "github.com/sirupsen/logrus"
    logCtx "github.com/wlad031/pp-logging"
)

type beanContainer struct {
    logger *logCtx.NamedLogger
    ls     []*bean
}

func newBeanContainer() *beanContainer {
    return &beanContainer{
        logger: logCtx.Get("IOC.BeanContainer"),
        ls:     []*bean{},
    }
}

func (container *beanContainer) iterate() <-chan *bean {
    c := make(chan *bean)
    go func() {
        for _, v := range container.ls {
            c <- v
        }
        close(c)
    }()
    return c
}

func (container *beanContainer) add(b *bean) {
    container.ls = append(container.ls, b)
    container.logger.WithFields(log.Fields{
        "beanDef": b.definition.String(),
    }).Info("Bean added")
}
