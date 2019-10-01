package pp_ioc

import "reflect"

type Context interface {
    Bind(type_ reflect.Type, o interface{})
    Get(type_ reflect.Type) interface{}
}

func NewContext() Context {
    return &contextImpl{
        container: map[reflect.Type]interface{}{},
    }
}

type contextImpl struct {
    container map[reflect.Type]interface{}
}

func (c *contextImpl) Bind(type_ reflect.Type, o interface{}) {
    c.container[type_] = o
}

func (c contextImpl) Get(type_ reflect.Type) interface{} {
    return c.container[type_]
}

