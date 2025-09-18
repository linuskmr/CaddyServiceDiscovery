package dockerconnector

import (
	"context"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"log"
	"strconv"
)

const (
	activeLabel   = "caddy.service.discovery.active"
	portLabel     = "caddy.service.discovery.port"
	upstreamLabel = "caddy.service.discovery.upstream"
)

type DockerConnector struct {
	dockerClient *client.Client
	ctx          context.Context
}

type ContainerInfo struct {
	Port          int
	Upstream      string
	ContainerName string
}

func NewDockerConnector() *DockerConnector {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &DockerConnector{
		dockerClient: cli,
		ctx:          ctx,
	}
}

func (dc *DockerConnector) GetAllContainersWithActiveLabel() ([]ContainerInfo, error) {

	containers, err := dc.dockerClient.ContainerList(dc.ctx, containertypes.ListOptions{})
	if err != nil {
		return nil, err
	}

	var activeContainers []ContainerInfo
	for _, container := range containers {
		if container.Labels[activeLabel] == "true" {

			port, err := strconv.Atoi(container.Labels[portLabel])
			if err != nil {
				log.Println("Error converting port to int")
				continue
			}

			containerInfo := ContainerInfo{
				Port:          port,
				Upstream:      container.Labels[upstreamLabel],
				ContainerName: container.Names[0],
			}

			activeContainers = append(activeContainers, containerInfo)
		}
	}

	return activeContainers, nil
}
