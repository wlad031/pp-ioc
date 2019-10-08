package pp_ioc

import (
    "github.com/pkg/errors"
    "reflect"
)

type beanFactory struct {
    factoryFunction interface{}
    isMethod        bool
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
