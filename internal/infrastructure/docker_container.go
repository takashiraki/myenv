package infrastructure

type DockerContainer struct {
	TargetRepo string
}

func NewDockerContainer(targetRepo string) *DockerContainer {
	return &DockerContainer{
		TargetRepo: targetRepo,
	}
}

func (d *DockerContainer) CreateContainer(targetRepo string) error {
	return nil
}