package property_source

type PropertySource interface {
    GetAll() map[string]string
    Get(key string) (string, error)
    GetOrDefault(key string, defaultValue string) string
}
