package pp_ioc

import (
    "github.com/pkg/errors"
    "github.com/wlad031/pp-algo/list"
    logCtx "github.com/wlad031/pp-logging"
    "reflect"
)

//region public

const (
    // Name of the bean that represents the context itself
    ContextBeanName     = "ApplicationContext"
    EnvironmentBeanName = "Environment"

    TagValue      = "value"
    TagFactory    = "factory"
    TagQualifiers = "qualifiers"
    TagPriority   = "priority"
    TagScope      = "scope"

    ValueTagSep = ":"
)

type Context interface {
    // Creates and returns new Binder instance
    NewBinder() *Binder
    NewPropertySourceBinder() *Binder

    GetBeanByName(name string) (interface{}, error)              // TODO: implement
    GetBeanByType(type_ reflect.Type) (interface{}, error)       // TODO: implement
    GetAllBeansByType(type_ reflect.Type) ([]interface{}, error) // TODO: implement

    // Returns the context's environment
    GetEnvironment() Environment

    // Builds the entire container. Should be called
    // to instantiate all the beans.
    Build() error
}

// Context constructor
func NewContext() Context {
    ctx := contextImpl{
        logger:               logCtx.Get("IOC"),
        beanFactoryValidator: newBeanFactoryValidator(),
        binders:              newBinderContainer(),
        beanDefinitions:      newBeanDefinitionList(),
        graph:                newContextGraph(),
        container:            newBeanContainer(),
        postProcessors:       newPostProcessorContainer(),
        environment:          newEnvironment(),
        initialized:          false,
    }
    return &ctx
}

//endregion

//region private

type contextImpl struct {
    logger               *logCtx.NamedLogger
    beanFactoryValidator *beanFactoryValidator
    binders              *binderContainer
    beanDefinitions      *beanDefinitionContainer
    graph                *contextGraph
    container            *beanContainer
    postProcessors       *postProcessorContainer
    environment          Environment
    initialized          bool // TODO: use this field somewhere
}

func (ctx *contextImpl) NewBinder() *Binder {
    binder := NewBinder()
    _ = ctx.binders.add(binder)
    return binder
}

func (ctx *contextImpl) NewPropertySourceBinder() *Binder {
    return ctx.NewBinder().
        Priority(PropertySourceHighestPriority).
        Scope(ScopeSingleton)
}

func (ctx *contextImpl) GetBeanByName(name string) (interface{}, error) {
    for bean := range ctx.container.iterate() {
        for _, beanName := range bean.definition.key.qualifiers {
            if beanName == name {
                return bean.instance, nil
            }
        }
    }
    return nil, errors.New("Cannot find bean with name " + name)
}

func (ctx *contextImpl) GetBeanByType(type_ reflect.Type) (interface{}, error) {
    for bean := range ctx.container.iterate() {
        if bean.definition.key.type_ == type_ {
            return bean.instance, nil
        }
    }
    return nil, nil
}

func (ctx *contextImpl) GetAllBeansByType(type_ reflect.Type) ([]interface{}, error) {
    var res []interface{}
    for bean := range ctx.container.iterate() {
        if bean.definition.key.type_.Implements(type_.Elem()) { // FIXME: doesn't work correctly
            res = append(res, bean.instance)
        }
    }
    return res, nil
}

func (ctx *contextImpl) GetEnvironment() Environment {
    return ctx.environment
}

func (ctx *contextImpl) Build() error {
    e := ctx.bindEverything()
    if e != nil {
        return e
    }
    e = ctx.graph.build(ctx.beanDefinitions)
    if e != nil {
        return e
    }
    e = ctx.instantiateBeans()
    if e != nil {
        return e
    }
    e = ctx.runPostProcessors()
    if e != nil {
        return e
    }
    ctx.initialized = true
    return nil
}

func (ctx *contextImpl) bindEverything() error {
    _ = ctx.binders.add(
        ctx.createContextBinder(),
        ctx.createEnvironmentBinder(),
    )

    nestedBinders := list.New()
    for binder := range ctx.binders.iterate() {
        e := binder.collectNestedBinders(nestedBinders)
        if e != nil {
            return e
        }
    }
    for b := range nestedBinders.Iterate() {
        _ = ctx.binders.add(b.(*Binder))
    }

    for binder := range ctx.binders.iterate() {
        e := ctx.bind(binder)
        if e != nil {
            return errors.Wrap(e, "Error happened during binding "+binder.String())
        }
    }
    return nil
}

// Creates a binder for context itself.
// Other beans will be able to use it as a normal dependency.
func (ctx *contextImpl) createContextBinder() *Binder {
    return NewBinder().
        Priority(ContextPriority).
        Scope(ScopeSingleton).
        Qualifiers(ContextBeanName).
        Factory(func() (Context, error) {
            return ctx, nil
        })
}

