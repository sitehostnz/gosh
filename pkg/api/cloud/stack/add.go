package stack

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/api/cloud/stack/image"
	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
	"gopkg.in/yaml.v2"
)

// Add creates a new cloud stack.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := "cloud/stack/add.json"
	keys := []string{
		"client_id",
		"server",
		"name",
		"label",
		"enable_ssl",
		"docker_compose",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("enable_ssl", strconv.Itoa(request.EnableSSL))
	values.Add("docker_compose", request.DockerCompose)
	values.Add("label", request.Label)
	values.Add("name", request.Name)

	var vars string
	for _, envVar := range request.EnvironmentVariables {
		vars += fmt.Sprintf("  %s: %s\n", envVar.Name, envVar.Content)
	}

	if vars != "" {
		values.Add("environments["+request.Name+".env]", fmt.Sprintf("vars: \n%s", vars))
		keys = append(keys, "environments["+request.Name+".env]")
	}

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}

// AddWithImage creates a new cloud stack with an image (Web Image and Service Image are supported).
func (s *Client) AddWithImage(ctx context.Context, request AddRequestWithImage) (response AddResponse, err error) {
	uri := "cloud/stack/add.json"
	keys := []string{
		"client_id",
		"server",
		"name",
		"label",
		"enable_ssl",
		"docker_compose",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("name", request.Name)
	values.Add("label", request.Label)
	values.Add("enable_ssl", strconv.Itoa(request.EnableSSL))

	// Generate Docker Compose file
	dockerCompose, err := s.GenerateDockerCompose(ctx, models.GenerateDockerComposeRequest{
		Name:      request.Name,
		Label:     request.Label,
		ImageCode: request.ImageCode,
	})
	if err != nil {
		return response, err
	}

	values.Add("docker_compose", dockerCompose)

	var vars string
	for _, envVar := range request.EnvironmentVariables {
		vars += fmt.Sprintf("  %s: %s\n", envVar.Name, envVar.Content)
	}

	if vars != "" {
		values.Add("environments["+request.Name+".env]", fmt.Sprintf("vars: \n%s", vars))
		keys = append(keys, "environments["+request.Name+".env]")
	}

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}

// GenerateDockerCompose generates a docker compose file for a stack.
func (s *Client) GenerateDockerCompose(ctx context.Context, request models.GenerateDockerComposeRequest) (dockerCompose string, err error) {
	// Get the image
	i := image.New(s.client)
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
