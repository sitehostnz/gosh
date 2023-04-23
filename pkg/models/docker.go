package models

type (
	// DockerCompose represents a docker-compose.yml file.
	// See https://docs.docker.com/compose/compose-file/

	// Service represents a service in the docker-compose.yml file.
	Service struct {
		ContainerName string   `yaml:"container_name"`
		Environment   []string `yaml:"environment"`
		Expose        []string `yaml:"expose"`
		Image         string   `yaml:"image"`
		Labels        []string `yaml:"labels"`
		Restart       string   `yaml:"restart"`
		Volumes       []string `yaml:"volumes"`
	}

	// Networks represents the networks in the docker-compose.yml file.
	Networks struct {
		Default DefaultNetwork `yaml:"default"`
	}

	// DefaultNetwork represents the default network in the docker-compose.yml file.
	DefaultNetwork struct {
		External ExternalNetwork `yaml:"external"`
	}

	// ExternalNetwork represents the external network in the docker-compose.yml file.
	ExternalNetwork struct {
		Name string `yaml:"name"`
	}

	// DockerCompose represents a docker-compose.yml file.
	DockerCompose struct {
		Version  string             `yaml:"version"`
		Services map[string]Service `yaml:"services"`
		Networks Networks           `yaml:"networks"`
	}
)