// Creates a binder for environment instance.
// Other beans will be able to use it as a normal dependency.
func (ctx *contextImpl) createEnvironmentBinder() *Binder {
    return NewBinder().
        Priority(EnvironmentPriority).
        Scope(ScopeSingleton).
        Qualifiers(EnvironmentBeanName).
        Factory(func() (Environment, error) {
            return ctx.environment, nil
        })
}

func (ctx *contextImpl) bind(binder *Binder) error {
    if e := ctx.beanFactoryValidator.validate(binder.beanFactory); e != nil {
        return e
    }

    dependencies, paramTypes := binder.beanFactory.collectDependencies()
    definition := &beanDefinition{
        key:          binder.buildBindKey(),
        dependencies: dependencies,
        paramTypes:   paramTypes,
        scope:        binder.scope,
        priority:     binder.priority,
        factory:      binder.beanFactory,
    }
    ctx.beanDefinitions.add(definition)
    return nil
}

func (ctx *contextImpl) addBeanToContainers(beanInstance *bean) error {
    ctx.container.add(beanInstance)
    if beanInstance.definition.isPropertySource() {
        if e := ctx.environment.addPropertySource(beanInstance); e != nil {
            return e
        }
    }
    if beanInstance.definition.isPostProcessor() {
        if e := ctx.postProcessors.add(beanInstance); e != nil {
            return e
        }
    }
    return nil
}

func (ctx *contextImpl) getDependencyValueInstance(dependency *dependency) (reflect.Value, error) {
    var propertyValue string
    var e error
    if dependency.hasDefault {
        propertyValue = ctx.environment.GetPropertyOrDefault(dependency.qualifier, dependency.defaultValue)
    } else {
        propertyValue, e = ctx.environment.GetProperty(dependency.qualifier)
        if e != nil {
            return reflect.Value{}, e
        }
    }
    return dependency.parsePropertyValue(propertyValue)
}

func (ctx *contextImpl) findDependencyBeanValue(dependency *dependency) (reflect.Value, error) {
    var isBeanFound = false
    var foundBean reflect.Value

    for bean := range ctx.container.iterate() {
        isBeanSuitable :=
            (dependency.hasQualifier &&
                bean.definition.isSuitableForDependencyByQualifier(dependency) &&
                bean.definition.isSuitableForDependencyByType(dependency)) ||
                (!dependency.hasQualifier &&
                    bean.definition.isSuitableForDependencyByType(dependency))
        if isBeanFound && isBeanSuitable {
            return reflect.Value{}, errors.New("Two or more beans are suitable for dependency " + dependency.String())
        }
        if isBeanSuitable {
            isBeanFound = true
            foundBean = reflect.ValueOf(bean.instance)
        }
    }

    if !isBeanFound {
        return reflect.Value{}, errors.New("Cannot find bean for dependency " + dependency.String())
    }
    return foundBean, nil
}

func (ctx *contextImpl) findDependencyValue(dependency *dependency) (reflect.Value, error) {
    if dependency.isValue {
        return ctx.getDependencyValueInstance(dependency)
    }
    if dependency.isBean {
        return ctx.findDependencyBeanValue(dependency)
    }
    return reflect.Value{}, errors.New("Invalid dependency " + dependency.String())
}

// TODO: refactor this function
func (ctx *contextImpl) instantiateBeans() error {
    ctx.logger.Info("Instantiation the beans...")

    for definition := range ctx.graph.iterate() {
        var paramValues []reflect.Value

        var paramIndex uint16 = 0
        for _, paramType := range definition.paramTypes {

            if paramType.Kind() == reflect.Struct {
                if !(definition.factory.isMethod && paramIndex == 0) {
                    structParam := reflect.New(paramType).Elem()
                    for i := 0; i < paramType.NumField(); i++ {
                        dependency := definition.dependencies[paramIndex]
                        instance, e := ctx.findDependencyValue(dependency)
                        if e != nil {
                            return e
                        }
                        structParam.FieldByName(dependency.name).Set(instance)
                        paramIndex += 1
                    }
                    paramValues = append(paramValues, structParam)
                    continue
                }
            }

            dependency := definition.dependencies[paramIndex]
            instance, e := ctx.findDependencyValue(dependency)
            if e != nil {
                return e
            }
            paramValues = append(paramValues, instance)
            paramIndex += 1
        }

        bean, e := definition.createBean(paramValues)
        if e != nil {
            return e
        }
        e = ctx.addBeanToContainers(bean)
        if e != nil {
            return e
        }
    }
    return nil
}

func (ctx *contextImpl) runPostProcessors() error {
    for pp := range ctx.postProcessors.iterate() {
        e := pp.PostProcess(ctx)
        if e != nil {
            return e
        }
    }
    return nil
}

//endregion
