package pp_ioc

import (
    "errors"
    "strings"
)

type BeanScope int

const (
    ScopeUnknown   BeanScope = -1
    ScopeSingleton           = iota
    ScopePrototype
)

func FromString(s string) (BeanScope, error) {
    switch strings.ToLower(s) {
    case "singleton":
        return ScopeSingleton, nil
    case "prototype":
        return ScopePrototype, nil
    }
    return ScopeUnknown, errors.New("Cannot parse bean scope from string " + s)
}

func (bs BeanScope) String() string {
    switch bs {
    case ScopeSingleton:
        return "Singleton"
    case ScopePrototype:
        return "Prototype"
    }
    return "Unknown scope"
}
