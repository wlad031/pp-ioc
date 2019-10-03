package pp_ioc

import (
    "reflect"
    "testing"
)

//type testInterface interface {
//    f() string
//}
//
//type testImpl1 struct {
//}
//
//func (impl *testImpl1) f() string {
//    return "impl1"
//}
//
//type testImpl2 struct {
//}
//
//func (impl *testImpl2) f() string {
//    return "impl2"
//}
//
//type testImpl3 struct {
//}
//
//func (impl *testImpl3) f() string {
//    return "impl3"
//}

type o1 struct {}
type o2 struct {}
type o3 struct {}
type o4 struct {}
type o5 struct {}
type o6 struct {}

func newO1(ctx Context) (o1, bool) {
    return o1{}
}

func newO2(ctx Context) o2 {
    _, ok := ctx.Get(reflect.TypeOf((*o1)(nil)))
    return o2{}
}

func newO3(ctx Context) o3 {
    ctx.Get(reflect.TypeOf((*o2)(nil)))
    return o3{}
}

func newO4(ctx Context) o4 {
    ctx.Get(reflect.TypeOf((*o1)(nil)))
    return o4{}
}

func newO5(ctx Context) o5 {
    ctx.Get(reflect.TypeOf((*o6)(nil)))
    return o5{}
}

func newO6(ctx Context) o6 {
    ctx.Get(reflect.TypeOf((*o5)(nil)))
    return o6{}
}

func TestContext(t *testing.T) {
    ctx := NewContext()

    ctx.Bind(reflect.TypeOf((*o1)(nil)), func() (interface{}, bool) {
        newO1(ctx)
    })

    e := ctx.Build()
    if e != nil {
        t.Fail()
    }
}