package application

type CreateConfigUseCase interface {
	Execute(lang string, containerRuntime string) error
}