package pp_ioc

const maxUint = ^uint(0)

const (
    DefaultPriority = 0
    ContextPriority = 1_000_000

    PropertySourceHighestPriority = 900_000
    PropertySourceLowestPriority = 899_000

    EnvironmentPriority = 800_000

    HighestPriority = 500_000
    LowestPriority  = -500_000
)
