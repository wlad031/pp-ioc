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
    beanFactory *beanFactory
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
    b.beanFactory = newBeanFactory(factoryFunc, false)
    return b
}

func (b *Binder) buildBindKey() *bindKey {
    return &bindKey{
        qualifiers: b.qualifiers,
        type_:      reflect.TypeOf(b.beanFactory.factoryFunction).Out(0),
    }
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
            factoryFuncName, hasFactoryFuncName := tag.Lookup(TagFactory)
            if !hasFactoryFuncName {
                continue
            }
            qualifiers, hasQualifiers := tag.Lookup(TagQualifiers)
            priority, hasPriority := tag.Lookup(TagPriority)
            scope, hasScope := tag.Lookup(TagScope)

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
            if e := nestedBinder.collectNestedBinders(res); e != nil {
                return e
            }
        }
    }
    return nil
}
