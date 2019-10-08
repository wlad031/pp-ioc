package pp_ioc

type PostProcessor interface {
    PostProcess(ctx Context) error
}
