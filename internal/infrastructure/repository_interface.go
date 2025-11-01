package infrastructure

type RepositoryInterface interface {
	CloneRepo(repoUrl string, targetPath string) error
}
