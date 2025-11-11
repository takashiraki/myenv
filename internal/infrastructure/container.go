package infrastructure

type ContainerInterface interface {
	CreateContainer(path string) error
	ChechProxyNetworkExists() error
	ChechInfraNetworkExists() error
	CreateProxyNetwork() error
	CreateInfraNetwork() error
	ExecCommand(serviceName string, arguments ...string) error
}
