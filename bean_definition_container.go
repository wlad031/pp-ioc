package pp_ioc

type beanDefinitionContainer struct {
    ls []*beanDefinition
}

func newBeanDefinitionList() *beanDefinitionContainer {
    return &beanDefinitionContainer{ls: []*beanDefinition{}}
}

func (container *beanDefinitionContainer) iterate() <-chan *beanDefinition {
    c := make(chan *beanDefinition)
    go func() {
        for _, v := range container.ls {
            c <- v
        }
        close(c)
    }()
    return c
}

func (container *beanDefinitionContainer) add(bd *beanDefinition) {
    // FIXME: performance is not good because of slice copying
    container.ls = append(container.ls, bd)
    for i := 0; i < len(container.ls); i++ {
        cur := container.ls[i]
        if cur.priority < bd.priority {
            copy(container.ls[i+1:], container.ls[i:])
            container.ls[i] = bd
            break
        }
    }
}
