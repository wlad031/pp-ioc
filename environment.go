package pp_ioc

import (
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    ps "github.com/wlad031/pp-ioc/property_source"
    logCtx "github.com/wlad031/pp-logging"
)

type Environment interface {
    addPropertySource(b *bean) error

    GetProperty(key string) (string, error)
    GetPropertyOrDefault(key string, defaultValue string) string
    GetAllProperties() map[string]string
}

func newEnvironment() Environment {
    return &environmentImpl{
        logger:          logCtx.Get("IOC.Environment"),
        propertySources: []ps.PropertySource{},
    }
}

type environmentImpl struct {
    logger          *logCtx.NamedLogger
    propertySources []ps.PropertySource
}

func (env *environmentImpl) addPropertySource(b *bean) error {
    propertySource := b.instance.(ps.PropertySource)
    env.propertySources = append(env.propertySources, propertySource)
    env.logger.WithFields(log.Fields{
        "beanDef": b.definition.shortString(),
    }).Info("Property source added")
    return nil
}

func (env *environmentImpl) GetProperty(key string) (string, error) {
    var res string
    for _, propertySource := range env.propertySources {
        if v, e := propertySource.Get(key); e == nil {
            res = v
        }
    }
    if res == "" {
        return "", errors.New("Cannot find property " + key)
    } else {
        return res, nil
    }
}

func (env *environmentImpl) GetPropertyOrDefault(key string, defaultValue string) string {
    if v, e := env.GetProperty(key); e != nil {
        return defaultValue
    } else {
        return v
    }
}

func (env *environmentImpl) GetAllProperties() map[string]string {
    res := map[string]string{}
    for _, propertySource := range env.propertySources {
        allProps := propertySource.GetAll()
        for k, v := range allProps {
            res[k] = v
        }
    }
    return res
}
