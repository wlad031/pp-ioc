package pp_ioc

import (
    logCtx "github.com/wlad031/pp-logging"
)

type binderContainer struct {
    logger *logCtx.NamedLogger
    ls     []*Binder
}

func newBinderContainer() *binderContainer {
    return &binderContainer{
        logger: logCtx.Get("IOC.BinderContainer"),
        ls:     []*Binder{},
    }
}

func (container *binderContainer) iterate() <-chan *Binder {
    c := make(chan *Binder)
    go func() {
        for _, v := range container.ls {
            c <- v
        }
        close(c)
    }()
    return c
}

func (container *binderContainer) add(binders ...*Binder) error {
    container.ls = append(container.ls, binders...)
    return nil
}
