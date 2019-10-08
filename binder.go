package pp_ioc

import (
    "github.com/pkg/errors"
    "github.com/wlad031/pp-algo/list"
    "reflect"
    "strconv"
    "strings"
)

type Binder struct {
    qualifiers  []string
    scope       BeanScope
    beanFactory beanFactory
    priority    int
}

func NewBinder() *Binder {
    return &Binder{
        qualifiers: []string{},
        scope:      ScopeSingleton,
        priority:   DefaultPriority,
    }
}

func (b *Binder) String() string {
    return "Binder{" + strconv.Itoa(b.priority) + ":" + b.scope.String() +
        ":[" + strings.Join(b.qualifiers, ",") + "]}"
}

func (b *Binder) Scope(scope BeanScope) *Binder {
    b.scope = scope
    return b
}

func (b *Binder) Qualifiers(qualifiers ...string) *Binder {
    b.qualifiers = qualifiers
    return b
}

func (b *Binder) Priority(priority int) *Binder {
    b.priority = priority
    return b
}

func (b *Binder) Factory(factoryFunc interface{}) *Binder {
    b.beanFactory = beanFactory{
        factoryFunction: factoryFunc,
        isMethod:        false,
    }
    return b
}

func (b *Binder) buildBindKey() (*bindKey, error) {
    e := b.validate()
    if e != nil {
        return nil, e
    }
    return &bindKey{
        qualifiers: b.qualifiers,
        type_:      reflect.TypeOf(b.beanFactory.factoryFunction).Out(0),
    }, nil
}

// TODO: refactor this function
// TODO: rework mechanism of configurations:
//       they should provide somehow the struct
//       with all data about bean definition
func (b *Binder) collectNestedBinders(res list.List) error {
    out := reflect.TypeOf(b.beanFactory.factoryFunction).Out(0)
    if out.Kind() == reflect.Ptr {
        out = out.Elem()
    }
    if out.Kind() == reflect.Struct {
        for i := 0; i < out.NumField(); i++ {
            field := out.Field(i)
            tag := field.Tag
            factoryFuncName, hasFactoryFuncName := tag.Lookup("factory")
            if !hasFactoryFuncName {
                continue
            }
            qualifiers, hasQualifiers := tag.Lookup("qualifiers")
            priority, hasPriority := tag.Lookup("priority")
            scope, hasScope := tag.Lookup("scope")

            names := []string{factoryFuncName}
            if hasQualifiers {
                names = strings.Split(qualifiers, ",")
            }

            nestedBinder := NewBinder().Qualifiers(names...)

            if hasScope {
                v, e := FromString(scope)
                if e != nil {
                    return e
                }
                nestedBinder.Scope(v)
            }
            if hasPriority {
                v, e := strconv.Atoi(priority)
                if e != nil {
                    return errors.Wrap(e, "Cannot parse priority for string " + priority)
                }
                nestedBinder.Priority(v)
            }

            f, ok := out.MethodByName(factoryFuncName)
            if !ok {
                return errors.New("Cannot find factory function " + factoryFuncName)
            }
            nestedBinder.Factory(f.Func.Interface())
            nestedBinder.beanFactory.isMethod = true

            res.Append(nestedBinder)
            return nestedBinder.collectNestedBinders(res)
        }
    }
    return nil
}

// TODO: use more expressive error messages
func (b *Binder) validate() error {
    if b.beanFactory.factoryFunction == nil {
        return errors.New("Invalid factory function: nil")
    }
    factoryType := reflect.TypeOf(b.beanFactory.factoryFunction)
    factoryValue := reflect.ValueOf(b.beanFactory.factoryFunction)
    if factoryValue.Kind() != reflect.Func {
        return errors.New("Invalid factory function: not a function")
    }
    // TODO: validate in params
    //numIn := factoryType.NumIn()
    //if numIn > 1 {
    //    return errors.New("Invalid factory function: " +
    //        "invalid number of IN parameters (must be 0 or 1)")
    //}
    //if numIn == 1 {
    //    paramType := factoryType.In(0)
    //    if paramType.Kind() != reflect.Struct {
    //        return errors.New("Invalid factory function: " +
    //            "invalid type of IN parameter (must be struct)")
    //    }
    //}
    numOut := factoryType.NumOut()
    if numOut != 1 && numOut != 2 {
        return errors.New("Invalid factory function: " +
            "invalid number of OUT parameters (must be 1 or 2)")
    }
    outParamType := factoryType.Out(0)
    if outParamType.Kind() != reflect.Interface &&
        outParamType.Kind() != reflect.Struct &&
        !(outParamType.Kind() == reflect.Ptr &&
            outParamType.Elem().Kind() != reflect.Interface &&
            outParamType.Kind() != reflect.Struct) {
        return errors.New("Invalid factory function: invalid type (" +
            outParamType.Kind().String() +
            ") of the first OUT parameters " +
            "(must be interface or struct or pointer to interface or struct)")
    }
    if numOut == 2 {
        if factoryType.Out(1).Kind() != reflect.Interface ||
            !factoryType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
            return errors.New("Invalid factory function: invalid type of " +
                "the second OUT parameter (must implement 'error')")
        }
    }
    return nil
}
