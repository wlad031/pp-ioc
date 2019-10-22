package pp_ioc

import (
    "github.com/pkg/errors"
    "reflect"
    "strconv"
)

type valueProvider struct {
    qualifier    string
    defaultValue string
    hasDefault   bool
}

type dependency struct {
    name          string
    qualifier     string
    hasQualifier  bool
    valueProvider *valueProvider
    type_         reflect.Type
    index         uint16
    isBean        bool
    isValue       bool
}

func newBeanDependency(
    name string,
    qualifier string,
    hasQualifier bool,
    type_ reflect.Type,
    index uint16,
) *dependency {
    return &dependency{
        name:          name,
        qualifier:     qualifier,
        hasQualifier:  hasQualifier,
        valueProvider: nil,
        type_:         type_,
        index:         index,
        isBean:        true,
        isValue:       false,
    }
}

func newValueDependency(
    name string,
    qualifier string,
    hasQualifier bool,
    valueProvider *valueProvider,
    type_ reflect.Type,
    index uint16,
) *dependency {
    return &dependency{
        name:          name,
        qualifier:     qualifier,
        hasQualifier:  hasQualifier,
        valueProvider: valueProvider,
        type_:         type_,
        index:         index,
        isBean:        false,
        isValue:       true,
    }
}

func (d *dependency) String() string {
    return "Dep{" + d.qualifier + ":" + d.type_.String() + "}"
}

func (d *dependency) parsePropertyValue(propValue string) (reflect.Value, error) {
    switch d.type_.Kind() {

    case reflect.Bool:
        parsed, e := strconv.ParseBool(propValue)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to bool")
        }
        return reflect.ValueOf(parsed), nil

    case reflect.Int:
        parsed, e := strconv.ParseInt(propValue, 10, 64)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to int")
        }
        return reflect.ValueOf(int(parsed)), nil
    case reflect.Int8:
        parsed, e := strconv.ParseInt(propValue, 10, 8)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to int8")
        }
        return reflect.ValueOf(int8(parsed)), nil
    case reflect.Int16:
        parsed, e := strconv.ParseInt(propValue, 10, 16)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to int16")
        }
        return reflect.ValueOf(int16(parsed)), nil
    case reflect.Int32:
        parsed, e := strconv.ParseInt(propValue, 10, 32)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to int32")
        }
        return reflect.ValueOf(int32(parsed)), nil
    case reflect.Int64:
        parsed, e := strconv.ParseInt(propValue, 10, 64)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to int64")
        }
        return reflect.ValueOf(int64(parsed)), nil

    case reflect.Uint:
        parsed, e := strconv.ParseUint(propValue, 10, 64)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to uint")
        }
        return reflect.ValueOf(parsed), nil
    case reflect.Uint8:
        parsed, e := strconv.ParseUint(propValue, 10, 8)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to uint8")
        }
        return reflect.ValueOf(uint8(parsed)), nil
    case reflect.Uint16:
        parsed, e := strconv.ParseUint(propValue, 10, 16)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to uint16")
        }
        return reflect.ValueOf(uint16(parsed)), nil
    case reflect.Uint32:
        parsed, e := strconv.ParseUint(propValue, 10, 32)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to uint32")
        }
        return reflect.ValueOf(uint32(parsed)), nil
    case reflect.Uint64:
        parsed, e := strconv.ParseUint(propValue, 10, 64)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to uint64")
        }
        return reflect.ValueOf(uint64(parsed)), nil

    case reflect.Float32:
        parsed, e := strconv.ParseFloat(propValue, 32)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to float32")
        }
        return reflect.ValueOf(float32(parsed)), nil
    case reflect.Float64:
        parsed, e := strconv.ParseFloat(propValue, 64)
        if e != nil {
            return reflect.Value{}, errors.Wrap(e, "Property "+propValue+" cannot be converted to float64")
        }
        return reflect.ValueOf(float64(parsed)), nil

    case reflect.String:
        return reflect.ValueOf(propValue), nil
    }

    return reflect.Value{}, errors.New("Unknown type=" + d.type_.String() + " of property " + d.qualifier)
}
