package scheduler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	dockerconnector "github.com/jaku01/caddyservicediscovery/internal/docker"
	"log"
	"time"
)

func StartScheduleDiscovery(caddyAdminUrl string, scheduleInterval int) error {
	dockerConnector := dockerconnector.NewDockerConnector()
	caddyConnector := caddy.NewConnector(caddyAdminUrl)

	log.Println("Starting scheduler for service discovery")
	log.Printf("Using caddy admin url: %s", caddyAdminUrl)

	err := createCaddyConfigIfMissing(caddyConnector)
	if err != nil {
		return err
	}

	lastMapState := make([]byte, 0)

	for {
		serverMap, err := getServerMap(err, dockerConnector)
		if err != nil {
			return err
		}

		currentState, _ := json.Marshal(serverMap)

		if bytes.Equal(currentState, lastMapState) {
			log.Println("No changes detected, skip updating")
		} else {
			log.Println("Server map changes detected, updating caddy configuration")
			lastMapState = currentState

			err = caddyConnector.ReplaceServers(serverMap)
			if err != nil {
				return err
			}
		}

		time.Sleep(time.Duration(scheduleInterval) * time.Second)
	}
}

func getServerMap(err error, dockerConnector *dockerconnector.DockerConnector) (map[string]caddy.Server, error) {
	containers, err := dockerConnector.GetAllContainersWithActiveLabel()
	if err != nil {
		return nil, err
	}

	serverMap := make(map[string]caddy.Server)
	for _, container := range containers {
		reverseProxyServer := caddy.NewReverseProxyServer(container.Port, container.Upstream)
		serverMap[container.ContainerName] = reverseProxyServer
	}
	return serverMap, nil
}

func createCaddyConfigIfMissing(caddyConnector *caddy.Connector) error {
	config, err := caddyConnector.GetCaddyConfig()
	if err != nil {
		panic(err)
	}

	if config == nil {
		fmt.Println("No config found")
		err = caddyConnector.CreateCaddyConfig()
		if err != nil {
			panic(err)
		}
	}
	return err
}
