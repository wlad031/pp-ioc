package pp_ioc

import (
    "reflect"
    "strings"
)

type bindKey struct {
    qualifiers []string
    type_      reflect.Type
}

func (b *bindKey) String() string {
    return "Key{[" + strings.Join(b.qualifiers, ",") + "]:" + b.type_.String() + "}"
}
