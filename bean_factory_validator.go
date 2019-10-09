package pp_ioc

import (
    "github.com/pkg/errors"
    logCtx "github.com/wlad031/pp-logging"
    "reflect"
)

type beanFactoryValidator struct {
    logger *logCtx.NamedLogger
}

func newBeanFactoryValidator() *beanFactoryValidator {
    return &beanFactoryValidator{
        logger: logCtx.Get("IOC.BeanFactoryValidator"),
    }
}

func (v *beanFactoryValidator) validate(beanFactory *beanFactory) error {
    if beanFactory.factoryFunction == nil {
        return errors.New("Invalid factory function: nil")
    }
    factoryValue := reflect.ValueOf(beanFactory.factoryFunction)
    if factoryValue.Kind() != reflect.Func {
        return errors.New("Invalid factory function: not a function")
    }

    for i := 0; i < beanFactory.type_.NumIn(); i++ {
        beanFactory._inParamTypes = append(beanFactory._inParamTypes, beanFactory.type_.In(i))
    }
    for i := 0; i < beanFactory.type_.NumOut(); i++ {
        beanFactory._outParamTypes = append(beanFactory._outParamTypes, beanFactory.type_.Out(i))
    }

    for _, paramType := range beanFactory._inParamTypes {
        if paramType.Kind() != reflect.Struct &&
            paramType.Kind() != reflect.Interface &&
            paramType.Kind() != reflect.Ptr &&
            !(paramType.Kind() == reflect.Ptr &&
                paramType.Elem().Kind() != reflect.Struct &&
                paramType.Kind() != reflect.Interface) {
            return errors.New("Invalid factory function: only struct/interface or " +
                "*struct/*interface IN params allowed")
        }
        if paramType.Kind() == reflect.Struct ||
            (paramType.Kind() == reflect.Ptr && paramType.Elem().Kind() == reflect.Struct) {
            if paramType.Kind() == reflect.Ptr {
                paramType = paramType.Elem()
            }
            for i := 0; i < paramType.NumField(); i++ {
                fieldType := paramType.Field(i)
                switch fieldType.Type.Kind() {

                case reflect.Bool:
                case reflect.Int:
                case reflect.Int8:
                case reflect.Int16:
                case reflect.Int32:
                case reflect.Int64:
                case reflect.Uint:
                case reflect.Uint8:
                case reflect.Uint16:
                case reflect.Uint32:
                case reflect.Uint64:
                case reflect.Uintptr:
                case reflect.Float32:
                case reflect.Float64:
                case reflect.String:
                case reflect.Interface:
                case reflect.Struct:
                case reflect.Ptr:
                default:
                    continue
                    // FIXME:
                    // return errors.New("Invalid factory function: invalid in-struct field type")
                }
            }
        }
    }
    numOut := len(beanFactory._outParamTypes)
    if numOut != 1 && numOut != 2 {
        return errors.New("Invalid factory function: " +
            "invalid number of OUT parameters (must be 1 or 2)")
    }
    outParamType := beanFactory._outParamTypes[0]
    if outParamType.Kind() != reflect.Interface &&
        outParamType.Kind() != reflect.Struct &&
        !(outParamType.Kind() == reflect.Ptr &&
            outParamType.Elem().Kind() != reflect.Interface &&
            outParamType/*FIXME: .Elem()*/.Kind() != reflect.Struct) {
        return errors.New("Invalid factory function: invalid type (" +
            outParamType.Kind().String() +
            ") of the first OUT parameters " +
            "(must be interface/struct or *interface/*struct)")
    }
    if numOut == 2 {
        if beanFactory.type_.Out(1).Kind() != reflect.Interface ||
            !beanFactory.type_.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
            return errors.New("Invalid factory function: invalid type of " +
                "the second OUT parameter (must implement 'error')")
        }
    }
    return nil
}
