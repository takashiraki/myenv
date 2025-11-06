package interfaces

type Project struct {
	ContainerName string            `json:"container_name"`
	ContainerProxy string            `json:"container_proxy"`
	Path          string            `json:"path"`
	Lang          string            `json:"lang"`
	Fw            string            `json:"framework"`
	Options       map[string]string `json:"options"`
}

type Config struct {
	Lang              string             `json:"lang"`
	Version           string             `json:"version"`
	ContainerRuntime  string             `json:"containerRuntime"`
	Projects          map[string]Project `json:"projects"`
	Modules           map[string]ModuleConfig `json:"modules"`
}

type ModuleConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}