package main

import (
	"fmt"
	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	dockerconnector "github.com/jaku01/caddyservicediscovery/internal/docker"
)

func main() {

	dockerConnector := dockerconnector.NewDockerConnector()
	caddyConnector := caddy.NewConnector("http://localhost:2019")

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

	containers, err := dockerConnector.GetAllContainersWithActiveLabel()
	if err != nil {
		panic(err)
	}

	serverMap := make(map[string]caddy.Server)
	for _, container := range containers {
		reverseProxyServer := caddy.NewReverseProxyServer(container.Port, container.Upstream)
		serverMap[container.ContainerName] = reverseProxyServer
	}

	err = caddyConnector.ReplaceServers(serverMap)
	if err != nil {
		panic(err)
	}
}
