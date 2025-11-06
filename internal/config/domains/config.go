package domains

type (
	Config struct {
		lang              string
		containerRuntime  string
	}

	ConfigRepository interface {
		Get() (Config, error)
		Create(Config) error
	}

	ContainerFactory interface {
		Create(containerRuntime string) (Config, error)
	}
)

func NewConfig(
	lang string,
	containerRuntime string,
) *Config {
	return &Config{
		lang: lang,
		containerRuntime: containerRuntime,
	}
}