package application

import "myenv/internal/config/domains"

type CreateConfigInteractor struct {
	repository domains.ConfigRepository
	factory domains.ContainerFactory
}

func NewCreateConfigInteractor(
	repository domains.ConfigRepository,
	factory domains.ContainerFactory,
) *CreateConfigInteractor {
	return &CreateConfigInteractor{
		repository: repository,
		factory: factory,
	}
}