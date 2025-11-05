package infrastructure

type ContainerInterface interface {
	CreateContainer(path string) error
	ChechProxyNetworkExists() error
}
