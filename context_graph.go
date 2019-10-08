package pp_ioc

import (
    "github.com/pkg/errors"
    log "github.com/sirupsen/logrus"
    g "github.com/wlad031/pp-algo/graph"
    logCtx "github.com/wlad031/pp-logging"
)

type contextGraph struct {
    logger *logCtx.NamedLogger
    graph  g.OrientedGraph
    sorted []int
}

func newContextGraph() *contextGraph {
    logger := logCtx.Get("IOC.ContextGraph")
    logger.SetLevel(log.TraceLevel)
    return &contextGraph{
        logger: logger,
        graph:  g.NewOrientedGraph(),
    }
}

func (ctxG *contextGraph) build(beanDefinitions *beanDefinitionContainer) error {
    ctxG.logger.Info("Building the dependency graph...")

    e := ctxG.addGraphNodes(beanDefinitions)
    if e != nil {
        return e
    }
    e = ctxG.addGraphEdges(beanDefinitions)
    if e != nil {
        return e
    }
    ctxG.sorted, e = ctxG.graph.TopologicalSort()
    if e != nil {
        return e
    }
    return nil
}

func (ctxG *contextGraph) iterate() <-chan *beanDefinition {
    c := make(chan *beanDefinition)
    go func() {
        for _, ind := range ctxG.sorted {
            data, _ := ctxG.graph.GetDataForIndex(ind)
            c <- data.(*beanDefinition)
        }
        close(c)
    }()
    return c
}

func (ctxG *contextGraph) addGraphNodes(beanDefinitions *beanDefinitionContainer) error {
    for definition := range beanDefinitions.iterate() {
        index, e := ctxG.graph.AddNode(definition)
        if e != nil {
            return errors.Wrap(e, "Cannot add binding key "+definition.shortString())
        }
        definition.updateGraphIndex(index)
        ctxG.logger.WithFields(log.Fields{
            "beanDef": definition.String(),
            "index":   index,
        }).Trace("Added binding key")
    }
    return nil
}

func (ctxG *contextGraph) addGraphEdges(beanDefinitions *beanDefinitionContainer) error {
    for beanDefinition := range beanDefinitions.iterate() {
        for _, dependency := range beanDefinition.dependencies {
            if !dependency.isBean {
                continue
            }
            from, e := findDefinitionIndexForDefinition(beanDefinitions, beanDefinition)
            if e != nil {
                return e
            }
            toList, e := findDefinitionIndexesForDependency(beanDefinitions, dependency)
            if e != nil {
                return e
            }
            for _, to := range toList {
                graphError := ctxG.graph.AddEdge(from, to)
                if graphError != nil {
                    return errors.Wrap(graphError, "Cannot add dependency for " + beanDefinition.shortString())
                }
                ctxG.logger.WithFields(log.Fields{
                    "from":      beanDefinition.String(),
                    "fromIndex": from,
                    "to":        dependency.String(),
                    "toIndex":   to,
                }).Trace("Added dependency")
            }
        }
    }
    return nil
}

func findDefinitionIndexesForDependency(
    beanDefinitions *beanDefinitionContainer,
    dependency *dependency,
) (foundIndexes []int, e error) {
    var isBeanDefinitionFound = false

    for beanDefinition := range beanDefinitions.iterate() {
        isBeanDefinitionSuitable :=
            (dependency.hasQualifier &&
                beanDefinition.isSuitableForDependencyByQualifier(dependency) &&
                beanDefinition.isSuitableForDependencyByType(dependency)) ||
                (!dependency.hasQualifier &&
                    beanDefinition.isSuitableForDependencyByType(dependency))
        if isBeanDefinitionSuitable {
            isBeanDefinitionFound = true
            foundIndexes = append(foundIndexes, beanDefinition.graphIndex)
        }
    }

    if !isBeanDefinitionFound {
        return nil, errors.New("Cannot find bean definition for dependency " + dependency.String())
    }
    return foundIndexes, nil
}

// TODO: refactor this function
func findDefinitionIndexForDefinition(
    beanDefinitions *beanDefinitionContainer,
    key *beanDefinition,
) (int, error) {
    found := false
    ind := -1
    for definition := range beanDefinitions.iterate() {
        if foo1(definition.key.qualifiers, key.key.qualifiers) /*&& definition.key.type_ == key.key.type_*/ {
            found = true
            ind = definition.graphIndex
            break
        }
    }
    if !found {
        return -1, errors.New("Cannot find bean definition " + key.shortString())
    }
    return ind, nil
}

func foo1(names1 []string, names2 []string) bool {
    if len(names1) != len(names2) {
        return false
    }
    allFound := true
    for _, name1 := range names1 {
        found := false
        for _, name2 := range names2 {
            if name1 == name2 {
                found = true
                break
            }
        }
        if !found {
            allFound = false
            break
        }
    }
    if !allFound {
        return false
    }
    return true
}
