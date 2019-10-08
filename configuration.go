package pp_ioc

type Configuration interface {
    Bind(ctx Context) error
}
