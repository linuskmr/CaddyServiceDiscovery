package dockerconnector

import (
	"context"
	"log"
	"strconv"

	containertypes "github.com/docker/docker/api/types/container"
	eventtypes "github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
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

type EventType int

const (
	ContainerCreatedEvent = iota
	ContainerDestroyedEvent
)

type DockerEvent struct {
	ContainerInfo ContainerInfo
	EventType     EventType
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

func (dc *DockerConnector) GetEventChannel() <-chan DockerEvent {
	transformedEvents := make(chan DockerEvent)

	go func() {
		defer close(transformedEvents)
		ctxWithCancel, cancelCtx := context.WithCancel(dc.ctx)
		rawEvents, err := dc.dockerClient.Events(ctxWithCancel, eventtypes.ListOptions{})
		defer cancelCtx()

		for {
			select {
			case event, ok := <-rawEvents:
				if !ok {
					return
				}
				transformedEvent := transformDockerEvent(event)
				if transformedEvent == nil {
					continue
				}
				transformedEvents <- *transformedEvent
			case err := <-err:
				log.Println("Error listening to docker events:", err)
			case <-dc.ctx.Done():
				return
			}
		}
	}()

	return transformedEvents
}

func transformDockerEvent(rawEvent eventtypes.Message) *DockerEvent {
	if rawEvent.Type != eventtypes.ContainerEventType {
		return nil
	}
	if rawEvent.Actor.Attributes[activeLabel] != "true" {
		return nil
	}

	portStr := rawEvent.Actor.Attributes[portLabel]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Error converting docker container port '%s' to int\n", portStr)
		return nil
	}

	containerInfo := ContainerInfo{
		Port:          port,
		Upstream:      rawEvent.Actor.Attributes[upstreamLabel],
		ContainerName: rawEvent.Actor.Attributes["name"],
	}

	var eventType EventType
	switch rawEvent.Action {
	case eventtypes.ActionCreate:
		eventType = ContainerCreatedEvent
	case eventtypes.ActionDestroy:
		eventType = ContainerDestroyedEvent
	default:
		return nil
	}

	return &DockerEvent{
		ContainerInfo: containerInfo,
		EventType:     eventType,
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
