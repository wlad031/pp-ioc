package pp_ioc

import (
    "github.com/pkg/errors"
    "reflect"
    "strings"
)

type beanFactory struct {
    factoryFunction interface{}
    isMethod        bool

    type_ reflect.Type

    _inParamTypes  []reflect.Type
    _outParamTypes []reflect.Type
}

func newBeanFactory(factoryFunction interface{}, isMethod bool) *beanFactory {
    bf := &beanFactory{
        factoryFunction: factoryFunction,
        isMethod:        isMethod,
        type_:           reflect.TypeOf(factoryFunction),
    }
    return bf
}

func (bf *beanFactory) call(params []reflect.Value) (interface{}, error) {
    factoryCallResult := reflect.ValueOf(bf.factoryFunction).Call(params)
    if len(factoryCallResult) > 1 {
        factoryError := factoryCallResult[1].Interface()
        if factoryError != nil {
            return nil, errors.Wrap(factoryError.(error),
                "Error happened during calling bean factory")
        }
    }
    instance := factoryCallResult[0].Interface()
    return instance, nil
}

func (bf *beanFactory) collectDependencies() (
    dependencies map[uint16]*dependency,
    paramTypes []reflect.Type,
) {
    dependencies = map[uint16]*dependency{}

    if len(bf._inParamTypes) == 0 {
        return dependencies, nil
    }

    var paramIndex uint16 = 0

    // TODO: work with different kinds of params (slices, arrays, etc.)

    for i := 0; i < len(bf._inParamTypes); i++ {
        paramType := bf._inParamTypes[i]

        if (i == 0 && bf.isMethod) ||
            (paramType.Kind() != reflect.Struct) ||
            (paramType.Kind() == reflect.Ptr && paramType.Elem().Kind() != reflect.Struct) {
            dependencies[paramIndex] = newBeanDependency(
                "",
                "", false,
                paramType,
                paramIndex)
            paramTypes = append(paramTypes, paramType)
            paramIndex += 1
        } else {
            structType := paramType
            if paramType.Kind() == reflect.Ptr {
                structType = paramType.Elem()
            }

            for j := 0; j < structType.NumField(); j++ {
                structField := structType.Field(j)
                if qualifierTag, ok := structField.Tag.Lookup(TagQualifier); ok {
                    dependencies[paramIndex] = newBeanDependency(
                        structField.Name,
                        qualifierTag, true,
                        structField.Type,
                        paramIndex)
                } else if valueTag, ok := structField.Tag.Lookup(TagValue); ok {
                    provider := parseValueTag(valueTag)
                    dependencies[paramIndex] = newValueDependency(
                        structField.Name,
                        provider.qualifier, true,
                        provider,
                        structField.Type,
                        paramIndex)
                } else {
                    dependencies[paramIndex] = newBeanDependency(
                        structField.Name,
                        "", false,
                        structField.Type,
                        paramIndex)
                }
                paramIndex += 1
            }

            paramTypes = append(paramTypes, structType)
        }
    }

    return dependencies, paramTypes
}

const (
    ValueTagSep    = ":"
    ValueTagPrefix = "${"
    ValueTagSuffix = "}"
)

func parseValueTag(tag string) *valueProvider {
    tag = tag[2 : len(tag)-1] // remove prefix and suffix
    splitRes := strings.Split(tag, ValueTagSep)
    var defaultValue string
    var hasDefault bool
    if len(splitRes) > 1 {
        defaultValue = strings.Join(splitRes[1:], ValueTagSep)
        hasDefault = true
    }
    return &valueProvider{
        qualifier:    splitRes[0],
        defaultValue: defaultValue,
        hasDefault:   hasDefault,
    }
}
