package infrastructure

type ContainerInterface interface {
	CreateContainer(targetRepo string, path string) error
}