package pp_ioc

import (
    "fmt"
    algo "github.com/wlad031/pp-algo"
    "reflect"
)

type ObjectFactory func(ctx Context) interface{}

type BeanDefinition struct {
    key BindingKey
    dep Dep
}

func NewBean(key BindingKey, dep Dep) BeanDefinition {
    return BeanDefinition{
        key: key,
        dep: dep,
    }
}

type BindingKey interface {
    fmt.Stringer
    Name(name string) BindingKey
    Type(type_ reflect.Type) BindingKey
}

type Dep interface {
    fmt.Stringer
    Add(key BindingKey) Dep
    Iterate() <-chan BindingKey
}

func (d depImpl) Iterate() <-chan BindingKey {
    c := make(chan BindingKey)
    go func() {
        for _, v := range d.keys {
            if v != nil {
                c <- v
            }
        }
        close(c)
    }()
    return c
}

func NewDep() Dep {
    return &depImpl{0, [100]BindingKey{}}
}

type depImpl struct {
    curI uint
    keys [100]BindingKey
}

func (d depImpl) String() string {
    res := ""
    var i uint
    for i = 0; i < d.curI; i += 1 {
        res = res + d.keys[i].String() + ", "
    }
    return res
}

func (d depImpl) Add(key BindingKey) Dep {
    d.keys[d.curI] = key
    d.curI += 1
    return d
}

func NewBindingKey() BindingKey {
    return &bindingKeyImpl{}
}

type bindingKeyImpl struct {
    name string
    type_ reflect.Type
}

func (b *bindingKeyImpl) String() string {
    return b.name + ":" + b.type_.Name()
}

func (b *bindingKeyImpl) Name(name string) BindingKey {
    b.name = name
    return b
}

func (b *bindingKeyImpl) Type(type_ reflect.Type) BindingKey {
    b.type_ = type_
    if b.name == "" {
        b.name = b.type_.String()
    }
    return b
}

type Context interface {
    Bind(key BeanDefinition, o ObjectFactory)
    Get(key BindingKey) interface{}
    Build() error
}

func NewContext() Context {
    return &contextImpl{
        beanDefinitions: map[BeanDefinition]ObjectFactory{},
        container: map[BindingKey]interface{}{},
    }
}

type contextImpl struct {
    beanDefinitions map[BeanDefinition]ObjectFactory
    container map[BindingKey]interface{}
}

func (c *contextImpl) Build() error {
    graph := algo.NewOrientedGraph()
    m := map[BindingKey]uint{}
    m1 := map[uint]BindingKey{}
    for key, of := range c.beanDefinitions {
        index, _ := graph.AddNode(of)
        m[key.key] = index
        m1[index] = key.key
    }
    for key, _ := range c.beanDefinitions {
        for key1 := range key.dep.Iterate() {
            if from, ok1 := m[key.key]; ok1 {
                if to, ok2 := m[key1]; ok2 {
                    graph.AddEdge(from, to)
                } else {
                    fmt.Println(key1.String())
                }
                fmt.Println(key.key)
            }
        }
    }
    sortIndexes := graph.TopologicalSort()
    for _, v := range sortIndexes {
        fmt.Println(v)
    }
    for _, ind := range sortIndexes {
        dataForIndex, _ := graph.GetDataForIndex(ind)
        i := dataForIndex.(ObjectFactory)(c)
        c.container[m1[ind]] = i
    }
    return nil
}

func (c *contextImpl) Bind(key BeanDefinition, o ObjectFactory) {
    if _, ok := c.beanDefinitions[key]; ok {
        panic("key is already in the context") // TODO: normal error handling
    }
    c.beanDefinitions[key] = o
}

func (c contextImpl) Get(key BindingKey) interface{} {
    if b, ok := c.container[key]; !ok {
       return nil
    } else {
        return b
    }
}
