package utils

import (
	"myenv/internal/config/application"
	"myenv/internal/infrastructure"
)

func ResolveDependenciesContainerBooting(
	container infrastructure.ContainerInterface,
	modules []string,
	config_service application.ConfigService,
) error {
	for _, module := range modules {
		moduleObject, err := config_service.GetModule(module)

		if err != nil {
			
			return err
		}

		if err := container.CreateContainer(moduleObject.Path); err != nil {
			return err
		}
	}

	return nil
}