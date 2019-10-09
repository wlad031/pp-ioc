package pp_ioc

import (
    "github.com/pkg/errors"
    ps "github.com/wlad031/pp-ioc/property_source"
    "reflect"
    "strings"
)

type beanDefinition struct {
    key          *bindKey
    dependencies map[uint16]*dependency
    priority     int
    paramTypes   []reflect.Type
    scope        BeanScope
    factory      *beanFactory
    graphIndex   int
    _bean        *bean // Do not use it directly!
}

func (bd *beanDefinition) createBean(params []reflect.Value) (*bean, error) {
    switch bd.scope {
    case ScopeSingleton:
        {
            if bd._bean != nil {
                return bd._bean, nil
            } else {
                instance, e := bd.factory.call(params)
                if e != nil {
                    return nil, e
                }
                bd._bean = &bean{
                    definition: bd,
                    instance:   instance,
                }
                return bd._bean, nil
            }
        }
    case ScopePrototype:
        {
            instance, e := bd.factory.call(params)
            if e != nil {
                return nil, e
            }
            return &bean{
                definition: bd,
                instance:   instance,
            }, nil
        }
    default:
        return nil, errors.New("Unknown bean scope " + bd.scope.String())
    }
}

func (bd *beanDefinition) isSuitableForDependencyByQualifier(dependency *dependency) bool {
    var isSuitableByName = false
    for _, name := range bd.key.qualifiers {
        if name == dependency.qualifier {
            isSuitableByName = true
            break
        }
    }
    if !isSuitableByName {
        return false
    }
    return true
}

// TODO: refactor this function
func (bd *beanDefinition, ) isSuitableForDependencyByType(dependency *dependency) bool {
    dependencyType := dependency.type_
    if dependencyType.Kind() == reflect.Ptr {
        dependencyType = dependencyType.Elem()
    }

    if dependencyType.Kind() == reflect.Struct {
        if bd.key.type_.Kind() == reflect.Struct {
            return bd.key.type_ == dependencyType
        }
        if bd.key.type_.Kind() == reflect.Ptr {
            if bd.key.type_.Elem().Kind() == reflect.Struct {
                return bd.key.type_.Elem() == dependencyType
            }
        }
    }
    if dependencyType.Kind() == reflect.Interface {
        if bd.key.type_.Kind() == reflect.Struct {
            return bd.key.type_.Implements(dependencyType)
        }
        if bd.key.type_.Kind() == reflect.Interface { // TODO: may not work correctly
            return bd.key.type_.Implements(dependencyType)
        }
        if bd.key.type_.Kind() == reflect.Ptr {
            if bd.key.type_.Elem().Kind() == reflect.Struct {
                return bd.key.type_.Elem().Implements(dependencyType)
            }
            if bd.key.type_.Elem().Kind() == reflect.Interface { // TODO: may not work correctly
                return bd.key.type_.Elem().Implements(dependencyType)
            }
        }
    }

    return false
}

func (bd *beanDefinition) updateGraphIndex(newGraphIndex int) {
    bd.graphIndex = newGraphIndex
}

func (bd *beanDefinition) isPropertySource() bool {
    return bd.key.type_.Implements(reflect.TypeOf((*ps.PropertySource)(nil)).Elem())
}

func (bd *beanDefinition) isPostProcessor() bool {
    return bd.key.type_.Implements(reflect.TypeOf((*PostProcessor)(nil)).Elem())
}

func (bd *beanDefinition) shortString() string {
    return "BeanDef{" + bd.key.String() + "}"
}

func (bd *beanDefinition) String() string {
    var depStrings []string
    for _, dep := range bd.dependencies {
        depStrings = append(depStrings, dep.String())
    }
    return "BeanDef{" + bd.key.String() + ":" +
        bd.scope.String() + ":[" +
        strings.Join(depStrings, ",") + "]}"
}