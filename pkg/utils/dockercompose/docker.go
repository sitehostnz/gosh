package dockercompose

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/api"
	"github.com/sitehostnz/gosh/pkg/api/cloud/stack/image"
	"github.com/sitehostnz/gosh/pkg/models"
	"gopkg.in/yaml.v2"
)

// GenerateDockerCompose generates a docker compose file for a stack.
func GenerateDockerCompose(ctx context.Context, client *api.Client, request models.GenerateDockerComposeRequest) (dockerCompose string, err error) {
	// Get the image
	i := image.New(client)
	image, err := i.GetImageByCode(ctx, image.GetRequest{Code: request.ImageCode})
	if err != nil {
		return dockerCompose, err
	}

	// registry url (staging)
	registryPath := "registry-staging.sitehost.co.nz"
	imageLastVersion := image.Versions[len(image.Versions)-1].Version

	// create volumes and ports
	volumes := []string{}
	for folder, volume := range image.Labels.NzSitehostImageVolumes.Volumes {
		volumes = append(volumes, "/data/docker0/"+image.Labels.NzSitehostImageType+"/"+request.Name+"/"+folder+":"+volume.Dest+":"+volume.Mode)
	}

	ports := []string{}

	for port, portInfo := range image.Labels.NzSitehostImagePorts {
		if portInfo.Exposed {
			ports = append(ports, port+"/"+portInfo.Protocol)
		}
	}

	compose, err := buildDockerCompose(models.BuildDockerCompose{
		Name:    request.Name,
		Label:   request.Label,
		Image:   registryPath + "/" + image.Code + ":" + imageLastVersion,
		Type:    image.Labels.NzSitehostImageType,
		Ports:   ports,
		Volumes: volumes,
	})
	if err != nil {
		return dockerCompose, err
	}

	return compose, nil
}

func buildDockerCompose(request models.BuildDockerCompose) (dockerCompose string, err error) {
	// Create docker compose file
	compose := models.DockerCompose{
		Version: "2.1",
		Services: map[string]models.Service{
			request.Name: {
				ContainerName: request.Name,
				Environment: []string{
					"VIRTUAL_HOST=" + request.Label,
					"CERT_NAME=" + request.Label,
				},
				Expose: request.Ports,
				Image:  request.Image,
				Labels: []string{
					"nz.sitehost.container.website.vhosts=" + request.Label,
					"nz.sitehost.container.image_update=True",
					"nz.sitehost.container.label=" + request.Label,
					"nz.sitehost.container.type=" + request.Type,
					"nz.sitehost.container.monitored=True",
					"nz.sitehost.container.backup_disable=False",
				},
				Restart: "unless-stopped",
				Volumes: request.Volumes,
			},
		},
		Networks: models.Networks{
			Default: models.DefaultNetwork{
				External: models.ExternalNetwork{
					Name: "infra_default",
				},
			},
		},
	}

	composeYaml, err := yaml.Marshal(&compose)
	if err != nil {
		return dockerCompose, err
	}

	return string(composeYaml), nil
}
