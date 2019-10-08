package pp_ioc

import (
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    "github.com/wlad031/pp-algo/list"
    logCtx "github.com/wlad031/pp-logging"
    "reflect"
    "strings"
)

//region public

const (
    // Name of the bean that represents the context itself
    ContextBeanName     = "ApplicationContext"
    EnvironmentBeanName = "Environment"

    QualifierTag = "qualifier"
    ValueTag     = "value"
    ValueTagSep  = ":"
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
    logger := logCtx.Get("IOC")
    logger.SetLevel(log.DebugLevel)
    ctx := contextImpl{
        logger:          logger,
        binders:         newBinderContainer(),
        beanDefinitions: newBeanDefinitionList(),
        graph:           newContextGraph(),
        container:       newBeanContainer(),
        postProcessors:  newPostProcessorContainer(),
        environment:     newEnvironment(),
        initialized:     false,
    }
    return &ctx
}

//endregion

//region private

type contextImpl struct {
    logger          *logCtx.NamedLogger
    binders         *binderContainer
    beanDefinitions *beanDefinitionContainer
    graph           *contextGraph
    container       *beanContainer
    postProcessors  *postProcessorContainer
    environment     Environment
    initialized     bool // TODO: use this field somewhere
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

// Creates the binder for context itself.
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
    key, e := binder.buildBindKey()
    if e != nil {
        return e
    }

    dependencies, paramTypes := collectDependencies(binder.beanFactory)
    definition := &beanDefinition{
        key:          key,
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
        e := ctx.environment.addPropertySource(beanInstance)
        if e != nil {
            return e
        }
    }
    if beanInstance.definition.isPostProcessor() {
        e := ctx.postProcessors.add(beanInstance)
        if e != nil {
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
                if !definition.factory.isMethod {
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

func collectDependencies(factoryFunc beanFactory) (
    dependencies map[uint16]*dependency,
    paramTypes []reflect.Type,
) {
    dependencies = map[uint16]*dependency{}

    factoryType := reflect.TypeOf(factoryFunc.factoryFunction)
    if factoryType.NumIn() == 0 {
        return dependencies, nil
    }

    var paramIndex uint16 = 0
    for i := 0; i < factoryType.NumIn(); i++ {
        paramType := factoryType.In(i)

        if paramType.Kind() == reflect.Struct {
            if !factoryFunc.isMethod {
                for j := 0; j < paramType.NumField(); j++ {
                    structField := paramType.Field(j)
                    // TODO: work with different kinds of params (structs, interfaces, primitives, arrays, etc.)
                    if qualifierTag, ok := structField.Tag.Lookup(QualifierTag); ok {
                        dependencies[paramIndex] = &dependency{
                            name:         structField.Name,
                            qualifier:    qualifierTag,
                            hasQualifier: true,
                            defaultValue: "",
                            hasDefault:   false,
                            type_:        structField.Type,
                            index:        paramIndex,
                            isBean:       true,
                        }
                        paramIndex += 1
                        continue
                    }
                    if valueTag, ok := structField.Tag.Lookup(ValueTag); ok {
                        splitRes := strings.Split(valueTag, ValueTagSep)
                        var defaultValue string
                        var hasDefault bool
                        if len(splitRes) > 1 {
                            defaultValue = splitRes[1]
                            hasDefault = true
                        }
                        dependencies[paramIndex] = &dependency{
                            name:         structField.Name,
                            qualifier:    splitRes[0],
                            hasQualifier: true,
                            defaultValue: defaultValue,
                            hasDefault:   hasDefault,
                            type_:        structField.Type,
                            index:        paramIndex,
                            isBean:       false,
                        }
                        paramIndex += 1
                        continue
                    }
                    dependencies[paramIndex] = &dependency{
                        name:         structField.Name,
                        qualifier:    "",
                        hasQualifier: false,
                        defaultValue: "",
                        hasDefault:   false,
                        type_:        structField.Type,
                        index:        paramIndex,
                        isBean:       true,
                    }
                    paramIndex += 1
                }
                paramTypes = append(paramTypes, paramType)
            } else {
                dependencies[paramIndex] = &dependency{
                    name:         "",
                    qualifier:    "",
                    hasQualifier: false,
                    defaultValue: "",
                    hasDefault:   false,
                    type_:        paramType,
                    index:        paramIndex,
                    isBean:       true,
                }
                paramTypes = append(paramTypes, paramType)
                paramIndex += 1
            }
        } else {
            dependencies[paramIndex] = &dependency{
                name:         "",
                qualifier:    "",
                hasQualifier: false,
                defaultValue: "",
                hasDefault:   false,
                type_:        paramType,
                index:        paramIndex,
                isBean:       true,
            }
            paramTypes = append(paramTypes, paramType)
            paramIndex += 1
        }
    }

    return dependencies, paramTypes
}

//endregion

